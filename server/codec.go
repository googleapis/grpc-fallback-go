// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
