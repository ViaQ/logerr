package log

import (
	"os"

	"github.com/ViaQ/logerr/v2/internal/sink"
	"github.com/go-logr/logr"
)

// NewLogger creates a logger with the provided opts and key value pairs
func NewLogger(component string, opts ...Option) logr.Logger {
	sink := sink.NewLogSink(component, os.Stdout, 0, sink.JSONEncoder{}, nil)

	for _, opt := range opts {
		opt(sink)
	}

	return logr.New(sink)
}
