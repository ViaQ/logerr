package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ViaQ/logerr/internal/kv"

	"io/ioutil"
	"testing"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test exists to confirm that the output actually works. There was a previous bug that broke the real logger
// because DefaultConfig specified a sampling which caused a panic. This uses a real logger and logs just to verify
// that it _can_ log successfully. There are no assertions because the content of the logs are irrelevant. See
// TestLogger_V above for a more comprehensive test.
func TestLogger_V_Integration(t *testing.T) {
	for i := 1; i < 5; i++ {
		verbosity := i
		testName := fmt.Sprintf("verbosity-%d", verbosity)
		l := log.NewLoggerWithOptions(testName, []log.Option{
			log.WithOutput(ioutil.Discard),
			log.WithLogLevel(verbosity),
		})
		t.Run(testName, func(t *testing.T) {
			for logLevel := 1; logLevel < 5; logLevel++ {
				l.V(logLevel).Info("hello, world")
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	component := "mycomponent"
	buf := bytes.NewBuffer(nil)

	l := log.NewLogger(component)
	s, err := log.GetSink(l)

	require.NoError(t, err)

	s.SetOutput(buf)
	l.Info("laskdjfhiausdc")

	expected := fmt.Sprintf(`%q:%q`, log.ComponentKey, component)
	actual := string(buf.Bytes())

	require.Contains(t, actual, expected)
}

func TestInfo(t *testing.T) {
	obs, l := NewObservedLogger()

	msg := t.Name()
	l.Info(msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
}

func TestError(t *testing.T) {
	obs, l := NewObservedLogger()

	msg := t.Name()
	err := errors.New("fail boat")

	l.Error(err, msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
	require.NotNil(t, logs[0].Error)
	require.EqualValues(t, err.Error(), logs[0].Error.Error())
}

func TestWithValues(t *testing.T) {
	obs, l := NewObservedLogger()

	msg := t.Name()
	ll := l.WithValues("hello", "world")

	t.Run("Error", func(t *testing.T) {
		ll.Error(errors.New("fail boat"), msg)
		logs := obs.TakeAll()
		require.Len(t, logs, 1)
		assert.EqualValues(t, msg, logs[0].Message)
		assert.EqualValues(t, kv.ToMap("hello", "world"), logs[0].Context)
	})

	t.Run("Info", func(t *testing.T) {
		ll.Info(msg, "a", "b")
		ll.Info(msg, "c", "d")
		logs := obs.TakeAll()
		require.Len(t, logs, 2)
		assert.EqualValues(t, msg, logs[0].Message)
		assert.EqualValues(t, kv.ToMap("a", "b", "hello", "world"), logs[0].Context)
		assert.EqualValues(t, kv.ToMap("c", "d", "hello", "world"), logs[1].Context)
	})
}

func TestSetLogLevel(t *testing.T) {
	obs, l := NewObservedLogger()
	s, err := log.GetSink(l)

	require.NoError(t, err)

	const logLevel = 4
	msg := t.Name()

	s.SetVerbosity(logLevel)
	l.V(logLevel).Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.EqualValues(t, msg, logs[0].Message)
}

func TestSetOutput(t *testing.T) {
	obs, l := NewObservedLogger()
	s, err := log.GetSink(l)

	require.NoError(t, err)

	msg := t.Name()
	buf := bytes.NewBuffer(nil)

	s.SetOutput(buf)

	l.Info(msg)
	logs := obs.TakeAll()

	require.NotEmpty(t, logs)
	require.Contains(t, logs[0].Message, msg)
}

func TestGetSink_WithUnknownLogSink_Errors(t *testing.T) {
	l := logr.New(nopLogSink{})
	_, err := log.GetSink(l)

	actual := kverrors.Root(err)
	require.Equal(t, log.ErrUnknownSinkType, actual)
}

func TestWithName(t *testing.T) {
	obs, _ := NewObservedLogger()

	l := logr.New(log.NewLogSink("", ioutil.Discard, 0, obs)).WithName("mycomponent")
	ll := l.WithName("mynameis")

	msg := t.Name()
	ll.Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.Contains(t, logs[0].Component, "mycomponent")
	require.Contains(t, logs[0].Component, "mynameis")
}

func TestV(t *testing.T) {
	obs, l := NewObservedLogger()
	s, err := log.GetSink(l)

	require.NoError(t, err)

	msg := t.Name()

	s.SetVerbosity(1)
	l.V(1).Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)
	require.Equal(t, msg, logs[0].Message)
}
