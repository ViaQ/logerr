package errors

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_StoresKeysAndValues(t *testing.T) {
	err := New("hello, world", "hello", "world")
	require.EqualValues(t, []interface{}{"hello", "world"}, err.kvs)
}

func TestWrap_StoresCause(t *testing.T) {
	err := Wrap(io.EOF, "hello, world")
	require.EqualValues(t, io.EOF, err.Cause)
}

func TestWrap_StoresKeysAndValues(t *testing.T) {
	err := Wrap(io.EOF, "hello, world", "hello", "world")
	require.EqualValues(t, []interface{}{"hello", "world"}, err.kvs)
}

func TestError_ReturnsMessageWhenThereIsNoCause(t *testing.T) {
	msg := "hello, world"
	err := New(msg)
	require.Equal(t, msg, err.Error())
}

func TestError_ReturnsMessageAndCauseWhenThereIsACause(t *testing.T) {
	msg := "hello, world"
	err := Wrap(io.EOF, msg)
	assert.Contains(t, err.Error(), msg)
	assert.Contains(t, err.Error(), io.EOF.Error())
}

func TestUnwrap_ReturnsCause(t *testing.T) {
	msg := "hello, world"
	err := Wrap(io.EOF, msg)
	assert.Equal(t, io.EOF, err.Unwrap())
}

func TestKVs_ReturnsAllKeysAndValues(t *testing.T) {
	kvs := []interface{}{"name", "Diogenes"}
	err := New("hello, world", kvs...)
	assert.EqualValues(t, kvs, err.KVs())
}

func TestKVs_ReturnsNilWhenNoKVs(t *testing.T) {
	err := New("hello, world")
	assert.Nil(t, err.KVs())
}
