package pipeline

import (
	"context"
	"fmt"
	"sync"
)

type (
	IsTrueFn func(context.Context, interface{}) (bool, error)

	IfPipe struct {
		Name          string
		Decider       IsTrueFn
		TrueFlow      Flow
		FalseFlow     Flow
		TrafficTagger TrafficTagger

		mtx      sync.Mutex
		counters map[string]*flowCounter
	}
)

func (s *IfPipe) connect(
	ctx context.Context,
	in <-chan interface{},
	b breaker,
) (
	<-chan interface{},
	<-chan error,
) {
	out := make(chan interface{}, cap(in))
	errors := make(chan error, 1)
	tracer := traceMe(ctx, s)

	panicProof(
		func() {
			select {
			case data, ok := <-in:
				if ok {
					s.handleInput(ctx, out, errors, tracer, b, data)
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

func (s *IfPipe) handleInput(
	ctx context.Context,
	out chan interface{},
	errors chan error,
	tracer stopwatch,
	b breaker,
	data interface{},
) {
	defer tracer.done()
	tracer.start(ctx)

	pipeIn := make(chan interface{}, 1)

	isTrue, err := s.Decider(ctx, data)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	flowCtx := openBranch(CtxBranch(ctx, fmt.Sprintf("%s#%v", s.Name, isTrue)), s, fmt.Sprint(isTrue))
	flow := s.selectedFlow(ctx, data, isTrue)
	pOut, pErrs := connectFlow(flowCtx, pipeIn, flow, b)

	go feed(flowCtx, pipeIn, data)

	for e := range mergeErrors(pErrs...) {
		tracer.canceled()
		errors <- e
	}

	done(ctx, out, pOut, tracer)
}

func (s *IfPipe) selectedFlow(ctx context.Context, data interface{}, isTrue bool) Flow {
	var flow Flow
	if isTrue {
		flow = s.TrueFlow
	} else {
		flow = s.FalseFlow
	}

	s.count(ctx, data, fmt.Sprint(isTrue))

	return flow
}

func (s *IfPipe) draw(skin Skin) string {
	var output string

	volume := s.flowVolume()
	output += fmt.Sprintf("if (%s?) then (yes)\n", s.Name)
	output += drawFlowVolume("left", volume["true"])

	for _, pipe := range s.TrueFlow {
		output += pipe.draw(skin)
	}

	output += "else (no)\n"
	output += drawFlowVolume("right", volume["false"])

	for _, pipe := range s.FalseFlow {
		output += pipe.draw(skin)
	}

	output += "endif \n"

	return output
}

func (s *IfPipe) traced(n *tracerNode) string {
	var output string

	if n.error != nil {
		output += fmt.Sprintf(": ♢  ☠ %+v     /\n", n.error)
		output += notesOf(n)

		return output
	}

	for name, b := range n.branches.dump() {
		output += fmt.Sprintf(": ♢  %s? : %s     /\n", s.Name, name)
		output += traceBranch(b)
	}

	output += notesOf(n)

	return output
}

func (s *IfPipe) count(ctx context.Context, data interface{}, selectedPath string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	if s.counters == nil {
		s.counters = make(map[string]*flowCounter)
	}

	var (
		counter *flowCounter
		ok      bool
	)

	if counter, ok = s.counters[selectedPath]; !ok {
		counter = &flowCounter{tagged: make(map[string]uint64)}
	}

	if s.TrafficTagger != nil {
		tag := s.TrafficTagger(ctx, data)
		counter.tagged[tag]++
	}

	counter.total++
	s.counters[selectedPath] = counter
}

func (s *IfPipe) flowVolume() map[string]flowsPercent {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	out := make(map[string]flowsPercent)

	if s.counters == nil {
		return out
	}

	var global uint64
	for _, v := range s.counters {
		global += v.total
	}

	for k, v := range s.counters {
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
