package futures

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureNoComplete(t *testing.T) {
	future, _ := NewFuture()
	select {
	case <-future.Done():
		t.Fatal("future unexpectedly completed")
	default:
	}
}

func TestFutureCompleteResult(t *testing.T) {
	future, complete := NewFuture()
	result := "result"
	complete(result, nil)
	select {
	case <-future.Done():
		res, err := future.Result()
		require.NoError(t, err)
		require.Equal(t, result, res)
	default:
		t.Fatal("future not completed")
	}
}

func TestFutureCompleteErr(t *testing.T) {
	future, complete := NewFuture()
	expectedErr := errors.New("TestFutureCompleteErr")
	complete(nil, expectedErr)
	select {
	case <-future.Done():
		res, err := future.Result()
		require.Equal(t, nil, res)
		require.EqualError(t, err, expectedErr.Error())
	default:
		t.Fatal("future not completed")
	}
}

func TestFutureMultipleComplete(t *testing.T) {
	_, complete := NewFuture()
	complete(nil, nil)
	require.Panics(t, func() {
		complete(nil, nil)
	})
}

func TestFutureResultBlocking(t *testing.T) {
	future, complete := NewFuture()
	expectedErr := errors.New("TestFutureResultBlocking")
	go func() {
		time.Sleep(10 * time.Millisecond)
		complete(nil, expectedErr)
	}()
	val, err := future.Result()
	require.EqualError(t, err, expectedErr.Error())
	require.Equal(t, nil, val)
}
