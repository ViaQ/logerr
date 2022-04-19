package log

import (
	"io"

	"github.com/ViaQ/logerr/v2/internal/sink"
)

// Option is a configuration option
type Option func(*sink.Sink)

// WithOutput sets the output of the internal sink of the logger
func WithOutput(w io.Writer) Option {
	return func(s *sink.Sink) {
		s.SetOutput(w)
	}
}

// WithVerbosity sets the verbosity of the internal sink of the logger
func WithVerbosity(v int) Option {
	return func(s *sink.Sink) {
		s.SetVerbosity(v)
	}
}
