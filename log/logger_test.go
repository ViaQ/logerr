package log_test

import (
	"io"
	"testing"

	"github.com/ViaQ/logerr/log"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger_Error_MakesUnstructuredErrorsStructured(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	log.Error(io.ErrClosedPipe, t.Name())

	logs := obs.All()
	assert.Len(t, logs, 1)

	assertLoggedFields(t,
		logs[0],
		map[string]interface{}{
			log.KeyError: map[string]interface{}{
				"msg": io.ErrClosedPipe.Error(),
			},
		},
	)
}

func TestLogger_Error_WorksWithNilError(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	log.UseLogger(zapr.NewLogger(zap.New(core)))

	log.Error(nil, t.Name())

	logs := obs.TakeAll()
	assert.Len(t, logs, 1)
	assert.NotContains(t, logs[0].ContextMap(), log.KeyError)
}