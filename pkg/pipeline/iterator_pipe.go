package pipeline

import (
	"context"
	"fmt"
)

const (
	IteratorParentValue ctxKey = "parent.value"
	IteratorParentCtx   ctxKey = "parent.ctx"
)

type (
	ctxKey string

	JoinerFn func(context.Context, interface{}, []interface{}) (interface{}, error)

	Iterator struct {
		Name     string
		MaxP     *int
		Splitter FanOutFn
		Stream   Flow
		Joiner   JoinerFn
		Tagger   BranchTagger
	}
)

func IteratorParent(ctx context.Context) (interface{}, context.Context) {
	pCtx, _ := ctx.Value(IteratorParentCtx).(context.Context)
	return ctx.Value(IteratorParentValue), pCtx
}

func (i *Iterator) connect(
	ctx context.Context,
	in <-chan interface{},
	b breaker,
) (
	<-chan interface{},
	<-chan error,
) {
	out := make(chan interface{}, 1)
	errors := make(chan error, 1)
	tracer := traceMe(ctx, i)

	panicProof(
		func() {
			select {
			case data, ok := <-in:
				if ok {
					i.handleInput(ctx, out, errors, tracer, b, data)
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

func (i *Iterator) handleInput(
	ctx context.Context,
	out chan interface{},
	errors chan error,
	tracer stopwatch,
	b breaker,
	data interface{},
) {
	defer tracer.done()

	tracer.start(ctx)

	paths, err := i.Splitter(ctx, data)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	var (
		chunks = i.chunks(paths)
		joined []interface{}
	)

	for _, chunk := range chunks {
		chunkResults, err := i.runFlows(ctx, data, chunk, b)
		if err != nil {
			cancel(tracer, err, errors, b)
			return
		}

		if isCancelled(len(chunkResults), len(chunk)) {
			tracer.canceled()
			return
		}

		joined = append(joined, chunkResults...)
	}

	merged, err := i.Joiner(ctx, data, joined)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	out <- merged
}

func (i *Iterator) runFlows(
	ctx context.Context,
	data interface{},
	paths []interface{},
	b breaker,
) (
	[]interface{},
	error,
) {
	var (
		pathOuts = make([]<-chan interface{}, len(paths))
		pathErrs []<-chan error
		flowCtx  context.Context
	)

	for idx, pathData := range paths {
		pathIn := make(chan interface{}, 1)

		txnName := fmt.Sprintf("%s#%v", i.Name, idx)
		flowCtx = CtxBranch(ctx, txnName)
		flowCtx = context.WithValue(flowCtx, IteratorParentValue, data)
		flowCtx = context.WithValue(flowCtx, IteratorParentCtx, ctx)
		tagger := i.Tagger(flowCtx, pathData)
		flowCtx = openBranch(flowCtx, i, tagger)

		pathOut, ferr := connectFlow(flowCtx, pathIn, i.Stream, b)

		pathOuts[idx] = pathOut
		pathErrs = append(pathErrs, ferr...)

		go feed(flowCtx, pathIn, pathData)
	}

	return mergeAll(
		func() <-chan interface{} { return fanIn(ctx, pathOuts...) },
		pathErrs,
	)
}

func (i *Iterator) draw(s Skin) string {
	output := "fork \n"

	for _, pipe := range i.Stream {
		output += pipe.draw(s)
	}

	output += "endfork \n"
	output += fmt.Sprintf("note right\n<font size=\"24\">%s</font>\nend note\n", i.Name)

	return output
}

func (i *Iterator) traced(n *tracerNode) string {
	var (
		output  string
		counter int
	)

	if n.error != nil {
		output += "fork \n"
		output += fmt.Sprintf(": ☠ %+v; \n", n.error)
		output += "endfork \n"
		output += notesOf(n)

		return output
	}

	output += "fork \n"

	for name, b := range n.branches.dump() {
		if counter > 0 {
			output += "fork again \n"
		}

		output += fmt.Sprintf(":%s }\n", fontMultiline(name, "<font color=\"$traceTagColor\">"))
		output += traceBranch(b)

		counter++
	}

	output += "endfork \n"
	output += notesOf(n)

	return output
}

func (i *Iterator) chunks(paths []interface{}) [][]interface{} {
	if i.MaxP == nil {
		return [][]interface{}{paths}
	}

	var (
		chunkSize = *i.MaxP
		out       [][]interface{}
	)

	if chunkSize == 0 {
		return [][]interface{}{paths}
	}

	for {
		if len(paths) == 0 {
			break
		}

		if len(paths) < chunkSize {
			chunkSize = len(paths)
		}

		out = append(out, paths[0:chunkSize])
		paths = paths[chunkSize:]
	}

	return out
}
