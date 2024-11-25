package pipeline

import (
	"context"
)

type (
	Flow []Pipe

	SourceFn func(context.Context, interface{}) (interface{}, error)
	SinkFn   func(context.Context, interface{}) (interface{}, error)

	Pipe interface {
		connect(context.Context, <-chan interface{}, breaker) (<-chan interface{}, <-chan error)
		Drawable
		Traceable
	}

	Pipeline struct {
		Name           string
		Description    string
		SourceComments string
		SinkComments   string

		Source SourceFn
		Flow   Flow
		Sink   SinkFn

		BlueprintSkin Skin
		TraceSkin     Skin
	}

	FanOutFn func(context.Context, interface{}) ([]interface{}, error)
	FanInFn  func(context.Context, []interface{}) (interface{}, error)

	ContextBranch func(ctx context.Context, name string) context.Context
	TrafficTagger func(context.Context, interface{}) string
	BranchTagger  func(context.Context, interface{}) string

	Drawable interface {
		draw(Skin) string
	}

	Traceable interface {
		traced(*tracerNode) string
	}

	Jotter interface {
		Note(string)
		LazyNote(func() string)
	}

	flowCounter struct {
		total  uint64
		tagged map[string]uint64
	}

	flowsPercent struct {
		total  float32
		tagged map[string]float32
	}
)

const oneHundred = 100

var (
	CtxBranch                ContextBranch = func(ctx context.Context, _ string) context.Context { return ctx }
	FunctionsNamePrefixPrune               = 3
	EncryptedMode                          = true
)

func Run(ctx context.Context, input interface{}, bp *Pipeline, traced bool) (interface{}, error) {
	var pCtx = ctx
	if traced {
		pCtx = newTracer(ctx, bp)
	}

	pCtx, breaker := newBreaker(pCtx)

	ch, err := source(pCtx, input, bp.Source)
	if err != nil {
		return nil, err
	}

	pCh, eCh := connectFlow(pCtx, ch, bp.Flow, breaker)

	return sink(pCtx, pCh, eCh, bp.Sink, breaker)
}

func RunWithTracer(ctx context.Context, input interface{}, bp *Pipeline) (interface{}, string, error) {
	tCtx := newTracer(ctx, bp)
	out, err := Run(tCtx, input, bp, false /* <- tCtx is traced */)

	return out, TracedLink(tCtx), err
}

func source(
	ctx context.Context,
	req interface{},
	resolver SourceFn,
) (
	chan interface{},
	error,
) {
	token, err := resolver(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan interface{}, 1)

	go func() {
		defer close(ch)

		select {
		case ch <- token:
		case <-ctx.Done():
			return
		}
	}()

	return ch, nil
}

func connectFlow(
	ctx context.Context,
	source <-chan interface{},
	flow Flow,
	b breaker,
) (
	<-chan interface{},
	[]<-chan error,
) {
	var errcList = make([]<-chan error, len(flow))

	in := source

	for idx, stage := range flow {
		next, errc := stage.connect(ctx, in, b)
		in = next

		errcList[idx] = errc
	}

	return in, errcList
}

func sink(
	ctx context.Context,
	in <-chan interface{},
	errc []<-chan error,
	resolver SinkFn,
	b breaker,
) (
	interface{},
	error,
) {
	err := WaitForPipeline(errc...)
	if err != nil {
		return nil, err
	}

	var out interface{}
	var ready = false

	for !ready {
		select {
		case out, ready = <-b.done:
			b.cancel()
		case out, ready = <-in:
		case <-ctx.Done():
			_, deadline := ctx.Deadline()
			if deadline {
				return nil, ctx.Err()
			}
		}
	}

	return resolver(ctx, out)
}

func fail(tracer stopwatch, err error, errors chan error, b breaker) {
	tracer.fail(err)
	sendError(err, errors, b)
}

func cancel(tracer stopwatch, err error, errors chan error, b breaker) {
	tracer.canceled()
	sendError(err, errors, b)
}

func sendError(err error, errors chan error, b breaker) {
	errors <- err

	b.cancel()
}

func done(ctx context.Context, out chan interface{}, in <-chan interface{}, cancel stopwatch) {
	select {
	case v, ok := <-in:
		if ok {
			out <- v
			return
		}

	case <-ctx.Done():
		cancel.canceled()
		return
	}
}
