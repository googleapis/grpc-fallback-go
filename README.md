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
fb := fallback.NewServer(":1337", "localhost"+port)
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

	showpb "github.com/googleapis/gapic-showcase/server/genproto"
	"github.com/googleapis/grpc-fallback-go/client"
)

func main() {
	req := &showpb.EchoRequest{Response: &showpb.EchoRequest_Content{"testing"}}
	res := &showpb.EchoResponse{}

	err := client.Do("http://localhost:1337", "google.showcase.v1beta1.Echo", "Echo", req, res, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}

```

## Client w/auth Usage Example

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/googleapis/grpc-fallback-go/client"

	lpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"golang.org/x/oauth2/google"
)

func main() {
	hdr := make(http.Header)
	ts, _ := google.DefaultTokenSource(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
	t, _ := ts.Token()
	hdr.Add("Authorization", "Bearer "+t.AccessToken)

	req := &lpb.AnalyzeSentimentRequest{
		Document: &lpb.Document{
			Type:   lpb.Document_PLAIN_TEXT,
			Source: &lpb.Document_Content{"hello, world"},
		},
	}
	res := &lpb.AnalyzeSentimentResponse{}
	

	// fallback-proxy is pointing at -address language.googleapis.com:443
	err := client.Do("http://localhost:1337", "google.cloud.language.v1.LanguageService", "AnalyzeSentiment", req, res, hdr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
}

```