# logerr

logerr provides structured logging and errors packages that work together.

This package does not:

* Create a new logging framework or library
* Provide a complex API
* Log to any external loggers
* Promise any other features that aren't central to the [logr](https://github.com/go-logr/logr) interface

This package does:

* Create a structured, uniform, singleton, sane logging package to use across the lifecycle of the application
* Provide structured errors with key/value pairs that can easily be extracted at log time

## Examples

```go
package main

import (
        "github.com/ViaQ/logerr/errors"
        "github.com/ViaQ/logerr/log"
)

func Logging() {
        err := TrySomething() 
        log.Error(err, "failed to do something", "application", "example")
        // {
        //  "level": "error",
        //  "ts": "<timestamp>",
        //  "msg": "failed to do something",
        //  "application": "example"
        //  "cause": {
        //     "msg": "this was never meant to pass",
        //     "reason": "unimplemented",
        //  }
        // }

        // Nested Errors
        err = TrySomethingElse() 
        log.Error(err, "failed to do something", "application", "example")
        // {
        //  "level": "error",
        //  "ts": "<timestamp>",
        //  "msg": "failed to do something",
        //  "application": "example"
        //  "cause": {
        //     "msg": "failed to execute method",
        //     "method": "TrySomething",
        //     "cause": {
        //       "msg": "this was never meant to pass",
        //       "reason": "unimplemented",
        //     }
        //  }
        // }
}


func TrySomething() error {
        return errors.New("this was never meant to pass", "reason", "unimplemented")
}

func TrySomethingElse() error {
	    err := TrySomething()
        return errors.Wrap(err, "failed to execute method", "method", "TrySomething")
}
```