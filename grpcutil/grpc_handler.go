// Package grpcutil implements gRPC HTTP handler. Useful to serve gRPC requests from an existing HTTP server.
package grpcutil

import (
	"net/http"
	"strings"

	"google.golang.org/grpc"
)

// Handler serves gRPC requests or hands over the request to the next handler if no
// application/grpc content type is found in the request.
func Handler(h http.Handler, server *grpc.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if r.ProtoMajor == 2 && strings.Contains(contentType, "application/grpc") {
			server.ServeHTTP(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}
