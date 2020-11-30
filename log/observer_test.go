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

	"github.com/ViaQ/logerr/internal/kv"
	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
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
	return kv.AppendMap(o.Context, map[string]interface{}{
		log.ErrorKey:     o.Error,
		log.MessageKey:   o.Message,
		log.TimeStampKey: o.Timestamp,
		log.ComponentKey: o.Component,
		log.LevelKey:     o.Verbosity,
	})
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
func (o *observableEncoder) Encode(_ io.Writer, m map[string]interface{}) error {
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
func parseEntry(m map[string]interface{}) *observedEntry {
	result := &observedEntry{}
	var ok bool

	result.Message, ok = m[log.MessageKey].(string)
	delete(m, log.MessageKey)
	if !ok {
		log.Error(kverrors.New("malformed/missing key", "key", log.MessageKey), "failed to parse message key from message")
	}

	result.Timestamp, ok = m[log.TimeStampKey].(string)
	delete(m, log.TimeStampKey)
	if !ok {
		log.Error(kverrors.New("malformed/missing key", "key", log.TimeStampKey), "failed to parse timestamp from message")
	}

	if erri, ok := m[log.ErrorKey]; ok {
		if result.Error, ok = erri.(error); !ok {
			fmt.Fprintf(os.Stderr, "failed to parse error from message: %v\n", kverrors.New("malformed/missing key", "key", log.ErrorKey))
		}
	}
	delete(m, log.ErrorKey)

	result.Component, ok = m[log.ComponentKey].(string)
	delete(m, log.ComponentKey)
	if !ok {
		log.Error(kverrors.New("malformed/missing key", "key", log.ComponentKey), "failed to parse component from message")
	}

	verbosityStr, ok := m[log.LevelKey].(string)
	delete(m, log.LevelKey)
	if !ok {
		log.Error(kverrors.New("malformed/missing key", "key", log.LevelKey), "failed to parse level from message")
	}

	verbosity, err := strconv.Atoi(verbosityStr)
	if err != nil {
		log.Error(err, "failed to parse string as verbosity")
	} else {
		result.Verbosity = log.Verbosity(verbosity)
	}


	// set the remaining items to Context
	result.Context = m
	return result
}

// NewObservedLogger creates a new observableEncoder and a logger that uses the encoder.
func NewObservedLogger() (*observableEncoder, *log.Logger) {
	now := time.Now().UTC().Format(time.RFC3339)
	log.TimestampFunc = func() string {
		return now
	}

	te := &observableEncoder{}

	ll := log.NewLogger("", ioutil.Discard, 0, te)

	return te, ll
}
