package log_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

type Fields map[string]interface{}

// assertLoggedFields checks that all of fields exists in the entry.
func assertLoggedFields(t *testing.T, entry *observedEntry, fields Fields) {
	var f []string
	for k := range fields {
		f = append(f, k)
	}

	ctx := entry.Fields(f...)

	for k, v := range fields {
		value, ok := ctx[k]
		require.True(t, ok, "expected key '%s:%v' to exist in logged entry %+v", k, v, entry)
		actual, e := json.Marshal(value)
		require.NoError(t, e)
		expected, e := json.Marshal(v)
		require.NoError(t, e)
		require.JSONEq(t, string(expected), string(actual), "key: %q", k)
	}
}

type observedEntry struct {
	Component string
	Message   string
	Timestamp string
	Context   map[string]interface{}
	Error     error
	Verbosity log.Verbosity
	FileLine  string
}

// Fields filters the entry to the specified fields and returns the result as a map.
// This will include all known/parsed fields (such as Message, Timestamp) as well as
// all Context items.
func (o *observedEntry) Fields(fields ...string) Fields {
	entry := o.ToMap()
	filtered := Fields{}

	for _, f := range fields {
		filtered[f] = entry[f]
	}
	return filtered
}

// RawFields drops all but the specified from the entry and returns the encoded result
func (o *observedEntry) RawFields(fields ...string) ([]byte, error) {
	filtered := o.Fields(fields...)
	b := bytes.NewBuffer(nil)
	err := log.JSONEncoder{}.Encode(b, filtered)
	return b.Bytes(), err
}

func (o *observedEntry) ToMap() map[string]interface{} {
	m := make(map[string]interface{}, len(o.Context))
	for k, v := range o.Context {
		m[k] = v
	}
	m[log.ErrorKey] = o.Error
	m[log.MessageKey] = o.Message
	m[log.TimeStampKey] = o.Timestamp
	m[log.ComponentKey] = o.Component
	m[log.LevelKey] = o.Verbosity
	m[log.FileLineKey] = o.FileLine
	return m
}

func (o *observedEntry) Raw() ([]byte, error) {
	entry := o.ToMap()

	b := bytes.NewBuffer(nil)
	err := log.JSONEncoder{}.Encode(b, entry)
	return b.Bytes(), err
}

// observableEncoder stores all entries in a buffer rather than encoding them to an output
type observableEncoder struct {
	entries []*observedEntry
}

// Encode stores all entries in a buffer rather than encoding them to an output
func (o *observableEncoder) Encode(_ io.Writer, m interface{}) error {
	o.entries = append(o.entries, parseEntry(m))
	return nil
}

// Logs returns all logs in the buffer
func (o *observableEncoder) Logs() []*observedEntry {
	return o.entries
}

// TakeAll returns all logs and clears the buffer
func (o *observableEncoder) TakeAll() []*observedEntry {
	defer o.Reset()
	return o.Logs()
}

// Reset empties the buffer
func (o *observableEncoder) Reset() {
	o.entries = make([]*observedEntry, 0)
}

// parseEntry parses all known fields into the observedEntry struct and places
// everything else in the Context field
func parseEntry(entry interface{}) *observedEntry {
	// Make a copy, don't alter the argument as a side effect.
	m := entry.(log.Line)
	verbosity, err := strconv.Atoi(m.Verbosity)
	if err != nil {
		log.DefaultLogger().Error(err, "failed to parse string as verbosity")
	}

	var resultErr error = nil
	if errVal, ok := m.Context[log.ErrorKey]; ok {
		if resultErr, ok = errVal.(error); !ok {
			fmt.Fprintf(os.Stderr, "failed to parse error from message: %v\n", kverrors.New("malformed/missing key", "key", log.ErrorKey))
		}
	}
	delete(m.Context, log.ErrorKey)

	result := &observedEntry{
		Timestamp: m.Timestamp,
		FileLine:  m.FileLine,
		Verbosity: log.Verbosity(verbosity),
		Component: m.Component,
		Message:   m.Message,
		Context:   m.Context,
		Error:     resultErr,
	}

	return result
}

// NewObservedLogger creates a new observableEncoder and a logger that uses the encoder.
func NewObservedLogger() (*observableEncoder, logr.Logger) {
	now := time.Now().UTC().Format(time.RFC3339)
	log.TimestampFunc = func() string {
		return now
	}

	te := &observableEncoder{}
	s := log.NewLogSink("", ioutil.Discard, 0, te)

	return te, logr.New(s)
}
