# grpc-fallback-go

`grpc-fallback-go` contains a [gRPC Fallback][]-compliant proxy server and HTTP client, both implemented in Go.

## Description

`grpc-fallback-go` consists of a [gRPC Fallback][] proxy server package, a proxy binary, and a HTTP client. 

*Note: The gRPC Fallback protocol will be referred to as just "fallback" for the remainder of this document.*

The proxy server is designed to front an arbitrary gRPC service. It accepts fallback requests, forwards a corresponding gRPC request to the gRPC backend, and translates the response to a HTTP response before returning it to the caller.

The proxy server can be used from a Go program, as a binary, or as a Docker image. Usage examples can be found in the [Server Usage Example](#server-usage-example) section below.

The Go client is basic and is meant for testing and simple use  cases. Usage examples can be found in the [Client Usage Examples](#client-usage-examples) section below.

## Installation

Check the Releases tab for pre-built binaries in a variety of architectures.

### Via Go Tooling

Proxy server binary:
```sh
> go get github.com/googleapis/grpc-fallback-go/cmd/fallback-proxy
```

Proxy server package:
```sh
> go get github.com/googleapis/grpc-fallback-go/server
```

HTTP Client
```sh
> go get github.com/googleapis/grpc-fallback-go/client
```

## Server Usage Example

### CLI Usage Example

```sh
> fallback-proxy -address "localhost:7469"
2019/06/13 18:35:01 Fallback server listening on port: :1337

```

### In-process w/gRPC Backend Usage Example

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

## Client Usage Examples

### Basic Client Usage Example

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

### Authenticated Client Usage Example

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

## Testing

If developing the `grpc-fallback-go` project, run tests via `make test`.

##  Releasing

Follow these steps to make a release:

1. Update the `VERSION` in [release.sh](/release.sh)
2. Open a PR with the **only** `VERSION` bump (notice the prepended `v` for the tag name)
```sh
git add release.sh
git commit -m "release v$VERSION"
```
3. Once version bump PR is merged, create and push the version tag (notice the prepended `v` for the tag name)
```sh
git tag v$VERSION && git push --tags
```
4. Build release assets (Note: must have Docker running for image building)
```sh
make release
```
5. Publish release with `VERSION` tag.   
    a. include the `grpc-fallback-go-*.tar.gz ` release assets

    b. push the `latest` and `VERSION` tagged Docker images (requires credentials)
  ```sh
  gcloud auth configure-docker
  gcloud docker -- push gcr.io/gapic-images/fallback-proxy
  gcloud docker -- push gcr.io/gapic-images/fallback-proxy:$VERSION
  ```
6. (optional) Clean up!
```sh
make clean
```

## Contributing

If you are planning to contribute to `grpc-fallback-go`, please read the [CONTRIBUTING.md][] located in the root directory, as well as the [Testing](#testing) and [Releasing](#releasing) sections.

## Disclaimer

This is not an official Google product.

[contributing.md]: /CONTRIBUTING.MD
[grpc fallback]: https://googleapis.github.io/HowToRPC#grpc-fallback-experimental