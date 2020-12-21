package log_test

import (
	"io"
)

// fakeEncoder is an encoder that you can use for testing
type fakeEncoder struct {
	EncodeFunc func(w io.Writer, entry interface{}) error
}

func (f fakeEncoder) Encode(w io.Writer, entry interface{}) error {
	if f.EncodeFunc != nil {
		return f.EncodeFunc(w, entry)
	}
	return nil
}
