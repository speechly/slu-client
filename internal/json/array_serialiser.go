package json

import (
	"bytes"
	"encoding/json"
)

const (
	arrStart = byte('[')
	arrDelim = byte(',')
	arrEnd   = byte(']')
)

// ArraySerialiser is a JSON serialiser which wraps serialised objects into an array.
// Under the hood it uses a combination of json.Encoder and a bytes.Buffer.
type ArraySerialiser struct {
	buf *bytes.Buffer
	enc *json.Encoder
}

// NewArraySerialiser returns a new ArraySerialiser with specified buffer size.
func NewArraySerialiser(bufSize int) (*ArraySerialiser, error) {
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	if err := buf.WriteByte(arrStart); err != nil {
		return nil, err
	}

	return &ArraySerialiser{
		buf: buf,
		enc: json.NewEncoder(buf),
	}, nil
}

func (a *ArraySerialiser) Write(val interface{}) error {
	if err := a.enc.Encode(val); err != nil {
		return err
	}

	// "unwrite" an extra newline which is added by json.Encoder.
	// See https://github.com/golang/go/issues/7767.
	a.buf.Truncate(a.buf.Len() - 1)

	if err := a.buf.WriteByte(arrDelim); err != nil {
		return err
	}

	return nil
}

// Finalise finalises the serialised data by closing the array
// and returning the resulting byte representation of the data.
func (a *ArraySerialiser) Finalise() ([]byte, error) {
	defer a.buf.Reset()

	if a.buf.Len() > 1 {
		// "unwrite" last comma
		a.buf.Truncate(a.buf.Len() - 1)
	}

	if err := a.buf.WriteByte(arrEnd); err != nil {
		return nil, err
	}

	return a.buf.Bytes(), nil
}
