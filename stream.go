package futures

import (
	"errors"
	"sync"
)

// ErrStreamClosed is returned by Stream.Next when the stream has been closed
var ErrStreamClosed = errors.New("stream closed")

// Stream abstracts a channel with idempotent error handling and cloning for broadcast
type Stream interface {
	Pending() <-chan struct{}
	Next() (interface{}, error)
	Clone() Stream
	Close()
}

// SendFunc is used to send items to a stream or close with an error
type SendFunc func(interface{}, error)

// NewStream creates a base stream and send function
func NewStream() (Stream, SendFunc) {
	tracker := &streamTracker{
		readers: make(map[*streamReader]struct{}),
		pending: make(chan struct{}),
	}
	reader := &streamReader{
		streamTracker: tracker,
	}
	tracker.readers[reader] = struct{}{}
	return reader, tracker.send
}

type streamTracker struct {
	sync.RWMutex
	readers map[*streamReader]struct{}
	err     error
	pending chan struct{}
}

type streamReader struct {
	*streamTracker
	items  []*interface{}
	closed bool
}

func (s *streamTracker) send(item interface{}, err error) {
	if item != nil && err != nil {
		panic("cannot send both item and error")
	}
	s.Lock()
	if err == nil {
		if s.pending == nil {
			panic("item sent to stream after error")
		}
		close(s.pending)
		s.pending = make(chan struct{})
		for reader := range s.readers {
			reader.items = append(reader.items, &item)
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

func (s *streamReader) Pending() <-chan struct{} {
	s.RLock()
	defer s.RUnlock()
	if s.closed || len(s.items) > 0 || s.pending == nil {
		pending := make(chan struct{})
		close(pending)
		return pending
	}
	return s.pending
}

func (s *streamReader) Next() (interface{}, error) {
	<-s.Pending()
	s.RLock()
	defer s.RUnlock()
	if s.closed {
		return nil, ErrStreamClosed
	}
	if len(s.items) > 0 {
		item := *s.items[0]
		s.items = s.items[1:]
		return item, nil
	}
	if s.pending == nil {
		return nil, s.err
	}
	panic("pending closed but no items or error set")
}

func (s *streamReader) Clone() Stream {
	s.Lock()
	clone := &streamReader{
		streamTracker: s.streamTracker,
		items:         s.items,
		closed:        s.closed,
	}
	s.readers[clone] = struct{}{}
	s.Unlock()
	return clone
}

func (s *streamReader) Close() {
	s.Lock()
	s.items = nil
	s.closed = true
	delete(s.readers, s)
	s.Unlock()
}
