# 1.1.0

As of `logr@v1.0.0`, the `logr.Logger` is considered to be a defined `struct` instead of an `interface`. The implementation layer (now referred to as `logr.LogSink`) has been entirely restructured. Now, the `logerr` library will provide `logr.Logger` objects and ways to affect the underlying `Sink` operations.

- Instead of `log.V()`, `log.Info()`, `log.Error()` methods, use `log.DefaultLogger().V()`, `log.DefaultLogger().Info()`, `log.DefaultLogger().Error()`

- Instead of `log.SetOutput()`, `log.SetLogLevel()`, use either of the following methods:

Method 1:
```go
l := log.DefaultLogger()
s, err := log.GetSink(l)

if err != nil {
    // Some action
}

s.SetVerbosity(1)
s.SetOutput(ioutil.Discard)

l.Info("hello world")
```

Method 2:
```go
l := log.DefaultLogger()
// This method panics, but DefaultLogger will not
// panic because it uses Sink.
log.MustGetSink(l).SetVerbosity(1)
log.MustGetSink(l).SetOutput(ioutil.Discard)

l.Info("hello world")
```

- Instead of `log.UseLogger`, `log.GetLogger`, keep the logger instance until it is no longer needed

- Instead of `log.Init` or `log.InitWithOptions`, use `log.NewLogger` or `log.NewLoggerWithOption`
