package log_test

import (
	"io"

	"github.com/ViaQ/logerr/pkg/errors"
	"github.com/ViaQ/logerr/pkg/log"
)

func Setup() {
	if err := log.Init("example", []log.Option{log.WithNoTimestamp()}); err != nil {
		panic(err)
	}
}

func ExampleInfo_nokvs() {
	Setup()
	log.Info("hello, world")
	// Output: {"level":"info","time":"","logger":"example","msg":"hello, world"}

}

func ExampleInfo_kvs() {
	Setup()
	log.Info("hello, world", "city", "Athens")
	// Output: {"level":"info","time":"","logger":"example","msg":"hello, world","city":"Athens"}
}

func ExampleError_nokvs() {
	Setup()
	log.Error(io.EOF, "hello, world")
	// Output: {"level":"error","time":"","logger":"example","msg":"hello, world","cause":{"msg":"EOF"}}

}

func ExampleError_kvs() {
	Setup()
	log.Error(io.EOF, "hello, world", "city", "Athens")
	// Output: {"level":"error","time":"","logger":"example","msg":"hello, world","city":"Athens","cause":{"msg":"EOF"}}
}

func ExampleError_pkg_error_kvs() {
	Setup()
	err := errors.New("busted", "city", "Athens")
	log.Error(err, "hello, world")
	// Output: {"level":"error","time":"","logger":"example","msg":"hello, world","cause":{"msg":"busted","city":"Athens"}}
}

func ExampleError_pkg_error_nested_kvs() {
	Setup()
	err1 := errors.New("err1", "city", "Athens")
	err2 := errors.Wrap(err1, "err2", "year", "500 BCE")
	log.Error(err2, "hello, world")
	// Output: {"level":"error","time":"","logger":"example","msg":"hello, world","cause":{"msg":"err2","year":"500 BCE","cause":{"msg":"err1","city":"Athens"}}}
}
