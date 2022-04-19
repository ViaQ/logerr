package sink_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"testing"

	"github.com/ViaQ/logerr/v2/internal/sink"
	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func TestSink_Enabled(t *testing.T) {
	s, _ := sinkWithBuffer("", 0)

	require.True(t, s.Enabled(0))
	require.False(t, s.Enabled(1))

	ss, _ := sinkWithBuffer("", 1)

	require.True(t, ss.Enabled(1))
	require.True(t, ss.Enabled(0))
	require.False(t, ss.Enabled(2))
}

func TestSink_Info(t *testing.T) {
	msg := "Same or lower than current verbosity. This should be logged."
	msgValue := fmt.Sprintf(`%q:%q`, sink.MessageKey, msg)

	s, b := sinkWithBuffer("", 0)

	s.Info(0, msg)

	logMsg := string(b.Bytes())
	require.Contains(t, logMsg, msgValue)

	// Reset the buffer to ensure to avoid false positive of message
	b.Reset()
	s.Info(0, msg, "hello", "world")

	logMsg = string(b.Bytes())
	require.Contains(t, logMsg, msgValue)
	require.Contains(t, logMsg, fmt.Sprintf(`%q:%q`, "hello", "world"))
}

func TestSink_Info_Enabled(t *testing.T) {
	s, b := sinkWithBuffer("", 1)

	s.Info(2, "Above current verbosity. This should not be logged.")
	require.Empty(t, b.Bytes())

	s.Info(0, "Same or lower than current verbosity. This should be logged.")
	require.NotEmpty(t, b.Bytes())
}

func TestSink_Error(t *testing.T) {
	err := kverrors.New("an error")
	msg := "Error bypasses the enabled check."

	kverrMsg, _ := err.(*kverrors.KVError).MarshalJSON()

	msgValue := fmt.Sprintf(`%q:%q`, sink.MessageKey, msg)
	levelValue := fmt.Sprintf(`%q:"%d"`, sink.LevelKey, 0)
	errorValue := fmt.Sprintf(`%q:%v`, sink.ErrorKey, string(kverrMsg))

	s, b := sinkWithBuffer("", 0)

	s.Error(err, msg)

	logMsg := string(b.Bytes())
	require.NotEmpty(t, logMsg)
	require.Contains(t, logMsg, msgValue)
	require.Contains(t, logMsg, levelValue)
	require.Contains(t, logMsg, errorValue)

	// Reset the buffer to ensure to avoid false positive of message
	b.Reset()

	s.Error(err, msg, "hello", "world")

	logMsg = string(b.Bytes())
	require.NotEmpty(t, logMsg)
	require.Contains(t, logMsg, msgValue)
	require.Contains(t, logMsg, levelValue)
	require.Contains(t, logMsg, errorValue)
	require.Contains(t, logMsg, fmt.Sprintf(`%q:%q`, "hello", "world"))
}

func TestSink_Error_Enabled(t *testing.T) {
	// Use a logger to create a scenario in which the logger's
	// level will be greater than the current verbosity. This means
	// that info will not work. However, error should always work, regardless.
	s, b := sinkWithBuffer("", 0)

	// Logger level: 2, Sink Verbosity: 0
	l := logr.New(s).V(2)
	require.Equal(t, 0, s.GetVerbosity())

	l.Info("Above current verbosity. This should not be logged.")
	require.Empty(t, b.Bytes())

	l.Error(kverrors.New("an error"), "Error bypasses the enabled check.")
	require.NotEmpty(t, b.Bytes())
}

func TestSink_Error_WithNilError(t *testing.T) {
	// Use a logger to create a scenario in which the logger's
	// level will be greater than the current verbosity. This means
	// that info will not work. However, error should always work, regardless.
	s, b := sinkWithBuffer("", 0)

	s.Error(nil, "No error sent. This should be logged.")
	msg := string(b.Bytes())

	require.NotEmpty(t, msg)
}

