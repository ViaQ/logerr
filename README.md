# logerr

logerr provides structured logging and kverrors packages that work together.

This package does not:

* Provide a complex API
* Log to any external loggers
* Promise any other features that aren't central to the
  [logr](https://github.com/go-logr/logr) interface

This package does:

* Create a structured, uniform, singleton, sane logging package to use across
  the life cycle of the application
* Provide structured errors with key/value pairs that can easily be extracted at
  log time
* Log structured (JSON) logs. If you are a human and you can't read your JSON
  logs then install [humanlog](https://github.com/aybabtme/humanlog)

## Logging TL;DR

As much as I hate the concept of TL;DR, the short of it is simple:

* logs are logs. There are no "levels" of logs. An error log is a log with an error
* log.Error = log.Info with an error key/value
* log.Info  = log.V(0).Info
* log.V(1).Info will not be printed unless you set the log level to 1 or above
* Use higher V for noisier, less useful logs
* Use lower V for more useful, less noisy logs
* Try to stick to a limited range of V because stretching the range only makes
  debugging that much more difficult. We recommend V 0-3 or 0-5 at most.

## Errors TL;DR

`kverrors` provides a package for creating key/value errors that create
key/value (aka Structured) errors. Errors should never contain sprintf strings,
instead you should place your key/value information into separate context that
can be easily queried later (with jq or an advanced log framework like
elasticsearch). The `log` package understands how to print kverrors as
structured.

## Log Verbosity

The log package of logerr does not have or support traditional "levels" such as
Trace, Debug, Info, Error, Fatal, Panic, etc. These levels are excessive and
unnecessary. Instead this library adapts the "verbosity" format. What does that
mean? If you have ever used tcpdump, memcached, ansible, ssh, or rsync you would
have used the -v, -vv, -v 3 idioms. These flags increase the logging verbosity
rather than changing levels. verbosity level 0 is always logged and should be
used for your standard logging. log.Info is identical to log.V(0).Info.

There is log.Error, which is _not_ a level, but rather a helper function to log
a go `error`. `log.Error(err, "message")` is identical to
`log.Info("message", "cause", err)`. It merely provides a uniform interface for
logging errors.

Let me explain with some examples:

```go
log.V(1).Info("I will only be printed if log level is set to 1+")
log.V(2).Info("I will only be printed if log level is set to 2+")
log.V(3).Info("I will only be printed if log level is set to 3+")

// These two logs are identical because Error is a helper that provides a
// uniform interface
log.V(3).Info("I will only be printed if log level is set to 3+", "cause", io.ErrClosedPipe)
log.V(3).Error(io.ErrClosedPipe, "I will only be printed if log level is set to 3+")
```

This allows output control by limiting verbose messages. By default, only V
level 0 is printed. If you set verbosity to 1 then V level 0-1 will be printed.
Instead of controlling arbitrary message "levels" think of controlling the
amount of output pushed by your application. More verbose messages produce more
logs, but provide more information. You probably don't care about higher
verbosity messages unless you are trying to find a bug. When you are running
locally 

### How to choose?

Traditional log levels give you the option of logging the level that you feel is
appropriate for the message. For example, log.Trace is typically referred to as
the most verbose output that contains lots of extra information such as
durations of method execution, printing variables, entering and exiting areas of
code, etc. However, this gets confusing when you compare it with log.Debug.
Where do you draw the line between trace and debug? There is no answer because
it's entirely arbitrary no matter how much you try to define it. With logerr the
answer of "how do I choose which verbosity level" is simple and also arbitrary.
You ask yourself "how much output is this going to produce and how useful is
that output? If you think the usefulness is high and the verbosity is low, then
you may just use log.Info which is _always_ logged. As the usefulness shrinks
and the verbosity raises, you should raise your the verbosity level of the log
line.

## Log Verbosity Examples

```go
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// this is going to be pretty noisy so lets set the verbosity level
	// pretty high
	log.V(3).Info("received HTTP request", 
		"url": r.URL,
		"method": r.Method,
		"remoteAddr": r.RemoteAddr,
	)

	userID, err := lookupUserID(r)
	if err != nil {
		// This is a critical path for us so we log.Error (verbosity 0)
		// which is always printed and can never be suppressed
		// log.Error and log.Info are identical to log.V(0).Error and
		// log.V(0).Info, respectively
		log.Error(err, "failed to lookup user ID")
		w.WriteHeader(500)
		return
	}

	cache, err := redis.LookupCachedResponse(userID, r.URL, r.Method)
	if err != nil {
		// an error means something didn't happen like Redis wasn't
		// reachable etc. Our caching layer isn't critical because we
		// only use it for quicker responses, however our backend can
		// handle 100% of traffic. This may cause a slight increase in
		// response times, but it's going to get noisy and may cause
		// Lets choose V 1 here
		log.V(1).Error(err, "response cache error")
	}

	if cache == nil {
		// a cache miss isn't really important or helpful in production
		// systems. We could use it when debugging to figure out how the
		// code is handling requests though. This may be somewhat noisy,
		// but it's not really helpful because a cache miss just means
		// we will go to the source of truth so it's not a big deal.
		log.V(2).Info("cache miss",
			"user_id": userID,
			"url": r.URL,
			"method": r.Method,
		)
	}
}

```

As you can see the verbosity is _arbitrarily_ chosen based on a quick guess of
noisiness and usefulness. Less useful, more noisy messages get a higher V
level. More useful, less noisy messages use the default V 0 output so that they
are always printed.  The philosophy is simple **spend less time thinking about
your log verbosity and more time coding**. Think of it like choosing scrum
points: compare it to previous log entries and ask yourself if it's greater
than, less than or the same in noise/usefulness.

## Usage Examples

```go
package main

import (
        "github.com/ViaQ/logerr/kverrors"
        "github.com/ViaQ/logerr/log"
)

func Logging() {
        err := TrySomething() 
        log.Error(err, "failed to do something", "application", "example")
        // {
        //  "v": 0, <- this is verbosity
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
        //  "v": 0, <- this is verbosity
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
        return kverrors.New("this was never meant to pass", "reason", "unimplemented")
}

func TrySomethingElse() error {
	    err := TrySomething()
        return kverrors.Wrap(err, "failed to execute method", "method", "TrySomething")
}
```
