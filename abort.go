package futures

import (
	"context"
	"sync"
	"time"
)

// AbortContext extends the context.Context with support for custom errors
type AbortContext interface {
	context.Context
}

// AbortFunc is a parallel of context.CancelFunc which allows a custom error
type AbortFunc func(error)

// NewAbort creates a new base context and abort function
func NewAbort() (AbortContext, AbortFunc) {
	future, completeFunc := NewFuture()
	ctx := &abortContext{
		future:       future,
		completeFunc: completeFunc,
	}
	return ctx, ctx.abort
}

// WithAbort is a parallel of context.WithCancel which wraps a context  and returns an abort fuction
func WithAbort(ctx context.Context) (AbortContext, AbortFunc) {
	future, completeFunc := NewFuture()
	abortCtx := &abortContext{
		inner:        ctx,
		future:       future,
		completeFunc: completeFunc,
	}
	go func() {
		select {
		case <-ctx.Done():
			abortCtx.abort(ctx.Err())
		case <-abortCtx.Done():
		}
	}()
	return abortCtx, abortCtx.abort
}

type abortContext struct {
	inner        context.Context
	once         sync.Once
	future       Future
	completeFunc CompleteFunc
}

func (a *abortContext) Done() <-chan struct{} {
	return a.future.Done()
}

func (a *abortContext) Err() error {
	select {
	case <-a.future.Done():
		_, err := a.future.Result()
		return err
	default:
		return nil
	}
}

func (a *abortContext) Deadline() (time.Time, bool) {
	if a.inner == nil {
		return time.Time{}, false
	}
	return a.inner.Deadline()
}

func (a *abortContext) Value(key interface{}) interface{} {
	if a.inner == nil {
		return nil
	}
	return a.inner.Value(key)
}

func (a *abortContext) abort(err error) {
	a.once.Do(func() {
		// to avoid racing in cases where the complete call triggers on inner
		// context cancel, always prefer the inner context before the abort
		if a.inner != nil {
			select {
			case <-a.inner.Done():
				a.completeFunc(nil, a.inner.Err())
				return
			default:
			}
		}
		a.completeFunc(nil, err)
	})
}
