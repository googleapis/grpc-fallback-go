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
