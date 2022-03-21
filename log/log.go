package log

import (
	"fmt"
	"os"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/go-logr/logr"
)

var (
	// ErrUnknownSinkType is returned when trying to perform a *Logger only function
	// that is incompatible with logr.Logger interface
	ErrUnknownSinkType = kverrors.New("unknown log sink type")

	defaultOutput = os.Stdout
)

// DefaultLogger creates a logger without any key value pairs
func DefaultLogger() logr.Logger {
	return logr.New(NewLogSink("", defaultOutput, 0, JSONEncoder{}))
}

// NewLogger creates a logger with the provided key value pairs
func NewLogger(component string, keyValuePairs ...interface{}) logr.Logger {
	return NewLoggerWithOptions(component, nil, keyValuePairs...)
}

// NewLoggerWithOptions creates a logger with the provided opts and key value pairs
func NewLoggerWithOptions(component string, opts []Option, keyValuePairs ...interface{}) logr.Logger {
	s := NewLogSink(component, defaultOutput, 0, JSONEncoder{}, keyValuePairs...)

	for _, opt := range opts {
		opt(s)
	}

	return logr.New(s)
}

// GetSink return the LogSink converted as a Sink object. It returns an
// error if it cannot convert it.
func GetSink(l logr.Logger) (*Sink, error) {
	s, ok := l.GetSink().(*Sink)

	if !ok {
		return nil, kverrors.Add(ErrUnknownSinkType,
			"sink_type", fmt.Sprintf("%T", s),
			"expected_type", fmt.Sprintf("%T", &Sink{}),
		)
	}
	return s, nil
}

// MustGetSink retrieves the Sink and panics if it gets an error from GetSink
func MustGetSink(l logr.Logger) *Sink {
	s, err := GetSink(l)
	if err != nil {
		panic(err)
	}
	return s
}
