package log_test

import (
	"bytes"
	"fmt"

	"testing"

	"github.com/ViaQ/logerr/v2/internal/sink"
	"github.com/ViaQ/logerr/v2/log"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	level := 1
	component := "mycomponent"
	componentValue := fmt.Sprintf(`%q:%q`, sink.ComponentKey, component)

	b := bytes.NewBuffer(nil)
	bb := bytes.NewBuffer(nil)

	l := log.NewLogger(component, log.WithOutput(b))
	l.Info("Default configuration. Though still need to override the stdout writer to test.")
	msg := string(b.Bytes())

	require.Contains(t, msg, componentValue)
	require.Contains(t, msg,fmt.Sprintf(`%q:"%d"`, sink.LevelKey, 0))

	ll := log.NewLogger(component, log.WithOutput(bb), log.WithVerbosity(level))
	ll.Info("Non-default configuration.")
	msg = string(bb.Bytes())

	require.Contains(t, msg, componentValue)
	require.Contains(t, msg,fmt.Sprintf(`%q:"%d"`, sink.LevelKey, level))
}
