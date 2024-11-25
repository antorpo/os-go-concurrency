package pipeline

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type (
	PartitionData struct {
		Name string
		Data interface{}
	}

	PartitionsFn    func(context.Context, interface{}) ([]PartitionData, error)
	PartitionTagger func(context.Context, PartitionData) string

	PartitionPipe struct {
		Name        string
		Partitioner PartitionsFn
		Merger      FanInFn
		Paths       map[string]Flow

		Tagger        PartitionTagger
		TrafficTagger TrafficTagger

		mtx      sync.Mutex
		counters map[string]*flowCounter
	}
)

func (pp *PartitionPipe) connect(
	ctx context.Context,
	in <-chan interface{},
	b breaker,
) (
	<-chan interface{},
	<-chan error,
) {
	out := make(chan interface{}, 1)
	errors := make(chan error, 1)
	tracer := traceMe(ctx, pp)

	panicProof(
		func() {
			select {
			case data, ok := <-in:
				if ok {
					pp.handleInput(ctx, out, errors, tracer, b, data)
				}
			case <-ctx.Done():
				tracer.canceled()
				return
			}
		},
		notifyPanicAsError(ctx, errors, b, tracer),
		closeOutput(out, errors),
	)

	return out, errors
}

func (pp *PartitionPipe) handleInput(
	ctx context.Context,
	out chan interface{},
	errors chan error,
	tracer stopwatch,
	b breaker,
	data interface{},
) {
	defer tracer.done()

	tracer.start(ctx)

	paths, err := pp.Partitioner(ctx, data)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	pp.count(ctx, paths)

	c, joined, err := pp.runFlows(ctx, paths, b)
	if err != nil {
		cancel(tracer, err, errors, b)
		return
	}

	if isCancelled(c, len(joined)) {
		tracer.canceled()
		return
	}

	merged, err := pp.Merger(ctx, joined)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	out <- merged
}

func (pp *PartitionPipe) runFlows(
	ctx context.Context,
	paths []PartitionData,
	b breaker,
) (
	int,
	[]interface{},
	error,
) {
	var (
		pathOuts []<-chan interface{}
		pathErrs []<-chan error
		flowCtx  context.Context
		counter  int
	)

	for idx, dataPath := range paths {
		if stream, ok := pp.Paths[dataPath.Name]; ok {
			counter++

			pathIn := make(chan interface{}, 1)

			flowCtx = CtxBranch(ctx, fmt.Sprintf("%s#%v", dataPath.Name, idx))
			flowCtx = openBranch(flowCtx, pp, pp.Tagger(flowCtx, dataPath))
			pathOut, ferr := connectFlow(flowCtx, pathIn, stream, b)

			pathOuts = append(pathOuts, pathOut)
			pathErrs = append(pathErrs, ferr...)

			go feed(flowCtx, pathIn, dataPath.Data)
		}
	}

	all, err := mergeAll(
		func() <-chan interface{} { return fanIn(ctx, pathOuts...) },
		pathErrs,
	)

	return counter, all, err
}

func (pp *PartitionPipe) draw(s Skin) string {
	var output string

	var sortedKeys []string
	for k := range pp.Paths {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	volume := pp.flowVolume()

	output += "split \n"

	for counter, k := range sortedKeys {
		flow := pp.Paths[k]

		flowVolume := volume[k]

		if counter > 0 {
			output += "split again \n"
		}

		output += fmt.Sprintf("note \n == ❖ %s \nend note\n", k)
		output += drawFlowVolume("left", flowVolume)

		for _, pipe := range flow {
			output += pipe.draw(s)
		}
	}

	output += "endsplit \n"

	return output
}

func (pp *PartitionPipe) traced(n *tracerNode) string {
	var (
		output  string
		counter int
	)

	if n.error != nil {
		output += "split \n"
		output += fmt.Sprintf(": ☠ %+v; \n", n.error)
		output += "endsplit \n"
		output += notesOf(n)

		return output
	}

	output += "split \n"

	for name, b := range n.branches.dump() {
		if counter > 0 {
			output += "split again \n"
		}

		output += fmt.Sprintf(":%s }\n", fontMultiline(name, "<font color=\"$traceTagColor\">"))
		output += traceBranch(b)

		counter++
	}

	output += "endsplit \n"
	output += notesOf(n)

	return output
}

func (pp *PartitionPipe) count(ctx context.Context, paths []PartitionData) {
	defer pp.mtx.Unlock()
	pp.mtx.Lock()

	if pp.counters == nil {
		pp.counters = make(map[string]*flowCounter)
	}

	for _, path := range paths {
		var (
			counter *flowCounter
			ok      bool
		)

		if counter, ok = pp.counters[path.Name]; !ok {
			counter = &flowCounter{tagged: make(map[string]uint64)}
		}

		if pp.TrafficTagger != nil {
			tag := pp.TrafficTagger(ctx, path.Data)

			counter.tagged[tag]++
		}

		counter.total++
		pp.counters[path.Name] = counter
	}
}

func (pp *PartitionPipe) flowVolume() map[string]flowsPercent {
	defer pp.mtx.Unlock()
	pp.mtx.Lock()

	out := make(map[string]flowsPercent)

	if pp.counters == nil {
		return out
	}

	var global uint64
	for _, v := range pp.counters {
		global += v.total
	}

	for k, v := range pp.counters {
		branchCounter := v.total

		percentMap := flowsPercent{
			total:  (float32(branchCounter) / float32(global)) * oneHundred,
			tagged: make(map[string]float32),
		}

		for kk, vv := range v.tagged {
			percentMap.tagged[kk] = (float32(vv) / float32(branchCounter)) * oneHundred
		}

		out[k] = percentMap
	}

	return out
}
