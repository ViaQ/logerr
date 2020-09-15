package log

import (
	"github.com/go-logr/logr"
)

type fakeLogger struct {
	EnabledFunc    func() bool
	InfoFunc       func(msg string, keysAndValues ...interface{})
	ErrorFunc      func(err error, msg string, keysAndValues ...interface{})
	VFunc          func(level int) logr.Logger
	WithValuesFunc func(keysAndValues ...interface{}) logr.Logger
	WithNameFunc   func(name string) logr.Logger
}

func (f *fakeLogger) Enabled() bool {
	if f.EnabledFunc != nil {
		return f.EnabledFunc()
	}
	return false
}

func (f *fakeLogger) Info(msg string, keysAndValues ...interface{}) {
	if f.InfoFunc != nil {
		f.InfoFunc(msg, keysAndValues...)
	}
}

func (f *fakeLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if f.ErrorFunc != nil {
		f.ErrorFunc(err, msg, keysAndValues...)
	}
}

func (f *fakeLogger) V(level int) logr.Logger {
	if f.VFunc != nil {
		return f.VFunc(level)
	}
	return nil
}

func (f *fakeLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	if f.WithValuesFunc != nil {
		return f.WithValuesFunc(keysAndValues...)
	}
	return nil
}

func (f *fakeLogger) WithName(name string) logr.Logger {
	if f.WithNameFunc != nil {
		return f.WithNameFunc(name)
	}
	return nil
}
