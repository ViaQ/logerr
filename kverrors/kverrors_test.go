package kverrors

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_StoresKeysAndValues(t *testing.T) {
	err := New(t.Name(), "hello", "world")
	require.EqualValues(t, "world", KVs(err)["hello"])
}

func TestWrap_StoresCause(t *testing.T) {
	err := Wrap(io.ErrUnexpectedEOF, t.Name())
	require.EqualValues(t, io.ErrUnexpectedEOF, errors.Unwrap(err))
}

func TestWrap_StoresKeysAndValues(t *testing.T) {
	err := Wrap(io.ErrUnexpectedEOF, t.Name(), "hello", "world")
	require.EqualValues(t, "world", KVs(err)["hello"])
}

func TestWrap_ReturnsNilWhenErrIsNil(t *testing.T) {
	assert.Nil(t, Wrap(nil, "some error"))
}

func TestError_ReturnsMessageWhenThereIsNoCause(t *testing.T) {
	msg := t.Name()
	err := New(msg)
	require.Equal(t, msg, err.Error())
}

func TestError_ReturnsMessageAndCauseWhenThereIsACause(t *testing.T) {
	msg := t.Name()
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Contains(t, err.Error(), msg)
	assert.Contains(t, err.Error(), io.ErrUnexpectedEOF.Error())
}

func TestNew_SkipsMissingKeyValues(t *testing.T) {
	err := New(t.Name(), "hello", "world", "missing")
	require.EqualValues(t, "world", KVs(err)["hello"])
	_, ok := KVs(err)["missing"]
	require.False(t, ok)
}

func TestUnwrap_ReturnsCause(t *testing.T) {
	msg := t.Name()
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, Unwrap(err))
}

func TestKVError_Unwrap_ReturnsCause(t *testing.T) {
	msg := t.Name()
	err := Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, errors.Unwrap(err))
}

func TestIs_MatchesError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err := Wrap(base, "something broke")
	assert.True(t, errors.Is(err, base))
}

func TestIs_MatchesDeepError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err1 := Wrap(base, "something broke")
	err2 := Wrap(err1, "something else broke")
	assert.True(t, errors.Is(err2, base))
}

func TestAs(t *testing.T) {
	err := Wrap(&MyError{"a"}, "some error")

	var expected *MyError
	require.True(t, errors.As(err, &expected), "expected %T to be %T", err, expected)
	require.EqualValues(t, "a", expected.Letter)
}

func TestKVError_Add(t *testing.T) {
	err := New(t.Name(), "key", "value")
	err = Add(err, "key2", "value2")
	expected := map[string]interface{}{
		"msg":  t.Name(),
		"key":  "value",
		"key2": "value2",
	}
	require.EqualValues(t, expected, KVs(err))
}

func TestRoot_FindsTheRootError(t *testing.T) {
	root := io.ErrUnexpectedEOF
	err := Wrap(Wrap(Wrap(root, "e1"), "e2"), "e3")
	require.Equal(t, root, Root(err))
}

func TestKVError_Ctx(t *testing.T) {
	errCtx := NewContext("k1", "v1", "k2", "v2")
	err := New("failed something or other")
	err = AddCtx(err, errCtx)
	for k, v := range toMap(errCtx) {
		require.Contains(t, err, k)
		require.EqualValues(t, v, KVs(err)[k])
	}

}

func TestContext_New_WrapsAllKeysAndValues(t *testing.T) {
	ctx := NewContext("foo", "bar")
	err := ctx.New("a broken mess", "baz", "foo")

	expected := map[string]interface{}{
		"msg": "a broken mess",
		"foo": "bar",
		"baz": "foo",
	}

	require.EqualValues(t, expected, KVs(err))
}

func TestContext_Wrap_WrapsAllKeysAndValues(t *testing.T) {
	ctx := NewContext("foo", "bar")
	err := ctx.Wrap(io.ErrNoProgress, "a broken mess", "baz", "foo")

	expected := map[string]interface{}{
		"msg": "a broken mess",
		"foo": "bar",
		"baz": "foo",
	}

	errkvs := KVs(err)
	for k, v := range expected {
		require.EqualValues(t, v, errkvs[k])
	}
}

type MyError struct {
	Letter string
}

func (e MyError) Error() string {
	return e.Letter
}