func TestSink_Error_WithNonKVError(t *testing.T) {
	err := io.ErrClosedPipe
	errValue := fmt.Sprintf(`%q:{"msg":%q}`, sink.ErrorKey, err.Error())

	s, b := sinkWithBuffer("", 0, "hello", "world")

	s.Error(err, "hello, world")
	msg := string(b.Bytes())

	require.NotEmpty(t, msg)
	require.Contains(t, msg, errValue)
}

func TestSink_WithValues(t *testing.T) {
	s, b := sinkWithBuffer("", 0, "hello", "world")

	s.Info(0, "First.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, "hello", "world"))

	ss := s.WithValues("foo", "bar")

	ss.Error(nil, "Second.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, "hello", "world"))
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, "foo", "bar"))
}

func TestSink_WithValues_NoKeyValues(t *testing.T) {
	s, b := sinkWithBuffer("", 0, nil)
	ss := s.WithValues("foo", "bar")

	ss.Error(nil, "Second.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, "foo", "bar"))
}

func TestSink_WithKeyAndNoValues(t *testing.T) {
	s, b := sinkWithBuffer("", 0, "hello")

	s.Info(0, "First.")

	// Ensuring that dangling key/values are not recorded
	require.NotContains(t, string(b.Bytes()), fmt.Sprintf(`%q:`, "hello"))
}

func TestSink_WithName(t *testing.T) {
	s, b := sinkWithBuffer("new", 0)

	s.Info(0, "First.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, sink.ComponentKey, "new"))

	ss := s.WithName("append")

	ss.Info(0, "Second.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, sink.ComponentKey, "new_append"))
}

func TestSink_WithEmptyName(t *testing.T) {
	s, b := sinkWithBuffer("", 0)

	s.Info(0, "First.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:""`, sink.ComponentKey))

	ss := s.WithName("new")

	ss.Info(0, "Second.")
	require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:%q`, sink.ComponentKey, "new"))
}

func TestSink_SetOutput(t *testing.T) {
	bb := bytes.NewBuffer(nil)

	s, b := sinkWithBuffer("", 0)
	s.SetOutput(bb)

	s.Info(0, "Same or lower than current verbosity. This should be logged.")
	require.Empty(t, b.Bytes())
	require.NotEmpty(t, bb.Bytes())
}

func TestSink_SetVerbosity(t *testing.T) {
	s, b := sinkWithBuffer("", 0)

	s.Info(1, "Above current verbosity. This should not be logged.")
	require.Empty(t, b.Bytes())

	s.SetVerbosity(1)

	s.Info(1, "Same or lower than current verbosity. This should be logged.")
	require.NotEmpty(t, b.Bytes())
}

func TestSink_GetVerbosity(t *testing.T) {
	s, _ := sinkWithBuffer("", 0)
	require.Equal(t, 0, s.GetVerbosity())

	s.SetVerbosity(10)
	require.Equal(t, 10, s.GetVerbosity())
}

func TestSink_Log_EncodingFailure(t *testing.T) {
	unsupportedValues := []float64{
		math.NaN(),
		math.Inf(-1),
		math.Inf(1),
	}

	for _, unsupportedValue := range unsupportedValues {
		s, b := sinkWithBuffer("", 0)

		// Info and Error use the same mechanism, so using Info for the test
		s.Info(0, "Test unsupported value", "value", unsupportedValue)
		msg := string(b.Bytes())

		require.NotEmpty(t, msg)
		require.Contains(t, msg, "failed to encode message")
		require.Contains(t, msg, fmt.Sprintf("json: unsupported value: %f", unsupportedValue))
	}
}

func sinkWithBuffer(component string, level int, keyValuePairs ...interface{}) (*sink.Sink, *bytes.Buffer) {
	buffer := bytes.NewBuffer(nil)
	sink := sink.NewLogSink(component, buffer, sink.Verbosity(level), sink.JSONEncoder{}, keyValuePairs...)

	return sink, buffer
}
