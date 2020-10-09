package log_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInfo_nokvs(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	log.Info("hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)
	assert.Len(t, logs[0].Context, 0)
}

func TestInfo_kvs(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	log.Info("hello, world", "city", "Athens")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)
	require.Len(t, logs[0].Context, 1)

	expected := []zap.Field{zap.Any("city", "Athens")}
	require.EqualValues(t, logs[0].Context, expected)
}

func TestError_nokvs(t *testing.T) {
	core, obs := observer.New(zapcore.ErrorLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	err := kverrors.New("an error")
	log.Error(err, "hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: err,
		},
	)
}

func TestError_kvs(t *testing.T) {
	core, obs := observer.New(zapcore.ErrorLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	err := kverrors.New("an error")
	log.Error(err, "hello, world", "key", "value")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: err,
			"key":        "value",
		},
	)
}

func TestError_pkg_error_kvs(t *testing.T) {
	core, obs := observer.New(zapcore.ErrorLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	err := kverrors.New("an error", "key", "value")
	log.Error(err, "hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: map[string]interface{}{
				"key": "value",
				"msg": err.Error(),
			},
		},
	)
}

func TestError_nested_error(t *testing.T) {
	core, obs := observer.New(zapcore.ErrorLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	err1 := kverrors.New("error1", "order", 1)
	err := kverrors.Wrap(err1, "main error", "key", "value")
	log.Error(err, "hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: map[string]interface{}{
				"key": "value",
				"msg": kverrors.Message(err),
				log.KeyError: map[string]interface{}{
					"order": 1,
					"msg":   kverrors.Message(err1),
				},
			},
		},
	)
}

func TestWithValues_AddsValues(t *testing.T) {
	core, obs := observer.New(zapcore.ErrorLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	err := io.ErrClosedPipe
	log.WithValues("key", "value").Error(err, "hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			"key": "value",
			log.KeyError: map[string]interface{}{
				"msg": err.Error(),
			},
		},
	)
}

func TestLogger_V(t *testing.T) {
	for verbosity := uint8(1); verbosity < 5; verbosity++ {
		cnf := &zap.Config{Level: zap.NewAtomicLevelAt(zapcore.PanicLevel)}
		// We could easily just configure this by doing -1 * verbosity, but we should
		// intentionally bypass the simple configuration and use the log.WithVerbosity to
		// guarantee log.WithVerbosity works correctly. It manipulates cnf and we then
		// use cnf.Level below. If WithVerbosity changes somehow then this test will break
		log.WithVerbosity(verbosity)(cnf)
		core, obs := observer.New(cnf.Level)

		log.UseLogger(zapr.NewLogger(zap.New(core)))

		// loop through log levels 1-5 and log all of them to verify that they either
		// are or are not logged according to verbosity above
		for logLevel := uint8(1); logLevel < 5; logLevel++ {
			log.V(logLevel).Info("hello, world")
			logs := obs.TakeAll()

			shouldBeLogged := verbosity >= logLevel

			if shouldBeLogged {
				assert.Len(t, logs, 1, "expected log to be present for verbosity:%d, logLevel:%d", verbosity, logLevel)
				assert.EqualValues(t, "hello, world", logs[0].Message)
			} else {
				assert.Empty(t, logs, "expected NO logs to be present for verbosity:%d, logLevel:%d", verbosity, logLevel)
			}
		}
	}
}

// This test exists to confirm that the output actually works. There was a previous bug that broke the real logger
// because DefaultConfig specified a sampling which caused a panic. This uses a real logger and logs just to verify
// that it _can_ log successfully. There are no assertions because the content of the logs are irrelevant. See
// TestLogger_V above for a more comprehensive test.
func TestLogger_V_Integration(t *testing.T) {
	for verbosity := uint8(1); verbosity < 5; verbosity++ {
		verbosity := verbosity
		testName := fmt.Sprintf("verbosity-%d", verbosity)
		t.Run(testName, func(t *testing.T) {
			log.MustInitWithOptions(testName, []log.Option{
				log.WithVerbosity(verbosity),
				log.WithNoOutputs(),
			})
			for logLevel := uint8(1); logLevel < 5; logLevel++ {
				log.V(logLevel).Info("hello, world")
			}
		})
	}
}

// assertLoggedFields checks that each field exists in the LoggedEntry
func assertLoggedFields(t *testing.T, entry observer.LoggedEntry, fields map[string]interface{}) {
	ctx := entry.ContextMap()
	for k, v := range fields {
		value, ok := ctx[k]
		require.True(t, ok, "expected key '%s:%v' to exist in logged entry %+v", k, v, entry)
		actual, e := json.Marshal(value)
		require.NoError(t, e)
		expected, e := json.Marshal(v)
		require.NoError(t, e)
		require.JSONEq(t, string(expected), string(actual))
	}
}
