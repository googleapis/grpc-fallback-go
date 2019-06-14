# grpc-fallback-go

## Server Usage Example

### CLI Usage Example

```sh
> fallback-proxy -address "localhost:7469"
2019/06/13 18:35:01 Fallback server listening on port: :1337

```

### In-process Usage Example

```go
// setup listener & server
lis, _ := net.Listen("tcp", port)
s := grpc.NewServer(opts...)

// Register Services to the server.
// ...

// Create a new grpc-fallback server on port 1337
// for gRPC server listening on "port"
fb := fallback.NewServer(":1337", port)
fb.StartBackground()
defer fb.Shutdown()

// Start gRPC server.
s.Serve(lis)
```

### Docker Usage Example

```sh
> docker build -t gcr.io/gapic-images/fallback-proxy .
> docker run \
  --net=host \
  gcr.io/gapic-images/fallback-proxy \
  -address=localhost:7469 
2019/06/13 18:35:01 Fallback server listening on port: :1337
```

## Client Usage Example

```go
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	showpb "github.com/googleapis/gapic-showcase/server/genproto"
	"github.com/googleapis/grpc-fallback-go/client"
)

func main() {
	req := &showpb.EchoRequest{Response: &show.EchoRequest_Content{"Hello, World!"}}
	r, err := client.Do("http://localhost:1337", "google.showcase.v1beta1.Echo", "Echo", req, nil)
	if err != nil {
		panic(err)
	}

	b, _ := ioutil.ReadAll(r.Body)
	res := &showpb.EchoResponse{}
	proto.Unmarshal(b, res)

	fmt.Println(res)
}
```