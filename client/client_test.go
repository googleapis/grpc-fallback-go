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

package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/mux"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testServiceName     = "my.test.Service"
	testMethodNameOK    = "ReturnOK"
	testMethodNameError = "ReturnError"
)

func TestDo(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/$rpc/%s/%s", testServiceName, testMethodNameOK), handleOK).Headers("Content-Type", "application/x-protobuf")
	r.HandleFunc(fmt.Sprintf("/$rpc/%s/%s", testServiceName, testMethodNameError), handleError).Headers("Content-Type", "application/x-protobuf")

	ts := httptest.NewServer(r)
	defer ts.Close()

	type args struct {
		address string
		serv    string
		meth    string
		req     proto.Message
		res     proto.Message
		hdr     http.Header
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "OK", args: args{address: ts.URL, serv: testServiceName, meth: testMethodNameOK, req: &empty.Empty{}, res: &empty.Empty{}, hdr: nil}},
		{name: "Error", wantErr: true, args: args{address: ts.URL, serv: testServiceName, meth: testMethodNameError, req: &empty.Empty{}, res: &empty.Empty{}, hdr: nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Do(tt.args.address, tt.args.serv, tt.args.meth, tt.args.req, tt.args.res, tt.args.hdr); (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr {
				if _, ok := status.FromError(err); !ok {
					t.Errorf("Do error = %v, wanted a gRPC error Status", err)
				}
			}
		})
	}
}

func handleOK(w http.ResponseWriter, r *http.Request) {
	e := &empty.Empty{}

	body, _ := ioutil.ReadAll(r.Body)
	err := proto.Unmarshal(body, e)
	if err != nil {
		handleError(w, r)
		return
	}

	// echo it back
	b, _ := proto.Marshal(e)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func handleError(w http.ResponseWriter, r *http.Request) {
	s := status.New(codes.InvalidArgument, "bad request")
	b, _ := proto.Marshal(s.Proto())

	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

func Test_buildURL(t *testing.T) {
	type args struct {
		address string
		service string
		method  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "basic", args: args{address: "http://localhost:1234", service: "my.v1.Service", method: "MyMethod"}, want: "http://localhost:1234/$rpc/my.v1.Service/MyMethod"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildURL(tt.args.address, tt.args.service, tt.args.method); got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
