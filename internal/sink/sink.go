package sink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ViaQ/logerr/v2/internal/kv"
	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/go-logr/logr"
)

// TimestampFunc returns a string formatted version of the current time.
// This should probably only be used with tests or if you want to change
// the default time formatting of the output logs.
var TimestampFunc = func() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// Sink writes logs to a specified output
type Sink struct {
	mtx       sync.RWMutex
	verbosity Verbosity
	output    io.Writer
	context   map[string]interface{}
	encoder   Encoder
	name      string
}

// NewLogSink creates a new logsink
func NewLogSink(name string, w io.Writer, v Verbosity, e Encoder, keysAndValues ...interface{}) *Sink {
	return &Sink{
		name:      name,
		verbosity: v,
		output:    w,
		context:   kv.ToMap(keysAndValues...),
		encoder:   e,
	}
}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (s *Sink) Init(info logr.RuntimeInfo) {}

// Enabled determines if a logger should record a log. If the log's verbosity
// is higher or equal to that the logger's level, the log is recorded. Otherwise,
// it is skipped.
func (s *Sink) Enabled(level int) bool {
	// Mutex lock is used here. Don't need to use it for Info & Error
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.verbosity >= Verbosity(level)
}

// Info logs a non-error message with the given key/value pairs as context. Info
// will check to see if the log is enabled for the logger's level before recording.
func (s *Sink) Info(level int, msg string, keysAndValues ...interface{}) {
	if !s.Enabled(level) {
		return
	}
	s.log(msg, combine(s.context, keysAndValues...))
}

// Error logs an error, with the given message and key/value pairs as context. Unlike
// Info, it bypasses the Enabled check. Logs will always be recorded from this method.
func (s *Sink) Error(err error, msg string, keysAndValues ...interface{}) {
	// Use 0 as the level since it is the smallest level a logger can have
	// and thus will always pass the Enabled check.
	if err == nil {
		s.Info(0, msg, keysAndValues)
		return
	}

	switch err.(type) {
	case *kverrors.KVError:
		// nothing to be done
	default:
		err = kverrors.New(err.Error())
	}

	s.Info(0, msg, append(keysAndValues, ErrorKey, err)...)
}

// WithValues clones the logsink and appends keysAndValues.
func (s *Sink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	ss := NewLogSink(s.name, s.output, s.verbosity, s.encoder)
	ss.context = combine(s.context, keysAndValues...)

	return ss
}

// WithName clones the logsink and overwrites the name.
func (s *Sink) WithName(name string) logr.LogSink {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	newName := name
	if s.name != "" {
		newName = fmt.Sprintf("%s_%s", s.name, name)
	}

	return NewLogSink(newName, s.output, s.verbosity, s.encoder, s.context)
}

// SetOutput sets the writer that JSON is written to
func (s *Sink) SetOutput(w io.Writer) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.output = w
}

// SetVerbosity sets the log level allowed by the logsink
func (s *Sink) SetVerbosity(v int) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.verbosity = Verbosity(v)
}

// GetVerbosity returns the log level
func (s *Sink) GetVerbosity() int {
	return int(s.verbosity)
}

// log will log the message. It DOES NOT check Enabled() first so that should
// be checked by it's callers
func (s *Sink) log(msg string, context map[string]interface{}) {
	_, file, line, _ := runtime.Caller(3)
	file = sourcePath(file)

	m := Line{
		Timestamp: TimestampFunc(),
		FileLine:  fmt.Sprintf("%s:%s", file, strconv.Itoa(line)),
		Verbosity: s.verbosity.String(),
		Component: s.name,
		Message:   msg,
		Context:   context,
	}

	err := s.encoder.Encode(s.output, m)
	if err != nil {
		// expand first so we can quote later
		orig := fmt.Sprintf("%#v", m)
		_, _ = fmt.Fprintf(s.output, `{"message","failed to encode message", "encoder":"%T","log":%q,"cause":%q}`, s.encoder, orig, err)
	}
}

// combine creates a new map combining context and keysAndValues.
func combine(context map[string]interface{}, keysAndValues ...interface{}) map[string]interface{} {
	nc := make(map[string]interface{}, len(context)+len(keysAndValues)/2)

	for k, v := range context {
		nc[k] = v
	}

	kve := kv.ToMap(keysAndValues...)
	for k, v := range kve {
		nc[k] = v
	}

	return nc
}

func sourcePath(file string) string {
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil && !strings.HasPrefix(rel, "../") {
			return rel
		}
	}
	return filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
}
