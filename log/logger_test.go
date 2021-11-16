package log_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"reflect"
	"testing"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_Info_WithKeysAndValues(t *testing.T) {
	obs, logger := NewObservedLogger()

	logger.Info("hello, world", "city", "Athens")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			"city": "Athens",
		},
	)
}

func TestLogger_Error_noKeysAndValues(t *testing.T) {
	obs, logger := NewObservedLogger()

	err := kverrors.New("an error")
	logger.Error(err, "hello, world")

	logs := obs.TakeAll()
	require.Len(t, logs, 1)
	require.EqualValues(t, "hello, world", logs[0].Message)
	require.EqualValues(t, err, logs[0].Error)
}

var unsupportedValues = []float64{
	math.NaN(),
	math.Inf(-1),
	math.Inf(1),
}

func TestLogger_Info_UnsupportedValues(t *testing.T) {
	for _, unsupportedValue := range unsupportedValues {
		buf := bytes.NewBuffer(nil)
		logger := log.NewLogger("", buf, 0, log.JSONEncoder{})
		logger.Info("Test unsupported value", "value", unsupportedValue)

		if buf.Len() == 0 {
			t.Fatal("expected log output, but buffer was empty")
		}
		assert.Contains(t, string(buf.Bytes()), "failed to encode message")
		assert.Contains(t, string(buf.Bytes()), fmt.Sprintf("json: unsupported value: %f", unsupportedValue))
	}
}

func TestLogger_Error_UnsupportedValues(t *testing.T) {
	for _, unsupportedValue := range unsupportedValues {
		buf := bytes.NewBuffer(nil)
		logger := log.NewLogger("", buf, 0, log.JSONEncoder{})
		err := kverrors.New("an error")
		logger.Error(err, "Test unsupported value", "key", unsupportedValue)

		if buf.Len() == 0 {
			t.Fatal("expected log output, but buffer was empty")
		}
		println(string(buf.Bytes()))
		assert.Contains(t, string(buf.Bytes()), "failed to encode message")
		assert.Contains(t, string(buf.Bytes()), fmt.Sprintf("json: unsupported value: %f", unsupportedValue))
	}
}

func TestLogger_Error_KeysAndValues(t *testing.T) {
	obs, logger := NewObservedLogger()

	err := kverrors.New("an error")
	logger.Error(err, "hello, world", "key", "value")

	logs := obs.TakeAll()
	require.Len(t, logs, 1)
	require.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			"key": "value",
		},
	)
}

func TestLogger_Error_pkg_error_KeysAndValues(t *testing.T) {
	obs, logger := NewObservedLogger()

	err := kverrors.New("an error", "key", "value")
	logger.Error(err, "hello, world")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)
	require.EqualValues(t, err, logs[0].Error)

	assertLoggedFields(t,
		logs[0],
		Fields{
			log.MessageKey: "hello, world",
			log.ErrorKey: map[string]interface{}{
				"msg": "an error",
				"key": "value",
			},
		},
	)
}

func TestLogger_Error_nested_error(t *testing.T) {
	obs, logger := NewObservedLogger()

	err1 := kverrors.New("error1", "order", 1)
	err := kverrors.Wrap(err1, "main error", "key", "value")
	logger.Error(err, "hello, world")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			log.ErrorKey: map[string]interface{}{
				"key": "value",
				"msg": kverrors.Message(err),
				kverrors.CauseKey: map[string]interface{}{
					"order": 1,
					"msg":   kverrors.Message(err1),
				},
			},
		},
	)
}

func TestLogger__PlainErrors_ConvertedToStructured(t *testing.T) {
	obs, logger := NewObservedLogger()

	err := io.ErrClosedPipe
	logger.Error(err, "hello, world")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			log.ErrorKey: map[string]interface{}{
				"msg": err.Error(),
			},
		},
	)
}

func TestLogger_WithValues_AddsValues(t *testing.T) {
	obs, logger := NewObservedLogger()

	err := io.ErrClosedPipe
	ll := logger.WithValues("key", "value")

	ll.Error(err, "hello, world")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			"key": "value",
		},
	)
}

func TestLogger_Error_MakesUnstructuredErrorsStructured(t *testing.T) {
	obs, logger := NewObservedLogger()

	logger.Error(io.ErrClosedPipe, t.Name())

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)

	assertLoggedFields(t,
		logs[0],
		Fields{
			log.ErrorKey: map[string]interface{}{
				"msg": io.ErrClosedPipe.Error(),
			},
		},
	)
}

