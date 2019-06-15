package server

import (
	"net/http"
	"reflect"
	"testing"
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
