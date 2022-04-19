# v2.0.0

## Removal of singleton behavior

Due to the structural changes of `logr@v1`, several features of this library have been modified. Most notable of these changes is the lack of a singleton logger. It is recommended that developers create a logger instance with `NewLogger` and either keep it as their own singleton or pass the instance throughout the application to use where needed. Methods like `Info` and `Error` are still callable with a `logr.Logger` instance.

ex:

```golang
import (
    "github.com/ViaQ/logerr/v2/log"
)

logger := log.NewLogger("example-logger")

logger.Info("Now logging info message")
logger.Error(errors.New("New error"), "Now logging new error")
```

## Logger Creation

`Init` and `InitWithOptions` have been removed. Please use `NewLogger(component string, opts ...Option)`.

## Removal of explicit `SetOutput` and `SetLevel`

As a byproduct of the changes in `logr@v1`, these methods have been moved into the internal logging implementation of `logr.Sink`: `Sink`. It is recommended that a logger instance is created with the intended writer and verbosity level. If the writer or verbosity needs to be changed, a new logger should be generated.
