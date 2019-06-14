package server

import (
	"bytes"
	"io"
)

// fallbackCodec consumes data read from an io.Reader and
// writes data into an io.Writer. This doesn't mean that the
// data is streamed, it merely abstracts handling of the data
// away from the RPC invocation site.
type fallbackCodec struct{}

func (fallbackCodec) Marshal(v interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	r := v.(io.Reader)

	_, err := io.Copy(buf, r)
	return buf.Bytes(), err
}

func (fallbackCodec) Unmarshal(data []byte, v interface{}) error {
	w := v.(io.Writer)
	_, err := w.Write(data)
	return err
}

func (fallbackCodec) Name() string {
	return "fallback"
}
