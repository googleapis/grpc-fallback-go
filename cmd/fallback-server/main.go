package main

import (
	"flag"
	"log"

	fb "github.com/googleapis/grpc-fallback-go/server"
)

var (
	port, addr string
)

func init() {
	flag.StringVar(&port, "port", ":1337", "port for the fallback server to listen on")
	flag.StringVar(&addr, "address", "", "address of the gRPC service backend")

	flag.Parse()

	if addr == "" {
		log.Fatalln("missing required flag -address")
	}
}

func main() {
	fb.NewServer(port, addr).Start()
}
