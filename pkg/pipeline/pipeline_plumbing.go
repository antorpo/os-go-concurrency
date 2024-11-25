package pipeline

import (
	"context"
	"fmt"
	"sync"
)

func mergeAll(data func() <-chan interface{}, errcList []<-chan error) ([]interface{}, error) {
	err := WaitForPipeline(errcList...)
	if err != nil {
		return []interface{}{}, err
	}

	var all []interface{}

	in := data()
	for p := range in {
		all = append(all, p)
	}

	return all, nil
}

func mergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error, len(cs))
	output := func(c <-chan error) {
		defer wg.Done()

		for n := range c {
			out <- n
		}
	}

	wg.Add(len(cs))

	for _, c := range cs {
		go output(c)
	}

	go func() {
		defer close(out)
		wg.Wait()
	}()

	return out
}

func fanIn(ctx context.Context, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup

	multiplexedStream := make(chan interface{})

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()

		for i := range c {
			select {
			case multiplexedStream <- i:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(channels))

	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

func WaitForPipeline(errs ...<-chan error) error {
	for err := range mergeErrors(errs...) {
		return err
	}

	return nil
}

func feed(ctx context.Context, ch chan interface{}, input interface{}) {
	defer close(ch)

	select {
	case ch <- input:
	case <-ctx.Done():
		return
	}
}

func notifyPanicAsError(
	ctx context.Context,
	errors chan error,
	b breaker,
	tracer stopwatch,
) func(panic interface{}) {
	return func(panic interface{}) {
		err := fmt.Errorf("panic recovered: %+v", panic)
		fail(tracer, err, errors, b)
	}
}

func closeOutput(out chan interface{}, errors chan error) func() {
	return func() {
		close(out)
		close(errors)
	}
}
