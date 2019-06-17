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
	"fmt"
	"io"
	"reflect"
	"testing"
)

func Test_fallbackCodec_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "basic", want: "fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fallbackCodec{}
			if got := f.Name(); got != tt.want {
				t.Errorf("fallbackCodec.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testWriter struct {
	err error
	buf *bytes.Buffer
}

func (w *testWriter) Write(d []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.buf.Write(d)
}

func Test_fallbackCodec_Unmarshal(t *testing.T) {
	type args struct {
		data []byte
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "basic", args: args{data: []byte("test"), v: &testWriter{buf: bytes.NewBuffer([]byte{})}}},
		{name: "error", wantErr: true, args: args{data: []byte{}, v: &testWriter{err: fmt.Errorf("error"), buf: bytes.NewBuffer([]byte{})}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fallbackCodec{}
			if err := f.Unmarshal(tt.args.data, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("fallbackCodec.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			} else if got := tt.args.v.(*testWriter).buf.String(); err == nil && !tt.wantErr && got != string(tt.args.data) {
				t.Errorf("fallbackCodec.Unmarshal() got = %s, want %s", got, string(tt.args.data))
			}
		})
	}
}

type testReader struct {
	err error
	buf []byte
}

func (r *testReader) WriteTo(w io.Writer) (n int64, err error) {
	if r.err != nil {
		return 0, r.err
	}
	nn, err := w.Write(r.buf)
	return int64(nn), err
}

func (r *testReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func Test_fallbackCodec_Marshal(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "basic", want: []byte("test"), args: args{v: &testReader{buf: []byte("test")}}},
		{name: "error", wantErr: true, want: []byte{}, args: args{v: &testReader{err: fmt.Errorf("error"), buf: []byte{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fallbackCodec{}
			got, err := f.Marshal(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("fallbackCodec.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fallbackCodec.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
