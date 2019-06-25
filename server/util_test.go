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
	hdr.Add("accept-encoding", "blah")
	hdr.Add("content-length", "7")
	hdr.Add("user-agent", "whoever")
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
