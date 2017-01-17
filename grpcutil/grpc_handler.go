// Package grpcutil implements a gRPC HTTP handler and OpenAPI proxy. Useful to serve gRPC requests from an existing HTTP servers.
package grpcutil

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServiceBinding are the gRPC server, and gRPC HTTP gateway to which the service will be bound to.
type ServiceBinding struct {
	GRPCServer        *grpc.Server
	GRPCGatewayClient *grpc.ClientConn
	GRPCGatewayMuxer  *runtime.ServeMux
}

// ServiceRegisterFn defines a function type for registering gRPC services.
type ServiceRegisterFn func(ServiceBinding) error

// Option implements http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type Option func(*options)

type options struct {
	serverOpts []grpc.ServerOption
	services   []ServiceRegisterFn
	cert       *tls.Certificate
	port       string
	skipPaths  []string
}

// WithServerOpts sets gRPC server options. Optional.
func WithServerOpts(opts []grpc.ServerOption) Option {
	return func(o *options) {
		o.serverOpts = opts
	}
}

// WithServices sets the list of services to register with the gRPC and OpenAPI servers. Required.
func WithServices(services []ServiceRegisterFn) Option {
	return func(o *options) {
		o.services = services
	}
}

// WithTLSCert sets the TLS certificate to use by the gRPC server and OpenAPI gRPC client. Required.
func WithTLSCert(cert *tls.Certificate) Option {
	return func(o *options) {
		o.cert = cert
	}
}

// WithPort sets the port for the OpenAPI gRPC client to use when connecting to the gRPC service.
func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

// WithSkipPath allows other handlers to serve static JSON files by instructing the GRPC Gateway Muxer to
// skip serving the given prefixed paths.
func WithSkipPath(path ...string) Option {
	return func(o *options) {
		o.skipPaths = path
	}
}

// Handler serves gRPC and OpenAPI requests or hands over to the next handler if they are not gRPC
// or application/json requests.
func Handler(h http.Handler, opts ...Option) http.Handler {
	options := new(options)
	for _, opt := range opts {
		opt(options)
	}

	if options.cert == nil {
		log.Fatal("grpcutil: TLS certificate required")
	}

	if options.port == "" {
		log.Fatal("grpcutil: port required")
	}

	if len(options.services) == 0 {
		log.Fatal("grpcutil: the list of gRPC services is required")
	}

	options.serverOpts = append(options.serverOpts, grpc.Creds(credentials.NewServerTLSFromCert(options.cert)))
	server := grpc.NewServer(options.serverOpts...)

	certPool := x509.NewCertPool()
	x509Cert, err := x509.ParseCertificate(options.cert.Certificate[0])
	if err != nil {
		log.Fatalf("grpcutil: failed parsing x509 certificate: %+v, ", err)
	}
	certPool.AddCert(x509Cert)
	clientCreds := credentials.NewClientTLSFromCert(certPool, "")

	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithUserAgent("grpc-gw"),
	}

	address := "localhost:" + options.port
	clientConn, err := grpc.Dial(address, clientOpts...)
	if err != nil {
		log.Fatalf("failed to connect to local gRPC server: %v", err)
	}

	gwMuxer := runtime.NewServeMux()

	serviceBinding := ServiceBinding{
		GRPCServer:        server,
		GRPCGatewayClient: clientConn,
		GRPCGatewayMuxer:  gwMuxer,
	}

	log.Println("registering GRPC services...")
	for _, register := range options.services {
		if err := register(serviceBinding); err != nil {
			log.Fatalf("failed to register service: %+v", err)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if r.ProtoMajor == 2 && strings.Contains(contentType, "application/grpc") {
			server.ServeHTTP(w, r)
			return
		}

		for _, p := range options.skipPaths {
			if strings.HasPrefix(r.URL.Path, p) {
				h.ServeHTTP(w, r)
				return
			}
		}

		accept := r.Header.Get("Accept")
		if strings.Contains(contentType, "application/json") ||
			strings.Contains(accept, "application/json") {
			gwMuxer.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
