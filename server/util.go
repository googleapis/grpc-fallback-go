package server

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

// httpStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
//
// Taken directly from https://github.com/grpc-ecosystem/grpc-gateway/blob/e6f18d33a7b3bfa5b94f3d5fb513252184ce2d90/runtime/errors.go
func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

func buildMethod(service, method string) string {
	return fmt.Sprintf("/%s/%s", service, method)
}
