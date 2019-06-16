package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/grpc/metadata"
)

func Test_buildMethod(t *testing.T) {
	type args struct {
		service string
		method  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "basic", args: args{service: "foo", method: "bar"}, want: "/foo/bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMethod(tt.args.service, tt.args.method); got != tt.want {
				t.Errorf("buildMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareHeaders(t *testing.T) {
	type args struct {
		ctx context.Context
		hdr http.Header
	}

	parent := context.Background()
	hdr := make(http.Header)
	hdr.Add("Content-Type", "test")
	hdr.Add("Foo", "bar")

	want := map[string][]string{
		"Foo": []string{"bar"},
	}

	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{name: "basic", want: want, args: args{hdr: hdr, ctx: parent}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := prepareHeaders(tt.args.ctx, tt.args.hdr)
			got, _ := metadata.FromOutgoingContext(ctx)
			for k, v := range tt.want {
				if !reflect.DeepEqual(got[k], v) {
					t.Errorf("prepareHeaders() = got[%s]: %v, want[%s]: %v", k, got[k], k, tt.want[k])
					return
				}
			}
		})
	}
}
