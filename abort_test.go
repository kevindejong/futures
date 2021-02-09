package futures

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAbortNoComplete(t *testing.T) {
	ctx, _ := NewAbort()
	require.Equal(t, nil, ctx.Err())
	select {
	case <-ctx.Done():
		t.Fatal("abort unexpectedly completed")
	default:
	}
}

func TestAbortCompleteErr(t *testing.T) {
	ctx, complete := NewAbort()
	expectedErr := errors.New("TestAbortCompleteErr")
	complete(expectedErr)
	select {
	case <-ctx.Done():
		require.EqualError(t, ctx.Err(), expectedErr.Error())
	default:
		t.Fatal("abort unexpectedly completed")
	}
}

func TestAbortCompleteNoErr(t *testing.T) {
	ctx, complete := NewAbort()
	complete(nil)
	select {
	case <-ctx.Done():
		require.Equal(t, nil, ctx.Err())
	default:
		t.Fatal("abort unexpectedly completed")
	}
}

func TestAbortMultipleComplete(t *testing.T) {
	ctx, complete := NewAbort()
	expectedErr := errors.New("TestAbortCompleteErr")
	complete(expectedErr)
	require.NotPanics(t, func() {
		complete(nil)
	})
	select {
	case <-ctx.Done():
		require.EqualError(t, ctx.Err(), expectedErr.Error())
	default:
		t.Fatal("abort not completed as expected")
	}
}

func TestAbortContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	abortCtx, _ := WithAbort(ctx)
	cancel()
	select {
	case <-abortCtx.Done():
		require.EqualError(t, abortCtx.Err(), context.Canceled.Error())
	case <-time.After(10 * time.Millisecond): // context cancelation is not instantaneous
		t.Fatal("abort not completed as expected")
	}
}

func TestAbortContextCancelThenComplete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	abortCtx, abortFunc := WithAbort(ctx)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		abortFunc(errors.New("unexpected error"))
		close(done)
	}()
	cancel()
	<-done
	select {
	case <-abortCtx.Done():
		require.EqualError(t, abortCtx.Err(), context.Canceled.Error())
	case <-time.After(10 * time.Millisecond): // context cancelation is not instantaneous
		t.Fatal("abort not completed as expected")
	}
}

func TestAbortNoDeadline(t *testing.T) {
	abortCtx, _ := NewAbort()
	ctxDeadline, ok := abortCtx.Deadline()
	if ok {
		log.Fatal("expected invalid deadline")
	}
	require.Equal(t, time.Time{}, ctxDeadline)
}

func TestAbortContextDeadline(t *testing.T) {
	deadline := time.Time{}.Add(1 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	abortCtx, _ := WithAbort(ctx)
	ctxDeadline, ok := abortCtx.Deadline()
	if !ok {
		log.Fatal("expected valid deadline")
	}
	require.Equal(t, deadline, ctxDeadline)
}

func TestAbortNoValues(t *testing.T) {
	abortCtx, _ := NewAbort()
	val := abortCtx.Value("abc")
	require.Equal(t, nil, val)
}

type abortTestKey string

func TestAbortContextValues(t *testing.T) {
	key := abortTestKey("key")
	value := "value"
	ctx := context.WithValue(context.Background(), key, value)
	abortCtx, _ := WithAbort(ctx)
	val := abortCtx.Value(key)
	require.Equal(t, value, val)
}

func TestAbortInner(t *testing.T) {
	abortCtx, abortFunc := NewAbort()
	defer abortFunc(nil)
	ctx, cancel := context.WithCancel(abortCtx)
	defer cancel()
	expectedErr := errors.New("TestAbortInner")
	abortFunc(expectedErr)
	<-ctx.Done()
	require.EqualError(t, ctx.Err(), expectedErr.Error())
}