func TestLogger_Error_WorksWithNilError(t *testing.T) {
	obs, logger := NewObservedLogger()

	logger.Error(nil, t.Name())

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.Nil(t, logs[0].Error)
}

func TestLogger_V_Info(t *testing.T) {
	for verbosity := 1; verbosity < 5; verbosity++ {
		log.SetLogLevel(verbosity)

		// loop through log levels 1-5 and log all of them to verify that they either
		// are or are not logged according to verbosity above
		for logLevel := 1; logLevel < 5; logLevel++ {
			obs, logger := NewObservedLogger()

			logger.V(logLevel).Info("hello, world")

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

func TestLogger_V_Error(t *testing.T) {
	for verbosity := 1; verbosity < 5; verbosity++ {
		log.SetLogLevel(verbosity)

		// loop through log levels 1-5 and log all of them to verify that they either
		// are or are not logged according to verbosity above
		for logLevel := 1; logLevel < 5; logLevel++ {
			obs, logger := NewObservedLogger()

			logger.V(logLevel).Error(io.ErrUnexpectedEOF, "hello, world")

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

func TestLogger_SetsVerbosity(t *testing.T) {
	obs, logger := NewObservedLogger()

	logger.Info(t.Name())

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, 0, logs[0].Verbosity)
}

func TestLogger_TestSetOutput(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger("", ioutil.Discard, 0, log.JSONEncoder{})
	logger.SetOutput(buf)

	msg := "hello, world"
	logger.Info(msg)

	if buf.Len() == 0 {
		t.Fatal("expected log output, but buffer was empty")
	}
	assert.Contains(t, string(buf.Bytes()), fmt.Sprintf(`%q:%q`, log.MessageKey, msg))
}

func TestLogger_Info_PrintsError_WhenEncoderErrors(t *testing.T) {
	err := io.ErrShortBuffer
	fenc := fakeEncoder{
		EncodeFunc: func(_ io.Writer, _ interface{}) error {
			return &json.MarshalerError{
				Type: reflect.TypeOf(&json.MarshalerError{}),
				Err:  err,
			}
		},
	}

	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger("", buf, 0, fenc)

	msg := "hello, world"
	logger.Info(msg)

	if buf.Len() == 0 {
		t.Fatal("expected buffer output, but buffer was empty")
	}

	output := string(buf.Bytes())
	assert.Contains(t, output, msg, "has the original message")
	assert.Contains(t, output, err.Error(), "shows the original error")
	assert.Contains(t, output, reflect.TypeOf(fenc).String(), "explains the encoder that failed")
}

func TestLogger_LogsLevel(t *testing.T) {
	const v = 2

	obs, logger := NewObservedLogger()
	log.SetLogLevel(v)

	logger.V(v).Info("hello, world", "city", "Athens")

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.EqualValues(t, "hello, world", logs[0].Message)

	assertLoggedFields(t,
		logs[0],
		Fields{
			log.LevelKey: v,
		},
	)
}

func TestLogger_ProductionLogsLevel(t *testing.T) {
	const v = 0

	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger("", ioutil.Discard, v, log.JSONEncoder{})
	logger.SetOutput(buf)

	msg := "hello, world"
	logger.Info(msg)

	if buf.Len() == 0 {
		t.Fatal("expected log output, but buffer was empty")
	}
	assert.Contains(t, string(buf.Bytes()), fmt.Sprintf(`%q:%q`, log.MessageKey, msg))
	assert.NotContains(t, string(buf.Bytes()), fmt.Sprintf(`%q`, log.FileLineKey))
}

func TestLogger_DeveloperLogsLevel(t *testing.T) {
	const v = 2

	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger("", ioutil.Discard, v, log.JSONEncoder{})
	logger.SetOutput(buf)

	msg := "hello, world"
	logger.Info(msg)

	if buf.Len() == 0 {
		t.Fatal("expected log output, but buffer was empty")
	}
	assert.Contains(t, string(buf.Bytes()), fmt.Sprintf(`%q:%q`, log.MessageKey, msg))
	assert.Contains(t, string(buf.Bytes()), fmt.Sprintf(`%q`, log.FileLineKey))
}

func TestLogger_LogLineWithNoContext(t *testing.T) {
	msg := "hello, world"
	l := log.Line{
		Message: msg,
	}

	buf,err := l.MarshalJSON()
	assert.Nil(t, err)
	assert.Contains(t, string(buf), fmt.Sprintf(`%q:%q`, log.MessageKey,msg))
}
