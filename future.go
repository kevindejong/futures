package futures

import "sync"

// Future provides idempotent access to the result of a concurrent process
type Future interface {
	Done() <-chan struct{}
	Result() (interface{}, error)
}

// CompleteFunc completes a future with a result; must be called at most once
type CompleteFunc func(interface{}, error)

// NewFuture returns a new Future along with the associated CompleteFunc
func NewFuture() (Future, CompleteFunc) {
	fut := &future{
		done: make(chan struct{}),
	}
	return fut, fut.complete
}

type future struct {
	mu   sync.Mutex
	done chan (struct{})
	val  interface{}
	err  error
}

// Done signals when the future has been completed
func (f *future) Done() <-chan struct{} {
	return f.done
}

// Result waits for future completion and returns the result
func (f *future) Result() (interface{}, error) {
	<-f.done
	return f.val, f.err
}

func (f *future) complete(val interface{}, err error) {
	f.mu.Lock()
	select {
	case <-f.done:
		panic("future completed multiple times")
	default:
		f.val = val
		f.err = err
		close(f.done)
	}
	f.mu.Unlock()
}
