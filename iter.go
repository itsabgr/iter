package iter

import (
	"context"
	"time"
)

type Iter[T any] struct {
	c      chan T
	ctx    context.Context
	cancel context.CancelFunc
}

func New[T any](ctx context.Context, cap int) *Iter[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &Iter[T]{
		c:      make(chan T, cap),
		cancel: cancel,
		ctx:    ctx,
	}
}
func (i *Iter[T]) Ctx() context.Context {
	return i.ctx
}
func (i *Iter[T]) Close() error {
	if i.ctx.Err() != nil {
		return i.ctx.Err()
	}
	i.cancel()
	return nil
}
func (i *Iter[T]) IsClosed() bool {
	return i.ctx.Err() != nil
}
func (i *Iter[T]) GetCtx(ctx context.Context) (t T, _ error) {
	select {
	case <-ctx.Done():
		return t, ctx.Err()
	case <-i.ctx.Done():
		return t, i.ctx.Err()
	case val := <-i.c:
		return val, nil
	}
}

func (i *Iter[T]) GetDeadline(deadline time.Time) (t T, _ error) {
	ctx, cancel := context.WithDeadline(i.ctx, deadline)
	defer cancel()
	return i.GetCtx(ctx)
}
func (i *Iter[T]) GetTimeout(timeout time.Duration) (t T, _ error) {
	ctx, cancel := context.WithTimeout(i.ctx, timeout)
	defer cancel()
	return i.GetCtx(ctx)
}
func (i *Iter[T]) Add(t T) error {
	if i.IsClosed() {
		return context.Canceled
	}
	select {
	case i.c <- t:
		return nil
	case <-i.ctx.Done():
		return i.ctx.Err()
	}
}

func (i *Iter[T]) Get() (t T, _ error) {
	select {
	case <-i.ctx.Done():
		return t, i.ctx.Err()
	case val := <-i.c:
		return val, nil
	}
}
