# logerr

logerr is a library which uses [logr](https://github.com/go-logr/logr) to provide structured logging for applications.

For migration information between versions, please check the [guide](./MIGRATION.md)

## Quick `logr` Overview

The `logr.Logger` differs from the traditional "log levels" (ex: trace, debug, info, error, fatal, etc) in favor of verbosity. The higher the verbosity level, the noisier the logs.

Ex:

```golang
// Traditional logger
logger.Trace("Entering method X. Useful for trace through, but not much else.")
logger.Fatal("Method X has crashed because of reason Y. This is useful info.")

// logr
logger.V(5).Info("Entering method X. Useful for trace through, but not much else.")
logger.V(0).Info("Method X has crashed because of reason Y. This is useful info.")
```

A `logr.Logger` is created with an initial level of 0. Calling the `V` method will generate a new logger with a higher level.

```golang
logger.V(1).Info("This logger's level has been raised by 1.")
logger.V(5).Info("This logger's level has been raised by 5.")
```

_Note: The `V` method is always additive. So logger.V(1).V(1) has a logger level of 2 instead of 1._

Every log has a verbosity, which is controlled by the internal logging mechanism `logr.Sink`.

Logs are recorded in two scenarios:

1. The log is an error. Errors - regardless of verbosity - are always logged (they are recorded at verbosity 0).
2. The log's verbosity is greater than or equal to the logger's level.

Ex:

```golang
logger.Info("This is a brand new logger. The log's verbosity is at 0 and the logger's level is 0. This log is recorded.")
logger.V(1).Info("Now this logger's level is 1, but the verbosity is still 0. This log is not recorded.")
logger.V(1).Error(errors.New("an error", "But this is an error. It's always recorded."))
```

_Note: As mentioned above, with no concept of an "error" log, there is no error level. The `error` method simply provides a way to print an error log in a structured manner._

## log

`log` provides methods for creating `logr.Logger` instances with a `logr.Sink` (`Sink`) that understands how to create JSON logs with key/value information. In general, the `Sink` does not need to be changed.

Logger Types:

```golang
logger := log.NewLogger("default-logger")
logger.Info("Log verbosity is 0. Logs are written to stdout.")

buffer := bytes.NewBuffer(nil)
newLogger := log.NewLogger("customized-logger", log.WithOutput(buffer), log.WithVerbosity(1))
newLogger.Info("Log verbosity is 1. Logs are written to byte buffer.")
```

As mentions, the `Sink` will transform messages into JSON logs. Key/value information that is included in the message is also included in the log.

Ex:

```golang
logger.Info("Log with key values", "key1", "value1", "key2", "value2")
```

## kverrors

`kverrors` provides a package for creating key/value errors that create key/value (aka structured) errors. Errors should never contain sprintf strings, instead place key/value information into separate context that can be easily queried later (with jq or an advanced log framework like elasticsearch).
