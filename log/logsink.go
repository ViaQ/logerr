package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ViaQ/logerr/internal/kv"
	"github.com/ViaQ/logerr/kverrors"
	"github.com/go-logr/logr"
)

// Keys used to log specific builtin fields
const (
	TimeStampKey = "_ts"
	FileLineKey  = "_file:line"
	LevelKey     = "_level"
	ComponentKey = "_component"
	MessageKey   = "_message"
	ErrorKey     = "_error"
)

// Line orders log line fields
type Line struct {
	Timestamp string
	FileLine  string
	Verbosity string
	Component string
	Message   string
	Context   map[string]interface{}
}

// LineJSON add json tags to Line struct (production logs)
type LineJSON struct {
	Timestamp string                 `json:"_ts"`
	FileLine  string                 `json:"-"`
	Verbosity string                 `json:"_level"`
	Component string                 `json:"_component"`
	Message   string                 `json:"_message"`
	Context   map[string]interface{} `json:"-"`
}

// LineJSONDev add json tags to Line struct (developer logs, enable using environment variable LOG_DEV)
type LineJSONDev struct {
	Timestamp string                 `json:"_ts"`
	FileLine  string                 `json:"_file:line"`
	Verbosity string                 `json:"_level"`
	Component string                 `json:"_component"`
	Message   string                 `json:"_message"`
	Context   map[string]interface{} `json:"-"`
}

// MarshalJSON implements custom marshaling for log line: (1) flattening context (2) support for developer mode
func (l Line) MarshalJSON() ([]byte, error) {
	lineTemp := LineJSON(l)

	lineValue, err := json.Marshal(lineTemp)
	if err != nil {
		return nil, err
	}
	verbosity, errConvert := strconv.Atoi(l.Verbosity)
	if verbosity > 1 && errConvert == nil {
		lineTempDev := LineJSONDev(l)
		lineValue, err = json.Marshal(lineTempDev)
		if err != nil {
			return nil, err
		}
	}
	lineValue = lineValue[1 : len(lineValue)-1]

	contextValue, err := json.Marshal(lineTemp.Context)
	if err != nil {
		return nil, err
	}
	contextValue = contextValue[1 : len(contextValue)-1]

	sep := ""
	if len(contextValue) > 0 {
		sep = ","
	}
	return []byte(fmt.Sprintf("{%s%s%s}", lineValue, sep, contextValue)), nil
}

// Verbosity is a level of verbosity to log between 0 and math.MaxInt32
// However it is recommended to keep the numbers between 0 and 3
type Verbosity int

func (v Verbosity) String() string {
	return strconv.Itoa(int(v))
}

// MarshalJSON marshals JSON
func (v Verbosity) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

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
func (ls *Sink) Init(info logr.RuntimeInfo) {}

// Enabled tests whether this logsink is enabled.  For example, commandline
// flags might be used to set the logging verbosity and disable some info
// logs.
func (ls *Sink) Enabled(level int) bool {
	ls.mtx.RLock()
	defer ls.mtx.RUnlock()
	return ls.verbosity <= Verbosity(level)
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
func (ls *Sink) Info(level int, msg string, keysAndValues ...interface{}) {
	if !ls.Enabled(level) {
		return
	}
	ls.log(msg, combine(ls.context, keysAndValues...))
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
func (ls *Sink) Error(err error, msg string, keysAndValues ...interface{}) {
	if err == nil {
		ls.Info(int(ls.verbosity), msg, keysAndValues)
		return
	}

	switch err.(type) {
	case *kverrors.KVError:
		// nothing to be done
	default:
		err = kverrors.New(err.Error())
	}

	ls.Info(int(ls.verbosity), msg, append(keysAndValues, ErrorKey, err)...)
}

// WithValues clones the logsink and appends keysAndValues
func (ls *Sink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return ls.withValues(keysAndValues...)
}

// WithName adds a new element to the logsink's name.
// Successive calls with WithName continue to append
// suffixes to the logsink's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (ls *Sink) WithName(name string) logr.LogSink {
	newName := name
	if ls.name != "" {
		newName = fmt.Sprintf("%s_%s", ls.name, name)
	}

	return NewLogSink(newName, ls.output, ls.verbosity, ls.encoder, ls.context)
}

// SetOutput sets the writer that JSON is written to
func (ls *Sink) SetOutput(w io.Writer) {
	ls.mtx.Lock()
	defer ls.mtx.Unlock()
	ls.output = w
}

// SetVerbosity sets the log level allowed by the logsink
func (ls *Sink) SetVerbosity(v int) {
	ls.mtx.Lock()
	defer ls.mtx.Unlock()
	ls.verbosity = Verbosity(v)
}

// withValues clones the logger and appends keysAndValues
// but returns a struct instead of the logr.Logger interface
func (ls *Sink) withValues(keysAndValues ...interface{}) *Sink {
	ll := NewLogSink(ls.name, ls.output, ls.verbosity, ls.encoder)
	ll.context = combine(ls.context, keysAndValues...)
	return ll
}

// log will log the message. It DOES NOT check Enabled() first so that should
// be checked by it's callers
func (ls *Sink) log(msg string, context map[string]interface{}) {
	_, file, line, _ := runtime.Caller(3)
	file = sourcePath(file)
	m := Line{
		Timestamp: TimestampFunc(),
		FileLine:  fmt.Sprintf("%s:%s", file, strconv.Itoa(line)),
		Verbosity: ls.verbosity.String(),
		Component: ls.name,
		Message:   msg,
		Context:   context,
	}

	err := ls.encoder.Encode(ls.output, m)
	if err != nil {
		// expand first so we can quote later
		orig := fmt.Sprintf("%#v", m)
		_, _ = fmt.Fprintf(ls.output, `{"message","failed to encode message", "encoder":"%T","log":%q,"cause":%q}`, ls.encoder, orig, err)
	}
}

// combine creates a new map combining context and keysAndValues.
func combine(context map[string]interface{}, keysAndValues ...interface{}) map[string]interface{} {
	nc := make(map[string]interface{}, len(context)+len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string) // It should be a string.
			if !ok {                             // But this is not the place to panic
				key = fmt.Sprintf("%s", keysAndValues[i]) // So use this expensive conversion instead.
			}
			nc[key] = keysAndValues[i+1]
		}
	}
	for k, v := range context {
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
