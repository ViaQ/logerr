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

func TestWrap_ReturnsNilWhenErrIsNil(t *testing.T) {
	assert.Nil(t, Wrap(nil, "some error"))
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
	err := New("hello, world", "key", "value").Add("key2", "value2")
	expected := map[string]interface{}{
		"msg":  "hello, world",
		"key":  "value",
		"key2": "value2",
	}
	require.EqualValues(t, expected, err.KVs())
}

func TestKVError_Wrap(t *testing.T) {
	base := New("a breaking change", "key1", "value1")
	err := base.Wrap(io.ErrShortWrite, "key2", "value2")
	assert.True(t, Is(err, io.ErrShortWrite), "expected err to be io.ErrShortWrite")
	expected := map[string]interface{}{
		"msg":   "a breaking change",
		"key1":  "value1",
		"key2":  "value2",
		"cause": io.ErrShortWrite,
	}
	assert.EqualValues(t, expected, err.KVs())
}

func TestRoot_FindsTheRootError(t *testing.T) {
	root := io.ErrUnexpectedEOF
	err := Wrap(Wrap(Wrap(root, "e1"), "e2"), "e3")
	require.Equal(t, root, Root(err))
}

func TestKVError_Ctx(t *testing.T) {
	errCtx := NewContext("k1", "v1", "k2", "v2")
	err := New("failed something or other").Ctx(errCtx)
	for k, v := range toMap(errCtx) {
		require.Contains(t, err, k)
		require.EqualValues(t, v, err.kv[k])
	}

}

type MyError struct {
	Letter string
}

func (e MyError) Error() string {
	return e.Letter
}
