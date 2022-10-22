package futures

import (
	"errors"
	"sync"
)

// ErrStreamClosed is returned by Stream.Next when the stream has been closed
var ErrStreamClosed = errors.New("stream closed")

// Stream is an unbounded channel with idempotent error handling and cloning for broadcast
type Stream[T any] interface {
	Pending() <-chan T
	Next() (T, error)
	Clone() Stream[T]
	Close()
}

// SendFunc is used to send items to a stream or close with an error
type SendFunc[T any] func(T, error)

// NewStream creates a base stream and send function
func NewStream[T any]() (Stream[T], SendFunc[T]) {
	tracker := &streamTracker[T]{
		readers: make(map[*streamReader[T]]struct{}),
		pending: make(chan T),
	}
	reader := &streamReader[T]{
		streamTracker: tracker,
	}
	tracker.readers[reader] = struct{}{}
	return reader, tracker.send
}

type streamTracker[T any] struct {
	sync.RWMutex
	readers map[*streamReader[T]]struct{}
	err     error
	pending chan T
}

type streamReader[T any] struct {
	*streamTracker[T]
	items  []T
	closed bool
}

func (s *streamTracker[T]) send(item T, err error) {
	s.Lock()
	if err == nil {
		if s.pending == nil {
			panic("item sent to stream after error")
		}
		close(s.pending)
		s.pending = make(chan T)
		for reader := range s.readers {
			reader.items = append(reader.items, item)
		}
	} else {
		if s.pending == nil {
			panic("multiple errors sent to stream")
		}
		s.err = err
		close(s.pending)
		s.pending = nil
	}
	s.Unlock()
}

func (s *streamReader[T]) Pending() <-chan T {
	s.RLock()
	defer s.RUnlock()
	if s.closed || len(s.items) > 0 || s.pending == nil {
		pending := make(chan T)
		close(pending)
		return pending
	}
	return s.pending
}

func (s *streamReader[T]) Next() (T, error) {
	<-s.Pending()
	s.RLock()
	defer s.RUnlock()
	var emptyT T
	if s.closed {
		return emptyT, ErrStreamClosed
	}
	if len(s.items) > 0 {
		item := s.items[0]
		s.items = s.items[1:]
		return item, nil
	}
	if s.pending == nil {
		return emptyT, s.err
	}
	panic("pending closed but no items or error set")
}

func (s *streamReader[T]) Clone() Stream[T] {
	s.Lock()
	clone := &streamReader[T]{
		streamTracker: s.streamTracker,
		items:         s.items,
	}
	s.readers[clone] = struct{}{}
	s.Unlock()
	return clone
}

func (s *streamReader[T]) Close() {
	s.Lock()
	s.items = nil
	s.closed = true
	delete(s.readers, s)
	s.Unlock()
}
