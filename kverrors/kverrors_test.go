package kverrors_test

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/ViaQ/logerr/internal/kv"
	"github.com/ViaQ/logerr/kverrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_StoresKeysAndValues(t *testing.T) {
	err := kverrors.New(t.Name(), "hello", "world")
	require.EqualValues(t, "world", kverrors.KVs(err)["hello"])
}

func TestWrap_StoresCause(t *testing.T) {
	err := kverrors.Wrap(io.ErrUnexpectedEOF, t.Name())
	require.EqualValues(t, io.ErrUnexpectedEOF, errors.Unwrap(err))
}

func TestWrap_StoresKeysAndValues(t *testing.T) {
	err := kverrors.Wrap(io.ErrUnexpectedEOF, t.Name(), "hello", "world")
	require.EqualValues(t, "world", kverrors.KVs(err)["hello"])
}

func TestWrap_ReturnsNilWhenErrIsNil(t *testing.T) {
	assert.Nil(t, kverrors.Wrap(nil, "some error"))
}

func TestError_ReturnsMessageWhenThereIsNoCause(t *testing.T) {
	msg := t.Name()
	err := kverrors.New(msg)
	require.Equal(t, msg, err.Error())
}

func TestError_ReturnsMessageAndCauseWhenThereIsACause(t *testing.T) {
	msg := t.Name()
	err := kverrors.Wrap(io.ErrUnexpectedEOF, msg)
	assert.Contains(t, err.Error(), msg)
	assert.Contains(t, err.Error(), io.ErrUnexpectedEOF.Error())
}

func TestNew_SkipsMissingKeyValues(t *testing.T) {
	err := kverrors.New(t.Name(), "hello", "world", "missing")
	require.EqualValues(t, "world", kverrors.KVs(err)["hello"])
	_, ok := kverrors.KVs(err)["missing"]
	require.False(t, ok)
}

func TestUnwrap_ReturnsCause(t *testing.T) {
	msg := t.Name()
	err := kverrors.Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, kverrors.Unwrap(err))
}

func TestKVError_Unwrap_ReturnsCause(t *testing.T) {
	msg := t.Name()
	err := kverrors.Wrap(io.ErrUnexpectedEOF, msg)
	assert.Equal(t, io.ErrUnexpectedEOF, errors.Unwrap(err))
}

func TestIs_MatchesError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err := kverrors.Wrap(base, "something broke")
	assert.True(t, errors.Is(err, base))
}

func TestIs_MatchesDeepError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	err1 := kverrors.Wrap(base, "something broke")
	err2 := kverrors.Wrap(err1, "something else broke")
	assert.True(t, errors.Is(err2, base))
}

func TestAs(t *testing.T) {
	err := kverrors.Wrap(&MyError{"a"}, "some error")

	var expected *MyError
	require.True(t, errors.As(err, &expected), "expected %T to be %T", err, expected)
	require.EqualValues(t, "a", expected.Letter)
}

func TestKVError_Add(t *testing.T) {
	t.Run("KVerror", func(t *testing.T) {
		err := kverrors.New(t.Name(), "key", "value")
		err = kverrors.Add(err, "key2", "value2")
		expected := map[string]interface{}{
			kverrors.MessageKey: t.Name(),
			"key":               "value",
			"key2":              "value2",
		}
		require.EqualValues(t, expected, kverrors.KVs(err))
	})

	t.Run("pkg/error", func(t *testing.T) {
		key := "test"
		err := kverrors.Add(errors.New(t.Name()), key, t.Name())
		expected := map[string]interface{}{
			kverrors.MessageKey: t.Name(),
			key:                 t.Name(),
		}
		require.EqualValues(t, expected, kverrors.KVs(err))
	})
}

func TestRoot_FindsTheRootError(t *testing.T) {
	root := io.ErrUnexpectedEOF
	err := kverrors.Wrap(kverrors.Wrap(kverrors.Wrap(root, "e1"), "e2"), "e3")
	require.Equal(t, root, kverrors.Root(err))
}

