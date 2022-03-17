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
		log.MustInitWithOptions(testName, []log.Option{
			log.WithOutput(ioutil.Discard),
			log.WithLogLevel(verbosity),
		})
		t.Run(testName, func(t *testing.T) {
			for logLevel := 1; logLevel < 5; logLevel++ {
				log.V(logLevel).Info("hello, world")
			}
		})
	}
}

func TestInit(t *testing.T) {
	component := "mycomponent"
	buf := bytes.NewBuffer(nil)

	log.MustInit(component)
	require.NoError(t, log.SetOutput(buf))

	ll := log.GetLogger()
	_, ok := ll.GetSink().(*log.LogSink)
	require.True(t, ok)

	ll.Info("laskdjfhiausdc")

	expected := fmt.Sprintf(`%q:%q`, log.ComponentKey, component)

	actual := string(buf.Bytes())

	require.Contains(t, actual, expected)
}

func TestUseLogger_SetsLogger(t *testing.T) {
	_, logger := NewObservedLogger()
	log.UseLogger(logger)
	require.Equal(t, logger, log.GetLogger())
}

func TestInfo(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)
	msg := t.Name()

	log.Info(msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
}

func TestError(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	msg := t.Name()
	err := errors.New("fail boat")

	log.Error(err, msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
	require.NotNil(t, logs[0].Error)
	require.EqualValues(t, err.Error(), logs[0].Error.Error())
}

func TestWithValues(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	msg := t.Name()
	ll := log.WithValues("hello", "world")

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
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	const logLevel = 4
	msg := t.Name()

	log.SetLogLevel(logLevel)
	log.V(logLevel).Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.EqualValues(t, msg, logs[0].Message)
}

func TestSetOutput_WithKnownLogger_SetsOutputOnLogger(t *testing.T) {
	logger := logr.New(log.NewLogSink("", ioutil.Discard, 0, log.JSONEncoder{}))
	log.UseLogger(logger)

	msg := t.Name()

	buf := bytes.NewBuffer(nil)
	require.NoError(t, log.SetOutput(buf))
	log.Info(msg)

	output := string(buf.Bytes())
	require.NotEmpty(t, output)

	require.Contains(t, output, msg)
}

func TestSetOutput_WithUnknownLogger_Errors(t *testing.T) {
	log.UseLogger(logr.New(nopLogSink{}))

	buf := bytes.NewBuffer(nil)
	err := log.SetOutput(buf)

	actual := kverrors.Root(err)
	require.Equal(t, log.ErrUnknownLoggerType, actual)
}

func TestWithName(t *testing.T) {
	obs, _ := NewObservedLogger()

	logger := logr.New(log.NewLogSink("", ioutil.Discard, 0, obs))
	logger = logger.WithName("mycomponent")
	log.UseLogger(logger)

	msg := t.Name()

	ll := log.WithName("mynameis")

	ll.Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.Contains(t, logs[0].Component, "mycomponent")
	require.Contains(t, logs[0].Component, "mynameis")
}

func TestV(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)
	log.SetLogLevel(1)

	msg := t.Name()

	log.V(1).Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)
	require.Equal(t, msg, logs[0].Message)
}
