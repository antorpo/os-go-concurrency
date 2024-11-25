package pipeline

import (
	"context"
)

type (
	StageFn func(context.Context, interface{}) (interface{}, error)

	SimplePipe struct {
		Resolver StageFn
		Comments string
	}
)

func Stage(r StageFn) *SimplePipe {
	return &SimplePipe{Resolver: r}
}

func (sp *SimplePipe) connect(
	ctx context.Context,
	in <-chan interface{},
	b breaker,
) (
	<-chan interface{},
	<-chan error,
) {
	out := make(chan interface{}, cap(in))
	errors := make(chan error)
	tracer := traceMe(ctx, sp)

	panicProof(
		func() {
			processData := func(data interface{}) {
				defer tracer.done()
				tracer.start(ctx)

				result, err := sp.Resolver(ctx, data)
				if err != nil {
					fail(tracer, err, errors, b)
					return
				}

				out <- result
			}

			select {
			case data, ok := <-in:
				if ok {
					processData(data)
				} else {
					tracer.canceled()
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

func (sp *SimplePipe) draw(_ Skin) string {
	out := drawStage(sp.Resolver)

	if sp.Comments != "" {
		out += "note right \n"
		out += sp.Comments
		out += "\n"
		out += "end note \n"
	}

	return out
}

func (sp *SimplePipe) traced(n *tracerNode) string {
	var out string

	switch {
	case n.error != nil:
		out = drawFailedStage(sp.Resolver, n.error)

	case n.cancelled:
		out = drawCanceledStage(sp.Resolver)

	default:
		out = drawStage(sp.Resolver)
	}

	out += notesOf(n)

	return out
}
