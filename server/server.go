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
	"log"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const fallbackPath = "/$rpc/{service:[.a-zA-Z0-9]+}/{method:[a-zA-Z]+}"

// FallbackServer is a grpc-fallback HTTP server.
type FallbackServer struct {
	backend string
	server  http.Server
	cc      connection //*grpc.ClientConn
}

// connection is an abstraction around the grpc.ClientConn
// to make testing easier.
type connection interface {
	Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
}

// NewServer creates a new grpc-fallback HTTP server on the
// given port that proxies to the given gRPC server backend.
func NewServer(port, backend string) *FallbackServer {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return &FallbackServer{
		backend: backend,
		server: http.Server{
			Addr: port,
		},
	}
}

// Start starts the grpc-fallback HTTP server listening on its port,
// and opens a connection to the gRPC backend.
func (f *FallbackServer) Start() {
	// setup connection and handler
	f.preStart()

	log.Println("Fallback server listening on port:", f.server.Addr)
	if err := f.server.ListenAndServe(); err != nil {
		log.Println("Error in fallback server while listening:", err)
	}
}

// StartBackground runs Start() in a goroutine.
func (f *FallbackServer) StartBackground() {
	go f.Start()
}

func (f *FallbackServer) preStart() {
	var err error

	// setup connection to gRPC backend
	f.cc, err = f.dial()
	if err != nil {
		log.Fatal("Error dialing gRPC backend server:", err)
	}

	// setup grpc-fallback complient router
	r := mux.NewRouter()
	r.HandleFunc(fallbackPath, f.options).
		Methods(http.MethodOptions)
	r.HandleFunc(fallbackPath, f.handler).
		Headers("Content-Type", "application/x-protobuf")
	f.server.Handler = r
}

// Shutdown turns down the grpc-fallback HTTP server.
func (f *FallbackServer) Shutdown() {
	if err := f.server.Shutdown(context.Background()); err != nil {
		log.Println("Error shutting down fallback server:", err)
	}
}

// handler is a generic HTTP handler that invokes the proper
// RPC given the grpc-fallback HTTP request.
func (f *FallbackServer) handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming grpc-fallback request:", r.RequestURI)
	v := mux.Vars(r)

	// craft service-method path
	m := buildMethod(v["service"], v["method"])

	// copy headers into out-going context metadata
	ctx := prepareHeaders(context.Background(), r.Header)

	// preemptively allow all origins in response
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// invoke the RPC, supplying the request body
	// and response writer directly
	err := f.cc.Invoke(ctx, m, r.Body, w)
	if err != nil {
		code := 500
		b := []byte(err.Error())

		// handle gRPC specific errors
		if st, ok := status.FromError(err); ok {
			code = httpStatusFromCode(st.Code())
			b, _ = proto.Marshal(st.Proto())
		}

		log.Println("Error handling request:", r.RequestURI, "-", err)
		w.WriteHeader(code)
		w.Write(b)
	}
}

// dial creates a connection with the gRPC service backend.
func (f *FallbackServer) dial() (connection, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.ForceCodec(fallbackCodec{})),
	}

	// default to basic CA, use insecure if on localhost
	auth := grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if strings.Contains(f.backend, "localhost") || strings.Contains(f.backend, "127.0.0.1") {
		auth = grpc.WithInsecure()
	}
	opts = append(opts, auth)

	return grpc.Dial(f.backend, opts...)
}

// options is a handler for the OPTIONS call that precedes CORS-enabled calls.
func (f *FallbackServer) options(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming OPTIONS for request:", r.RequestURI)
	w.Header().Add("access-control-allow-credentials", "true")
	w.Header().Add("access-control-allow-headers", "*")
	w.Header().Add("access-control-allow-methods", http.MethodPost)
	w.Header().Add("access-control-allow-origin", "*")
	w.Header().Add("access-control-max-age", "3600")
	w.WriteHeader(http.StatusOK)
}
