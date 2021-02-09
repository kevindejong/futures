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
