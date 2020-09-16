package errors

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_StoresKeysAndValues(t *testing.T) {
	err := New("hello, world", "hello", "world")
	require.EqualValues(t, "world", err.kv["hello"])
}

func TestWrap_StoresCause(t *testing.T) {
	err := Wrap(io.ErrUnexpectedEOF, "hello, world")
	require.EqualValues(t, io.ErrUnexpectedEOF, err.Unwrap())
}

func TestWrap_StoresKeysAndValues(t *testing.T) {
	err := Wrap(io.ErrUnexpectedEOF, "hello, world", "hello", "world")
	require.EqualValues(t, "world", err.kv["hello"])
}

func TestError_ReturnsMessageWhenThereIsNoCause(t *testing.T) {
	msg := "hello, world"
	err := New(msg)
	require.Equal(t, msg, err.Error())
}

func TestError_ReturnsMessageAndCauseWhenThereIsACause(t *testing.T) {
	msg := "hello, world"
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Contains(t, err.Error(), msg)
	assert.Contains(t, err.Error(), io.ErrUnexpectedEOF.Error())
}

func TestNew_SkipsMissingKeyValues(t *testing.T) {
	err := New("hello, world", "hello", "world", "missing")
	require.EqualValues(t, "world", err.kv["hello"])
	_, ok := err.kv["missing"]
	require.False(t, ok)
}

func TestUnwrap_ReturnsCause(t *testing.T) {
	msg := "hello, world"
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, Unwrap(err))
}

func TestKVError_Unwrap_ReturnsCause(t *testing.T) {
	msg := "hello, world"
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, err.Unwrap())
}

func TestIs_MatchesError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err := Wrap(base, "something broke")
	assert.True(t, Is(err, base))
}

func TestIs_MatchesDeepError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err1 := Wrap(base, "something broke")
	err2 := Wrap(err1, "something else broke")
	assert.True(t, Is(err2, base))
}

func TestAs(t *testing.T) {
	err := Wrap(&MyError{"a"}, "some error")

	var expected *MyError
	require.True(t, As(err, &expected), "expected %T to be %T", err, expected)
	require.EqualValues(t, "a", expected.Letter)
}

func TestKVError_Add(t *testing.T) {
	err := New("hello, world", "key", "value")
	err.Add("key2", "value2")
	expected := map[string]interface{}{
		"msg": "hello, world",
		"key": "value",
		"key2": "value2",
	}
	require.EqualValues(t, expected, err.KVs())
}

type MyError struct {
	Letter string
}

func (e MyError) Error() string {
		return e.Letter
}
