package futures

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamSendItemNoRecv(t *testing.T) {
	_, sendFunc := NewStream()
	sendFunc("test", nil)
}

func TestStreamSendErrNoRecv(t *testing.T) {
	_, sendFunc := NewStream()
	sendFunc(nil, errors.New("TestStreamSendErrNoRecv"))
}

func TestStreamSendBoth(t *testing.T) {
	_, sendFunc := NewStream()
	require.Panics(t, func() {
		sendFunc("TestStreamSendBoth", errors.New("TestStreamSendBoth"))
	})
}

func TestStreamSendMultipleErr(t *testing.T) {
	_, sendFunc := NewStream()
	sendFunc(nil, errors.New("TestStreamSendMultipleErr"))
	require.Panics(t, func() {
		sendFunc(nil, errors.New("TestStreamSendMultipleErr"))
	})
}

func TestStreamSendItemAfterError(t *testing.T) {
	_, sendFunc := NewStream()
	sendFunc(nil, errors.New("TestStreamSendItemAfterError"))
	require.Panics(t, func() {
		sendFunc("TestStreamSendItemAfterError", nil)
	})
}

func TestStreamPendingNoSend(t *testing.T) {
	stream, _ := NewStream()
	select {
	case <-stream.Pending():
		t.Fatal("stream unexpectedly pending")
	default:
	}
}

func TestStreamPendingSendItem(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamPendingSendItem"
	sendFunc(item, nil)
	select {
	case <-stream.Pending():
	default:
		t.Fatal("stream not pending")
	}
}

func TestStreamPendingSendErr(t *testing.T) {
	stream, sendFunc := NewStream()
	err := errors.New("TestStreamPendingSendErr")
	sendFunc(nil, err)
	select {
	case <-stream.Pending():
	default:
		t.Fatal("stream not pending")
	}
}

func TestStreamPendingSendItemBlocking(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamPendingSendItemBlocking"
	go func() {
		time.Sleep(10 * time.Millisecond)
		sendFunc(item, nil)
	}()
	select {
	case <-stream.Pending():
	case <-time.After(20 * time.Millisecond):
		t.Fatal("stream not pending")
	}
}

func TestStreamPendingSendErrBlocking(t *testing.T) {
	stream, sendFunc := NewStream()
	err := errors.New("TestStreamPendingSendErr")
	go func() {
		time.Sleep(10 * time.Millisecond)
		sendFunc(nil, err)
	}()
	select {
	case <-stream.Pending():
	case <-time.After(20 * time.Millisecond):
		t.Fatal("stream not pending")
	}
}

