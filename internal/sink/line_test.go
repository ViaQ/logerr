package sink_test

import (
	"fmt"
	"testing"

	"github.com/ViaQ/logerr/v2/internal/sink"
	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/stretchr/testify/require"
)

func TestLine_ProductionLogs(t *testing.T) {
	// Logs with a level of 0 or 1 are considered to be production level logs
	msg := "hello, world"
	s, b := sinkWithBuffer("", 1)

	s.Info(0, msg)
	logMsg := string(b.Bytes())

	require.NotEmpty(t, logMsg)
	require.Contains(t, logMsg, fmt.Sprintf(`%q:%q`, sink.MessageKey, msg))
	require.NotContains(t, logMsg, fmt.Sprintf(`%q`, sink.FileLineKey))
}

func TestLine_DeveloperLogs(t *testing.T) {
	// Logs with a higher level than 1 are considered to be developer level logs
	msg := "hello, world"
	s, b := sinkWithBuffer("", 2)

	s.Info(1, msg)
	logMsg := string(b.Bytes())

	require.NotEmpty(t, logMsg)
	require.Contains(t, logMsg, fmt.Sprintf(`%q:%q`, sink.MessageKey, msg))
	require.Contains(t, logMsg, fmt.Sprintf(`%q`, sink.FileLineKey))
}

func TestLine_WithNoContext(t *testing.T) {
	msg := "hello, world"
	l := sink.Line{Message: msg}

	b, err := l.MarshalJSON()

	require.NoError(t, err)
	require.Contains(t, string(b), fmt.Sprintf(`%q:%q`, sink.MessageKey, msg))
}

func TestLine_LogLevel(t *testing.T) {
	s, b := sinkWithBuffer("", 0)

	for level := 0; level < 5; level++ {
		b.Reset()
		s.SetVerbosity(level)

		s.Info(0, "hello, world")

		require.Contains(t, string(b.Bytes()), fmt.Sprintf(`%q:"%d"`, sink.LevelKey, level))
	}
}

func TestLine_WithKVError(t *testing.T) {
	err := kverrors.New("an error", "key", "value")
	msg := "Error bypasses the enabled check."

	kverrMsg, _ := err.(*kverrors.KVError).MarshalJSON()

	msgValue := fmt.Sprintf(`%q:%q`, sink.MessageKey, msg)
	errorValue := fmt.Sprintf(`%q:%v`, sink.ErrorKey, string(kverrMsg))

	s, b := sinkWithBuffer("", 0)

	s.Error(err, msg)

	logMsg := string(b.Bytes())
	require.Contains(t, logMsg, msgValue)
	require.Contains(t, logMsg, errorValue)
}

func TestLine_WithNestedKVError(t *testing.T) {
	err := kverrors.New("an error", "key", "value")
	wrappedErr := kverrors.Wrap(err, "main error", "key", "value")
	msg := "Error bypasses the enabled check."

	kverrMsg, _ := wrappedErr.(*kverrors.KVError).MarshalJSON()

	msgValue := fmt.Sprintf(`%q:%q`, sink.MessageKey, msg)
	errorValue := fmt.Sprintf(`%q:%v`, sink.ErrorKey, string(kverrMsg))

	s, b := sinkWithBuffer("", 0)

	s.Error(wrappedErr, msg)

	logMsg := string(b.Bytes())
	require.Contains(t, logMsg, msgValue)
	require.Contains(t, logMsg, errorValue)
}
