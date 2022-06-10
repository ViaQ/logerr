// Package static provides a simple static logger implementation for easy migration of existing pre-v2 code.
//
// The package is usually imported aliased as "log" to be compatible with the v1 approach.
//
// The static logger instance needs to be initialized during startup of the application using SetLogger.
package static

import (
	"sync"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/go-logr/logr"
)

var (
	lock   sync.RWMutex
	logger = log.NewLogger("uninitialized")
)

// SetLogger sets the static logger instance.
func SetLogger(replacement logr.Logger) {
	lock.Lock()
	defer lock.Unlock()

	logger = replacement
}

// Logger returns the static logger instance.
func Logger() logr.Logger {
	lock.RLock()
	defer lock.RUnlock()
	return logger
}

// WithName returns a logger with name added to the component.
// This uses the static logger instance.
func WithName(name string) logr.Logger {
	lock.RLock()
	defer lock.RUnlock()

	return logger.WithName(name)
}

// V returns a logger with the verbosity set to level.
// This uses the static logger instance.
func V(level int) logr.Logger {
	lock.RLock()
	defer lock.RUnlock()

	return logger.V(level)
}

// Info logs an informational message with optional key-value pairs.
// This uses the static logger instance.
func Info(msg string, keysAndValues ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()

	logger.Info(msg, keysAndValues...)
}

// Error logs an error message with optional key-value pairs.
// This uses the static logger instance.
func Error(err error, msg string, keysAndValues ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()

	logger.Error(err, msg, keysAndValues...)
}
