package server

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// FallbackServer is a grpc-fallback HTTP server.
type FallbackServer struct {
	backend string
	server  http.Server
	cc      *grpc.ClientConn
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
	var err error

	// setup connection to gRPC backend
	f.cc, err = f.dial()
	if err != nil {
		log.Fatal("Error dialing gRPC backend server:", err)
	}

	// setup grpc-fallback complient router
	r := mux.NewRouter()
	r.HandleFunc("/$rpc/{service:[.a-zA-Z0-9]+}/{method:[a-zA-Z]+}", f.handler).Headers("Content-Type", "application/x-protobuf")
	f.server.Handler = r

	log.Println("Fallback server listening on port:", f.server.Addr)
	err = f.server.ListenAndServe()
	if err != nil {
		log.Println("Error in fallback server while listening:", err)
	}
}

// StartBackground runs Start() in a goroutine.
func (f *FallbackServer) StartBackground() {
	go f.Start()
}

// Shutdown turns down the grpc-fallback HTTP server.
func (f *FallbackServer) Shutdown() {
	err := f.server.Shutdown(context.Background())
	if err != nil {
		log.Println("Error shutting down fallback server:", err)
	}
}

// handler is a generic HTTP handler that invokes the proper
// RPC given the grpc-fallback HTTP request.
func (f *FallbackServer) handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming grpc-fallback request:", r.RequestURI)
	v := mux.Vars(r)
	service := v["service"]
	method := v["method"]

	// invoke the desired RPC
	err := f.invoke(r.Body, w, service, method)
	if err != nil {
		code := 500

		// handle gRPC specific errors
		if st, ok := status.FromError(err); ok {
			code = httpStatusFromCode(st.Code())
		}

		w.WriteHeader(code)
		w.Write([]byte(err.Error()))
	}
}

// invoke buffers the request payload, invokes the RPC, and
// writes a succesful response to the given io.Writer.
func (f *FallbackServer) invoke(in io.Reader, out io.Writer, service, method string) error {
	res := &[]byte{}
	req, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	err = f.cc.Invoke(context.Background(), buildMethod(service, method), req, res)
	if err != nil {
		return err
	}

	_, err = out.Write(*res)

	return err
}

// dial creates a connection with the gRPC service backend.
func (f *FallbackServer) dial() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.ForceCodec(rawCodec{})),
	}

	// default to basic CA, use insecure if on localhost
	auth := grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if strings.Contains(f.backend, "localhost") || strings.Contains(f.backend, "127.0.0.1") {
		auth = grpc.WithInsecure()
	}
	opts = append(opts, auth)

	return grpc.Dial(f.backend, opts...)
}
