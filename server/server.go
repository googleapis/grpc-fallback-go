package server

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	gdyn "github.com/jhump/protoreflect/dynamic/grpcdynamic"
	gref "github.com/jhump/protoreflect/grpcreflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

// FallbackServer is a grpc-fallback HTTP server
type FallbackServer struct {
	addr     string
	server   http.Server
	services map[string]desc.ServiceDescriptor
	stub     *gdyn.Stub
	remote   bool
}

// NewServer creates a new grpc-fallback HTTP server on the
// given port. On start, it will configure itself via gRPC
// reflect against the gRPC server located at addr.
func NewServer(port, addr string) *FallbackServer {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return &FallbackServer{
		addr:   addr,
		remote: true,
		server: http.Server{
			Addr: port,
		},
		services: make(map[string]desc.ServiceDescriptor),
	}
}

// NewServerInProcess creates a new grpc-fallback HTTP server on the
// given port. It acts as a reverse-proxy for the given in-process gRPC
// Server located on localhost:{grpcPort}.
func NewServerInProcess(port, grpcPort string, gServer *grpc.Server) *FallbackServer {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	if !strings.HasPrefix(grpcPort, ":") {
		grpcPort = ":" + grpcPort
	}

	sds, err := gref.LoadServiceDescriptors(gServer)
	if err != nil {
		log.Fatal("Error initializing service descriptor references:", err)
	}

	services := make(map[string]desc.ServiceDescriptor)
	for _, d := range sds {
		services[d.GetFullyQualifiedName()] = *d
	}

	return &FallbackServer{
		addr: "localhost" + grpcPort,
		server: http.Server{
			Addr: port,
		},
		services: services,
	}
}

// Start starts the grpc-fallback HTTP server listening on its port.
// A connection is open to the gRPC backend and, if remote, the
// service list is reflected.
func (f *FallbackServer) Start() {
	// setup connection to gRPC backend
	cc, err := f.dial()
	if err != nil {
		log.Fatal("Error dialing gRPC backend server:", err)
	}

	// init dynamic gRPC client
	s := gdyn.NewStub(cc)
	f.stub = &s

	// reflect gRPC Service list
	if f.remote {
		if err := f.reflectServices(cc); err != nil {
			log.Fatal("Error configuring against reflection server:", err)
		}
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

// StartBackground runs Start() in a goroutine
func (f *FallbackServer) StartBackground() {
	go f.Start()
}

// Shutdown turns down the grpc-fallback HTTP server
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

	sd, ok := f.services[service]
	if !ok || sd.FindMethodByName(method) == nil {
		w.WriteHeader(404)
		return
	}

	// invoke the desired RPC
	err := f.invoke(r.Body, w, sd.FindMethodByName(method))
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

// invoke deserializes the request payload, invokes the RPC, and
// writes a succesful response to the given io.Writer.
func (f *FallbackServer) invoke(in io.Reader, out io.Writer, md *desc.MethodDescriptor) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	inDyn := dynamic.NewMessage(md.GetInputType())
	err = inDyn.Unmarshal(data)
	if err != nil {
		return err
	}

	res, err := f.stub.InvokeRpc(context.Background(), md, inDyn)
	if err != nil {
		return err
	}

	d, err := proto.Marshal(res)
	if err != nil {
		return err
	}

	out.Write(d)

	return nil
}

// dial creates a connection with the gRPC service
// backend.
func (f *FallbackServer) dial() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}

	// default to basic CA, use insecure if on localhost
	auth := grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if strings.Contains(f.addr, "localhost") || strings.Contains(f.addr, "127.0.0.1") {
		auth = grpc.WithInsecure()
	}
	opts = append(opts, auth)

	cc, err := grpc.Dial(f.addr, opts...)
	if err != nil {
		log.Printf("Error dialing gRPC backend at %s: %v\n", f.addr, err)
		return nil, err
	}

	return cc, nil
}

// reflectServices uses a gRPC reflection client to list
// the services avilable on the gRPC backend and configure
// the grpc-fallback proxy.
func (f *FallbackServer) reflectServices(cc *grpc.ClientConn) error {
	// setup reflection client, cancel it when done
	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	refStub := rpb.NewServerReflectionClient(cc)
	refc := gref.NewClient(ctx, refStub)

	servs, err := refc.ListServices()
	if err != nil {
		return err
	}

	// gather service descriptors
	for _, s := range servs {
		sdesc, err := refc.ResolveService(s)
		if err != nil {
			return err
		}

		f.services[s] = *sdesc
		log.Println("Adding service:", s)
	}

	return nil
}
