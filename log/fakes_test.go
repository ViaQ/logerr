package log_test

import (
	"github.com/go-logr/logr"
)

type nopLogSink struct{}

func (nopLogSink) Init(logr.RuntimeInfo) {
}

func (nopLogSink) Enabled(int) bool {
	return false
}

func (nopLogSink) Info(int, string, ...interface{}) {
}

func (nopLogSink) Error(error, string, ...interface{}) {
}

func (nopLogSink) WithValues(...interface{}) logr.LogSink {
	return nil
}

func (nopLogSink) WithName(string) logr.LogSink {
	return nil
}
