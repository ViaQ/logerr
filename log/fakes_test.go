package log_test

import (
	"github.com/go-logr/logr"
)

type nopLogger struct{}

func (nopLogger) Enabled() bool {
	return false
}

func (nopLogger) Info(string, ...interface{}) {
}

func (nopLogger) Error(error, string, ...interface{}) {
}

func (nopLogger) V(int) logr.Logger {
	return nil
}

func (nopLogger) WithValues(...interface{}) logr.Logger {
	return nil
}

func (nopLogger) WithName(string) logr.Logger {
	return nil
}
