package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/grpc/status"

	statuspb "google.golang.org/genproto/googleapis/rpc/status"

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
// A successful response is deserialized into the given response proto.
// A non-2xx response status is returned as an error containing the
// underlying gRPC status.
func Do(address, serv, meth string, req, res proto.Message, hdr http.Header) error {
	// serialize msg payload
	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(b)

	// build request URL
	url := buildURL(address, serv, meth)

	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	// force content-type header
	if hdr == nil {
		hdr = make(http.Header)
	}
	hdr.Set(ct, typ)

	request.Header = hdr

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		stpb := &statuspb.Status{}
		if err := proto.Unmarshal(resBody, stpb); err != nil {
			return err
		}

		st := status.FromProto(stpb)

		return st.Err()
	}

	return proto.Unmarshal(resBody, res)
}

func buildURL(address, service, method string) string {
	return fmt.Sprintf("%s/$rpc/%s/%s", address, service, method)
}
