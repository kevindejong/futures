# Futures

### Future

A future provides idempotent access to the result of a concurrent process.

```
future, completeFunc := futures.NewFuture()
go func() {
    time.Sleep(1 * time.Second)
    completeFunc("future completed!", nil)
}()
<-future.Done()
result, err := future.Result()
fmt.Print(result, " ", err) // future completed! <nil>
```

### Stream

A stream is an unbounded channel with idempotent error handling and cloning for broadcast.

```
stream, sendFunc := futures.NewStream()
clone := stream.Clone()

sendFunc("stream item!", nil)
streamItem, streamErr := stream.Next()
fmt.Print(streamItem, " ", streamErr) // stream item! <nil>
cloneItem, cloneErr := clone.Next()
fmt.Print(cloneItem, " ", cloneErr) // stream item! <nil>

sendFunc(nil, errors.New("stream closed!"))
streamItem2, streamErr2 := stream.Next()
fmt.Print(streamItem2, " ", streamErr2) // <nil> stream closed!
cloneItem2, cloneErr2 := clone.Next()
fmt.Print(cloneItem2, " ", cloneErr2) // <nil> stream closed!

stream.Close()
clone.Close()
```


### Abort

Abort extends `context.Context` with support for custom error messages.

```
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
abortCtx, abortFunc := futures.WithAbort(ctx)
go func() {
    time.Sleep(1 * time.Second)
    abortFunc(errors.New("goroutine aborted!"))
}()
<-abortCtx.Done()
fmt.Print(abortCtx.Err()) // goroutine aborted!
```

```
abortCtx, abortFunc := futures.NewAbort()
ctx, cancel := context.WithCancel(abortCtx)
defer cancel()
go func() {
    time.Sleep(1 * time.Second)
    abortFunc(errors.New("goroutine aborted!"))
}()
<-ctx.Done()
fmt.Print(ctx.Err()) // goroutine aborted!
```