func TestKVError_Ctx(t *testing.T) {
	errCtx := kverrors.NewContext("k1", "v1", "k2", "v2")
	err := kverrors.New("failed something or other")
	err = kverrors.AddCtx(err, errCtx)
	for k, v := range kv.ToMap(errCtx) {
		require.Contains(t, err, k)
		require.EqualValues(t, v, kverrors.KVs(err)[k])
	}
}

func TestContext_New_WrapsAllKeysAndValues(t *testing.T) {
	ctx := kverrors.NewContext("foo", "bar")
	err := ctx.New("a broken mess", "baz", "foo")

	expected := map[string]interface{}{
		kverrors.MessageKey: "a broken mess",
		"foo":               "bar",
		"baz":               "foo",
	}

	require.EqualValues(t, expected, kverrors.KVs(err))
}

func TestContext_Wrap_WrapsAllKeysAndValues(t *testing.T) {
	ctx := kverrors.NewContext("foo", "bar")
	err := ctx.Wrap(io.ErrNoProgress, "a broken mess", "baz", "foo")

	expected := map[string]interface{}{
		kverrors.MessageKey: "a broken mess",
		"foo":               "bar",
		"baz":               "foo",
	}

	errkvs := kverrors.KVs(err)
	for k, v := range expected {
		require.EqualValues(t, v, errkvs[k])
	}
}

func TestNewCtx(t *testing.T) {
	key := "test"
	ctx := kverrors.NewContext(key, t.Name())
	err := kverrors.NewCtx(t.Name(), ctx)

	kvs := kverrors.KVs(err)
	actual, ok := kvs[key]
	require.True(t, ok, "key should exist")
	require.Equal(t, t.Name(), actual)
}

func TestKVs(t *testing.T) {
	t.Run("KVError", func(t *testing.T) {
		key := "test"
		err := kverrors.New(t.Name(), key, t.Name())
		kvs := kverrors.KVs(err)

		require.EqualValues(t, map[string]interface{}{
			key:                 t.Name(),
			kverrors.MessageKey: t.Name(),
		},
			kvs,
		)
	})

	t.Run("pkg/error", func(t *testing.T) {
		kvs := kverrors.KVs(errors.New(t.Name()))
		require.Nil(t, kvs)
	})
}

func TestKVSlice(t *testing.T) {
	t.Run("KVError", func(t *testing.T) {
		key := "test"
		err := kverrors.New(t.Name(), key, t.Name())

		kvs := kverrors.KVSlice(err)

		require.Contains(t, kvs, key)
		require.Contains(t, kvs, t.Name())
	})

	t.Run("pkg/error", func(t *testing.T) {
		slice := kverrors.KVSlice(errors.New(t.Name()))
		assert.Empty(t, slice)
	})
}

func TestMessage(t *testing.T) {
	t.Run("KVError", func(t *testing.T) {
		err := kverrors.New(t.Name())
		require.Equal(t, t.Name(), kverrors.Message(err))
	})

	t.Run("pkg/error", func(t *testing.T) {
		err := errors.New(t.Name())
		require.Equal(t, t.Name(), kverrors.Message(err))
	})
}

func TestKVError_MarshalJSON(t *testing.T) {
	key := "test"
	kverr := kverrors.New(t.Name(), key, t.Name()).(*kverrors.KVError)
	b, err := kverr.MarshalJSON()
	require.NoError(t, err)

	actual := string(b)
	b, err = json.Marshal(map[string]interface{}{
		key:                 t.Name(),
		kverrors.MessageKey: t.Name(),
	})
	require.NoError(t, err)

	expected := string(b)

	require.JSONEq(t, expected, actual)
}

type MyError struct {
	Letter string
}

func (e MyError) Error() string {
	return e.Letter
}
