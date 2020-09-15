package log

import (
	"sync"

	"github.com/ViaQ/logerr/pkg/errors"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	keyError = "cause"
)

var (
	mtx sync.RWMutex

	// empty logger to prevent nil pointer panics before Init is called
	logger = zapr.NewLogger(zap.NewNop())

	defaultConfig = &zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
		DisableStacktrace: true,
	}
)

type Option func(*zap.Config)

// WithNoTimestamp removes the timestamp from the logged output
// this is primarily used for testing purposes
func WithNoTimestamp() Option {
	return func(c *zap.Config) {
		c.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("")
	}
}


// Init initializes the logger. This is required to use logging correctly
// component is the name of the component being used to log messages. Typically this is your application name
// keyValuePairs are default key/value pairs to be used with all logs in the future
func Init(component string, opts []Option, keyValuePairs ...interface{}) error {
	mtx.Lock()
	defer mtx.Unlock()

	defaultConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	defaultConfig.EncoderConfig.TimeKey = "time"
	for _, opt := range opts {
		opt(defaultConfig)
	}

	zl, err := defaultConfig.Build(zap.AddCallerSkip(2))
	if err != nil {
		return err
	}

	logger = zapr.NewLogger(zl).
		WithName(component)
	if len(keyValuePairs) > 0 {
		logger = logger.WithValues(keyValuePairs)
	}
	return nil
}

// Info logs a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to
// the log line.  The key/value pairs can then be used to add additional
// variable information.  The key/value pairs should alternate string
// keys and arbitrary values.
//
// This is a package level function that is a shortcut for log.Logger().Info(...)
func Info(msg string, keysAndValues ...interface{}) {
	mtx.RLock()
	defer mtx.RUnlock()
	logger.Info(msg, keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs as context.
// It functions similarly to calling Info with the "error" named value, but may
// have unique behavior, and should be preferred for logging errors (see the
// package documentations for more information).
//
// The msg field should be used to add context to any underlying error,
// while the err field should be used to attach the actual error that
// triggered this log line, if present.
//
// This is a package level function that is a shortcut for log.Logger().Error(...)
func Error(err error, msg string, keysAndValues ...interface{}) {
	mtx.RLock()
	defer mtx.RUnlock()
	// this uses a nil err because the base zapr.Error implementation enforces zap.Error(err)
	// which converts the provided err to a standard string. Since we are using a complex err
	// which could be a pkg/errors.KVError we want to pass err as a complex object which zap
	// can then serialize according to KVError.MarshalLogObject()
	var e error
	if ee, ok := err.(errors.Error); ok {
		e = ee
	} else {
		// If err is not structured then convert to a KVError so that it is structured for consistency
		e = errors.New(err.Error())
	}
	logger.Error(nil, msg, append(keysAndValues, []interface{}{keyError, e}...)...)
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
//
// This is a package level function that is a shortcut for log.Logger().WithValues(...)
func WithValues(keysAndValues ...interface{}) logr.Logger {
	mtx.RLock()
	defer mtx.RUnlock()
	return logger.WithValues(keysAndValues...)
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly recommended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
//
// This is a package level function that is a shortcut for log.Logger().WithName(...)
func WithName(name string) logr.Logger {
	mtx.RLock()
	defer mtx.RUnlock()
	return logger.WithName(name)
}
