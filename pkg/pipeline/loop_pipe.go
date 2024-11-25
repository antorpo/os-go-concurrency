package pipeline

import (
	"context"
	"fmt"
)

type (
	Loop struct {
		Name     string
		Splitter FanOutFn
		Stream   Flow
		Joiner   JoinerFn
		Tagger   BranchTagger
	}
)

func (l *Loop) connect(ctx context.Context, in <-chan interface{}, b breaker) (<-chan interface{}, <-chan error) {
	out := make(chan interface{}, 1)
	errors := make(chan error, 1)
	tracer := traceMe(ctx, l)

	panicProof(
		func() {
			select {
			case data, ok := <-in:
				if ok {
					l.handleInput(ctx, out, errors, tracer, b, data)
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

func (l *Loop) handleInput(
	ctx context.Context,
	out chan interface{},
	errors chan error,
	tracer stopwatch,
	b breaker,
	data interface{},
) {
	defer tracer.done()

	tracer.start(ctx)

	values, err := l.Splitter(ctx, data)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	joined, err := l.runFlow(ctx, data, values, b)
	if err != nil {
		cancel(tracer, err, errors, b)
		return
	}

	if isCancelled(len(values), len(joined)) {
		tracer.canceled()
		return
	}

	merged, err := l.Joiner(ctx, data, joined)
	if err != nil {
		fail(tracer, err, errors, b)
		return
	}

	out <- merged
}

func (l *Loop) runFlow(
	ctx context.Context,
	data interface{},
	values []interface{},
	b breaker,
) (
	[]interface{},
	error,
) {
	var (
		pathOuts = make([]<-chan interface{}, len(values))
		pathErrs []<-chan error
		flowCtx  context.Context
	)

	for idx, pathData := range values {
		pathIn := make(chan interface{}, 1)

		txnName := fmt.Sprintf("%s#%v", l.Name, idx)
		flowCtx = CtxBranch(ctx, txnName)
		flowCtx = context.WithValue(flowCtx, IteratorParentValue, data)
		flowCtx = context.WithValue(flowCtx, IteratorParentCtx, ctx)
		tagger := l.Tagger(flowCtx, pathData)
		flowCtx = openBranch(flowCtx, l, tagger)

		pathOut, ferr := connectFlow(flowCtx, pathIn, l.Stream, b)

		pathOuts[idx] = pathOut
		pathErrs = append(pathErrs, ferr...)

		feed(flowCtx, pathIn, pathData)

		for e := range mergeErrors(pathErrs...) {
			return nil, e
		}
	}

	return mergeAll(
		func() <-chan interface{} { return fanIn(ctx, pathOuts...) },
		pathErrs,
	)
}

func (l *Loop) draw(s Skin) string {
	output := "repeat\n"

	for _, pipe := range l.Stream {
		output += pipe.draw(s)
	}

	output += "repeat while \n"

	return output
}

func (l *Loop) traced(n *tracerNode) string {
	var (
		output string
	)

	if n.error != nil {
		output += "repeat \n"
		output += fmt.Sprintf(": ☠ %+v; \n", n.error)
		output += "repeat while \n"
		output += notesOf(n)

		return output
	}

	for name, b := range n.branches.dump() {
		output += fmt.Sprintf(":%s }\n", fontMultiline(name, "<font color=\"$traceTagColor\">"))
		output += traceBranch(b)
	}

	output += notesOf(n)

	return output
}
