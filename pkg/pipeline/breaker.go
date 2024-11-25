package pipeline

import (
	"context"
	"sync"
)

type (
	breaker struct {
		context.CancelFunc
		done chan interface{}
		once *sync.Once
	}
)

func newBreaker(ctx context.Context) (context.Context, breaker) {
	cCtx, cancelFunc := context.WithCancel(ctx)
	return cCtx, breaker{
		CancelFunc: cancelFunc,
		done:       make(chan interface{}, 1),
		once:       &sync.Once{},
	}
}

func (b *breaker) cancel() {
	b.CancelFunc()
}

func (b *breaker) earlyExit(value interface{}) {
	b.once.Do(func() {
		b.done <- value
		close(b.done)
	})
}
