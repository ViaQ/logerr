package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ViaQ/logerr/internal/kv"
	"github.com/ViaQ/logerr/log"
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
	log.SetOutput(buf)
	ll, ok := log.GetLogger().(*log.Logger)
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
	msg := "asjklhasdfhuasfu"

	log.Info(msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
}

func TestError(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	msg := "iou3jkhcopiuyasdf"
	err := errors.New("fail boat")

	log.Error(err, msg)

	logs := obs.Logs()
	require.Len(t, logs, 1)
	require.EqualValues(t, msg, logs[0].Message)
	require.EqualValues(t, err.Error(), logs[0].Error.Error())
}

func TestWithValues(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	msg := "yuio3hjk4reklcyasdf"
	ctx := map[string]interface{}{
		"left":  "right",
		"hello": "world",
	}

	ll := log.WithValues(kv.FromMap(ctx)...)

	assertions := func(t *testing.T) {
		logs := obs.TakeAll()
		if assert.Len(t, logs, 1) {
			assert.EqualValues(t, msg, logs[0].Message)
			assert.EqualValues(t, ctx, logs[0].Context)
		}
	}

	t.Run("Error", func(t *testing.T) {
		ll.Error(errors.New("fail boat"), msg)
		assertions(t)
	})

	t.Run("Info", func(t *testing.T) {
		ll.Info(msg, kv.FromMap(ctx)...)
		assertions(t)
	})
}

func TestSetLogLevel(t *testing.T) {
	obs, logger := NewObservedLogger()
	log.UseLogger(logger)

	const logLevel = 4
	msg := "iohu3hjk4o8iuas9df8"

	log.SetLogLevel(logLevel)
	log.V(logLevel).Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.EqualValues(t, msg, logs[0].Message)
}

func TestSetOutput(t *testing.T) {
	logger := log.NewLogger("", ioutil.Discard, 0, log.JSONEncoder{})
	log.UseLogger(logger)

	msg := "ahjklsdfhuiasdfh"

	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	log.Info(msg)

	output := string(buf.Bytes())
	require.NotEmpty(t, output)

	require.Contains(t, output, msg)
}

func TestWithName(t *testing.T) {
	obs, _ := NewObservedLogger()

	logger := log.NewLogger("", ioutil.Discard, 0, obs)
	logger = logger.WithName("mycomponent").(*log.Logger)
	log.UseLogger(logger)

	msg := "uioy3948dc9ad"

	ll := log.WithName("mynameis")

	ll.Info(msg)

	logs := obs.TakeAll()
	require.NotEmpty(t, logs)

	require.Contains(t, logs[0].Component, "mycomponent")
	require.Contains(t, logs[0].Component, "mynameis")
}

func TestV(t *testing.T) {
}
