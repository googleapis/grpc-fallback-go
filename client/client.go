package client

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
)

const (
	ct  = "Content-Type"
	typ = "application/x-protobuf"
)

// Do is a helper for invoking grpc-fallback requests. It uses
// the default HTTP client, crafts the URL based on the address,
// fully qualified name of the gRPC Service and the Method name.
// The given request protobuf is serialized and used as the payload.
func Do(address, serv, meth string, msg proto.Message, hdr http.Header) (*http.Response, error) {
	// serialize msg payload
	b, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(b)

	// build request URL
	url := buildURL(address, serv, meth)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	// force content-type header
	if hdr == nil {
		hdr = make(http.Header)
	}
	hdr.Set(ct, typ)

	req.Header = hdr

	return http.DefaultClient.Do(req)
}

func buildURL(address, service, method string) string {
	return fmt.Sprintf("%s/$rpc/%s/%s", address, service, method)
}
