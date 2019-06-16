package server

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
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
}

func (w *testRespWriter) Header() http.Header {
	return nil
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
		})
	}
}