func TestStreamNextNoSend(t *testing.T) {
	stream, _ := NewStream()
	// this leaks but it's only a test
	done := make(chan struct{})
	go func() {
		stream.Next()
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("unexpected next return")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestStreamNextSendItem(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamNextSendItem"
	sendFunc(item, nil)
	recv, err := stream.Next()
	require.NoError(t, err)
	require.Equal(t, item, recv)
}

func TestStreamNextSendErr(t *testing.T) {
	stream, sendFunc := NewStream()
	expectedErr := errors.New("TestStreamNextSendErr")
	sendFunc(nil, expectedErr)
	recv, err := stream.Next()
	require.EqualError(t, err, expectedErr.Error())
	require.Equal(t, nil, recv)
}

func TestStreamNextSendItemBlocking(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamNextSendItemBlocking"
	go func() {
		time.Sleep(10 * time.Millisecond)
		sendFunc(item, nil)
	}()
	recv, err := stream.Next()
	require.NoError(t, err)
	require.Equal(t, item, recv)
}

func TestStreamNextSendErrBlocking(t *testing.T) {
	stream, sendFunc := NewStream()
	expectedErr := errors.New("TestStreamNextSendErr")
	go func() {
		time.Sleep(10 * time.Millisecond)
		sendFunc(nil, expectedErr)
	}()
	recv, err := stream.Next()
	require.EqualError(t, err, expectedErr.Error())
	require.Equal(t, nil, recv)
}

func TestStreamNextDoubleItem(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamNextDoubleItem"
	sendFunc(item, nil)
	recv, err := stream.Next()
	require.NoError(t, err)
	require.Equal(t, item, recv)
	done := make(chan struct{})
	go func() {
		stream.Next()
		close(done)
	}()
	select {
	case <-stream.Pending():
		t.Fatal("unexpected pending")
	case <-done:
		t.Fatal("unexpected done")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestStreamNextDoubleErr(t *testing.T) {
	stream, sendFunc := NewStream()
	err := errors.New("TestStreamNextDoubleErr")
	sendFunc(nil, err)
	recv, recvErr := stream.Next()
	require.Error(t, recvErr)
	require.Equal(t, nil, recv)
	recv2, recvErr2 := stream.Next()
	require.Error(t, recvErr2)
	require.Equal(t, nil, recv2)
}

func TestStreamNextItemErr(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamNextItemErr"
	err := errors.New("TestStreamNextItemErr")
	sendFunc(item, nil)
	sendFunc(nil, err)
	recv, recvErr := stream.Next()
	require.NoError(t, recvErr)
	require.Equal(t, item, recv)
	recv2, recvErr2 := stream.Next()
	require.Error(t, recvErr2)
	require.Equal(t, nil, recv2)
}

func TestStreamClonePendingNoSend(t *testing.T) {
	stream, _ := NewStream()
	clone := stream.Clone()
	select {
	case <-stream.Pending():
		t.Fatal("unexpected pending")
	case <-clone.Pending():
		t.Fatal("unexpected pending")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestStreamClonePendingItem(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	sendFunc("TestStreamClonePending", nil)
	select {
	case <-stream.Pending():
	case <-time.After(10 * time.Millisecond):
		t.Fatal("expected pending")
	}
	select {
	case <-clone.Pending():
	case <-time.After(10 * time.Millisecond):
		t.Fatal("expected pending")
	}
}

func TestStreamClonePendingErr(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	sendFunc(nil, errors.New("TestStreamClonePendingItem"))
	select {
	case <-stream.Pending():
	case <-time.After(10 * time.Millisecond):
		t.Fatal("expected pending")
	}
	select {
	case <-clone.Pending():
	case <-time.After(10 * time.Millisecond):
		t.Fatal("expected pending")
	}
}

func TestStreamCloneNextItem(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	item := "TestStreamCloneNextItem"
	sendFunc(item, nil)
	streamItem, streamErr := stream.Next()
	cloneItem, cloneErr := clone.Next()
	require.NoError(t, streamErr)
	require.NoError(t, cloneErr)
	require.Equal(t, item, streamItem)
	require.Equal(t, item, cloneItem)
}

func TestStreamCloneNextErr(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	err := errors.New("TestStreamCloneNextErr")
	sendFunc(nil, err)
	streamItem, streamErr := stream.Next()
	cloneItem, cloneErr := clone.Next()
	require.Error(t, streamErr)
	require.Error(t, cloneErr)
	require.Equal(t, nil, streamItem)
	require.Equal(t, nil, cloneItem)
}

func TestStreamCloneBeforeNext(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamCloneBeforeNext"
	sendFunc(item, nil)
	clone := stream.Clone()
	streamItem, streamErr := stream.Next()
	cloneItem, cloneErr := clone.Next()
	require.NoError(t, streamErr)
	require.NoError(t, cloneErr)
	require.Equal(t, item, streamItem)
	require.Equal(t, item, cloneItem)
}

func TestStreamCloneAfterNext(t *testing.T) {
	stream, sendFunc := NewStream()
	item1 := "TestStreamCloneAfterNext1"
	item2 := "TestStreamCloneAfterNext2"
	sendFunc(item1, nil)
	streamItem, streamErr := stream.Next()
	require.NoError(t, streamErr)
	require.Equal(t, item1, streamItem)
	sendFunc(item2, nil)
	clone := stream.Clone()
	streamItem2, streamErr2 := stream.Next()
	require.NoError(t, streamErr2)
	require.Equal(t, item2, streamItem2)
	cloneItem, cloneErr := clone.Next()
	require.NoError(t, cloneErr)
	require.Equal(t, item2, cloneItem)
}

func TestStreamCloneAfterNextErr(t *testing.T) {
	stream, sendFunc := NewStream()
	err := errors.New("TestStreamCloneAfterNextErr")
	sendFunc(nil, err)
	streamItem, streamErr := stream.Next()
	require.Error(t, streamErr)
	require.Equal(t, nil, streamItem)
	clone := stream.Clone()
	streamItem2, streamErr2 := stream.Next()
	require.Error(t, streamErr2)
	require.Equal(t, nil, streamItem2)
	cloneItem, cloneErr := clone.Next()
	require.Error(t, cloneErr)
	require.Equal(t, nil, cloneItem)
}

func TestStreamClose(t *testing.T) {
	stream, _ := NewStream()
	stream.Close()
	select {
	case <-stream.Pending():
	default:
		t.Fatal("expected stream pending")
	}
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
}

func TestStreamCloseAferItem(t *testing.T) {
	stream, sendFunc := NewStream()
	item := "TestStreamCloseAferItem"
	sendFunc(item, nil)
	stream.Close()
	select {
	case <-stream.Pending():
	default:
		t.Fatal("expected stream pending")
	}
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
}

func TestStreamCloseAferErr(t *testing.T) {
	stream, sendFunc := NewStream()
	err := errors.New("TestStreamCloseAferErr")
	sendFunc(nil, err)
	stream.Close()
	select {
	case <-stream.Pending():
	default:
		t.Fatal("expected stream pending")
	}
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
}

func TestStreamCloseAferCloneItem(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	stream.Close()
	expectedItem := "TestStreamCloseAferCloneItem"
	sendFunc(expectedItem, nil)
	select {
	case <-stream.Pending():
	default:
		t.Fatal("expected stream pending")
	}
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
	cloneItem, err := clone.Next()
	require.NoError(t, err)
	require.Equal(t, expectedItem, cloneItem)
}

func TestStreamCloseAferCloneErr(t *testing.T) {
	stream, sendFunc := NewStream()
	clone := stream.Clone()
	stream.Close()
	expectedErr := errors.New("TestStreamCloseAferCloneErr")
	sendFunc(nil, expectedErr)
	select {
	case <-stream.Pending():
	default:
		t.Fatal("expected stream pending")
	}
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
	cloneItem, err := clone.Next()
	require.EqualError(t, err, expectedErr.Error())
	require.Equal(t, nil, cloneItem)
}

func TestStreamCloneAfterClose(t *testing.T) {
	stream, sendFunc := NewStream()
	stream.Close()
	clone := stream.Clone()
	expected := "TestStreamCloseClone"
	sendFunc(expected, nil)
	nextItem, err := stream.Next()
	require.EqualError(t, err, ErrStreamClosed.Error())
	require.Equal(t, nil, nextItem)
	cloneItem, err := clone.Next()
	require.NoError(t, err)
	require.Equal(t, expected, cloneItem)
}
