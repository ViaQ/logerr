package log_test

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/ViaQ/logerr/errors"
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

	err := errors.New("an error")
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

	err := errors.New("an error")
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

	err := errors.New("an error", "key", "value")
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

	err1 := errors.New("error1", "order", 1)
	err := errors.Wrap(err1, "main error", "key", "value")
	log.Error(err, "hello, world")

	logs := obs.All()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: map[string]interface{}{
				"key": "value",
				"msg": err.Message(),
				log.KeyError: map[string]interface{}{
					"order": 1,
					"msg": err1.Message(),
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
