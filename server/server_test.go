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
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewServer(t *testing.T) {
	type args struct {
		port    string
		backend string
	}
	tests := []struct {
		name string
		args args
		want *FallbackServer
	}{
		{
			name: "basic, no colon",
			args: args{
				port:    "1234",
				backend: "localhost:4321",
			},
			want: &FallbackServer{
				backend: "localhost:4321",
				server: http.Server{
					Addr: ":1234",
				},
			},
		},
		{
			name: "basic, w/colon",
			args: args{
				port:    ":1234",
				backend: "localhost:4321",
			},
			want: &FallbackServer{
				backend: "localhost:4321",
				server: http.Server{
					Addr: ":1234",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.port, tt.args.backend); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testConnection struct {
	err error
}

func (c *testConnection) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	return c.err
}

type testRespWriter struct {
	buf  []byte
	code int
	h    http.Header
}

func (w *testRespWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}

	return w.h
}

func (w *testRespWriter) Write(b []byte) (int, error) {
	w.buf = b
	return 0, nil
}

func (w *testRespWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func TestFallbackServer_handler(t *testing.T) {
	type fields struct {
		cc connection
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}

	req, _ := http.NewRequest("POST", "/test", nil)
	st := status.New(codes.NotFound, "test")
	stBytes, _ := proto.Marshal(st.Proto())

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantCode int
		wantBody []byte
	}{
		{
			name: "basic",
			args: args{
				r: req,
				w: &testRespWriter{},
			},
			fields: fields{
				cc: &testConnection{},
			},
		},
		{
			name: "random error",
			args: args{
				r: req,
				w: &testRespWriter{},
			},
			fields: fields{
				cc: &testConnection{
					err: fmt.Errorf("Oops"),
				},
			},
			wantErr:  true,
			wantCode: 500,
			wantBody: []byte("Oops"),
		},
		{
			name: "gRPC error status",
			args: args{
				r: req,
				w: &testRespWriter{},
			},
			fields: fields{
				cc: &testConnection{
					err: st.Err(),
				},
			},
			wantErr:  true,
			wantCode: 404,
			wantBody: stBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FallbackServer{
				cc: tt.fields.cc,
			}
			f.handler(tt.args.w, tt.args.r)
			resp := tt.args.w.(*testRespWriter)

			if !tt.wantErr && resp.code != 0 {
				t.Errorf("handler() %s: got = %d, want = %d", tt.name, resp.code, tt.wantCode)
				return
			}

			if resp.code != tt.wantCode {
				t.Errorf("handler() %s: got = %d, want = %d", tt.name, resp.code, tt.wantCode)
			}

			if !reflect.DeepEqual(resp.buf, tt.wantBody) {
				t.Errorf("handler() %s: got = %s, want = %s", tt.name, resp.buf, tt.wantBody)
			}

			if cors := resp.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
				t.Errorf("handler() %s: got = %s, want = %s", tt.name, cors, "*")
			}
		})
	}
}

func TestFallbackServer_dial(t *testing.T) {
	tests := []struct {
		backend string
		name    string
		want    string
		wantErr bool
	}{
		{name: "basic localhost", backend: "localhost:1234", want: "localhost:1234"},
		{name: "basic non-local", backend: "test.api.dev:443", want: "test.api.dev:443"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FallbackServer{
				backend: tt.backend,
			}
			got, err := f.dial()
			if (err != nil) != tt.wantErr {
				t.Errorf("FallbackServer.dial() %s error = %v, wantErr = %v", tt.name, err, tt.wantErr)
				return
			}

			gotTarget := got.(*grpc.ClientConn).Target()
			if gotTarget != tt.want {
				t.Errorf("FallbackServer.dial() %s target: got = %v, want = %v", tt.name, gotTarget, tt.want)
			}
		})
	}
}

func TestFallbackServer_preStart(t *testing.T) {
	type fields struct {
		backend string
		server  http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "basic", fields: fields{backend: "localhost:1234", server: http.Server{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FallbackServer{
				backend: tt.fields.backend,
				server:  tt.fields.server,
			}
			f.preStart()

			if _, ok := f.server.Handler.(*mux.Router); !ok {
				t.Errorf("FallbackServer.preStart() %s handler: got = %v, want = mux.Router", tt.name, f.server.Handler)
			}
		})
	}
}

func TestFallbackServer_options(t *testing.T) {
	type fields struct {
		backend string
		server  http.Server
		cc      connection
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	hdr := make(http.Header)
	hdr.Add("access-control-allow-credentials", "true")
	hdr.Add("access-control-allow-headers", "*")
	hdr.Add("access-control-allow-methods", http.MethodPost)
	hdr.Add("access-control-allow-origin", "*")
	hdr.Add("access-control-max-age", "3600")

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantHeader http.Header
	}{
		{
			name: "basic",
			args: args{
				r: req,
				w: &testRespWriter{},
			},
			fields: fields{
				cc: &testConnection{},
			},
			wantHeader: hdr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FallbackServer{
				backend: tt.fields.backend,
				server:  tt.fields.server,
				cc:      tt.fields.cc,
			}
			f.options(tt.args.w, tt.args.r)

			resp := tt.args.w.(*testRespWriter)

			if resp.code != http.StatusOK {
				t.Errorf("handler() %s: got = %d, want = %d", tt.name, resp.code, http.StatusOK)
			}

			if !reflect.DeepEqual(resp.Header(), tt.wantHeader) {
				t.Errorf("handler() %s: got = %s, want = %s", tt.name, resp.Header(), tt.wantHeader)
			}
		})
	}
}
