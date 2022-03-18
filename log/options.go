package log

import (
	"io"
)

// Option is a configuration option
type Option func(*Sink)

// WithOutput sets the output to w
func WithOutput(w io.Writer) Option {
	return func(s *Sink) {
		s.SetOutput(w)
	}
}

// WithLogLevel sets the output log level and controls which verbosity logs are printed
func WithLogLevel(v int) Option {
	return func(s *Sink) {
		s.SetVerbosity(v)
	}
}
