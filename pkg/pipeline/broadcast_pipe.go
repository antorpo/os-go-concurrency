package pipeline

import (
	"context"
	"fmt"
)

type (
	Broadcast struct {
		Name     string
		Label    string
		Comments string

		Streams []Flow
		Merger  FanInFn
	}
)

func (b *Broadcast) connect(
	ctx context.Context,
	in <-chan interface{},
	br breaker,
) (
	<-chan interface{},
	<-chan error,
) {
	out := make(chan interface{}, 1)
	errors := make(chan error, 1)
	tracer := traceMe(ctx, b)

	panicProof(
		func() {
			select {
			case data, ok := <-in:
				if ok {
					b.handleInput(ctx, out, errors, tracer, br, data)
				}
			case <-ctx.Done():
				tracer.canceled()
				return
			}
		},
		notifyPanicAsError(ctx, errors, br, tracer),
		closeOutput(out, errors),
	)

	return out, errors
}

func (b *Broadcast) handleInput(
	ctx context.Context,
	out chan interface{},
	errors chan error,
	tracer stopwatch,
	br breaker,
	data interface{},
) {
	defer tracer.done()

	tracer.start(ctx)

	joined, err := b.runFlows(ctx, data, br)
	if err != nil {
		cancel(tracer, err, errors, br)
		return
	}

	if isCancelled(len(b.Streams), len(joined)) {
		tracer.canceled()
		return
	}

	merged, err := b.Merger(ctx, joined)
	if err != nil {
		fail(tracer, err, errors, br)
		return
	}

	out <- merged
}

func (b *Broadcast) runFlows(
	ctx context.Context,
	data interface{},
	br breaker,
) (
	[]interface{},
	error,
) {
	var (
		pathOuts = make([]<-chan interface{}, len(b.Streams))
		pathErrs []<-chan error
		flowCtx  context.Context
	)

	for idx, flow := range b.Streams {
		pathIn := make(chan interface{}, 1)

		branchName := fmt.Sprintf("%s#%v", b.Name, idx)
		flowCtx = openBranch(CtxBranch(ctx, branchName), b, branchName)
		pathOut, ferr := connectFlow(flowCtx, pathIn, flow, br)

		pathOuts[idx] = pathOut
		pathErrs = append(pathErrs, ferr...)

		go feed(flowCtx, pathIn, data)
	}

	return mergeAll(
		func() <-chan interface{} { return fanIn(ctx, pathOuts...) },
		pathErrs,
	)
}

func (b *Broadcast) draw(s Skin) string {
	output := b.forks(s)
	output += b.comments()

	return output
}

func (b *Broadcast) forks(s Skin) string {
	output := "fork \n"

	for i, flow := range b.Streams {
		if i > 0 {
			output += "fork again\n"
		}

		for _, pipe := range flow {
			output += pipe.draw(s)
		}
	}

	output += "endfork \n"

	return output
}

func (b *Broadcast) comments() string {
	var output string

	output += "note right \n"
	output += fontMultiline(b.Label, "<font size=\"24\">")

	if b.Comments != "" {
		output += "\n"
		output += b.Comments
	}

	output += "\nend note \n"

	return output
}

func (*Broadcast) traced(n *tracerNode) string {
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

	for _, b := range n.branches.dump() {
		if counter > 0 {
			output += "fork again \n"
		}

		output += traceBranch(b)
		counter++
	}

	output += "endfork \n"
	output += notesOf(n)

	return output
}

func isCancelled(expected, actual int) bool {
	return expected != actual
}
