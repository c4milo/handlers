package grpcutil

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ForwardMetadataCLIntcp returns a client unary interceptor that forwards
// incoming context metadata to its outgoing counterpart. In some scenarios,
// using this interceptor may pose security risks since authorization tokens
// or credentials can be accidentally leaked to third-party services. It can
// be very handy otherwise when used with trusted services. Since it allows
// to delegate or impersonate users when reaching out to internal services by
// forwarding their original access tokens or authentication credentials.
func ForwardMetadataCLIntcp() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req,
		reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
