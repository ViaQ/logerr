package errors

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_ Error                   = &KVError{}
	_ zapcore.ObjectMarshaler = &KVError{}
)

const (
	keyMessage string = "msg"
	keyCause   string = "cause"
)

// Error is a structured error
type Error interface {
	Unwrap() error
	Message() string
	Error() string
}

// New creates a new KVError with keys and values
func New(msg string, keysAndValues ...interface{}) *KVError {
	return &KVError{
		kv: appendMap(map[string]interface{}{
			keyMessage: msg,
		}, toMap(keysAndValues...)),
	}
}

// Wrap wraps an error as a new error with keys and values
func Wrap(err error, msg string, keysAndValues ...interface{}) *KVError {
	if err == nil {
		return nil
	}
	e := New(msg, append(keysAndValues, []interface{}{keyCause, err}...)...)
	return e
}

// KVError is an error that contains structured keys and values
type KVError struct {
	kv map[string]interface{}
}

// KVs returns the key/value pairs associated with this error
func (e *KVError) KVs() map[string]interface{} {
	return e.kv
}

// Unwrap returns the error that caused this error
func (e *KVError) Unwrap() error {
	if cause, ok := e.kv[keyCause]; ok {
		e, _ := cause.(error)
		// if ok is false then e will be empty anyway so no need to check if ok
		return e
	}
	return nil
}

func (e *KVError) Error() string {
	base := e.Unwrap()
	if base != nil {
		return fmt.Sprintf("%s: %s", e.Message(), base.Error())
	}
	return e.Message()
}

func (e *KVError) Message() string {
	if msg, ok := e.kv[keyMessage]; ok {
		return fmt.Sprint(msg)
	}
	return ""
}

// Add adds key/value pairs to an error and returns the error
// WARNING: The original error is modified with this operation
func (e *KVError) Add(keyValuePairs ...interface{}) *KVError {
	for k, v := range toMap(keyValuePairs...) {
		e.kv[k] = v
	}
	return e
}

func (e *KVError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.kv)
}

func (e *KVError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range e.kv {
		zap.Any(k, v).AddTo(enc)
	}
	return nil
}

// Unwrap provides compatibility with the standard errors package
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Root unwraps the error until it reaches the root error
func Root(err error) error {
	root := err
	for next := Unwrap(root); next != nil; next = Unwrap(root) {
		root = next
	}
	return root
}

// Is provides compatibility with the standard errors package
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As provides compatibility with the standard errors package
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func toMap(keysAndValues ...interface{}) map[string]interface{} {
	kve := map[string]interface{}{}

	for i, kv := range keysAndValues {
		if i%2 == 1 {
			continue
		}
		if len(keysAndValues) <= i+1 {
			continue
		}
		kve[fmt.Sprintf("%s", kv)] = keysAndValues[i+1]
	}
	return kve
}

func appendMap(a, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		a[k] = v
	}
	return a
}
