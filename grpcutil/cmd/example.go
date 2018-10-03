// Package main shows an example of gRPC handler usage.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/c4milo/handlers/grpcutil"
	"github.com/c4milo/handlers/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var usage = `
Usage:
example server               Runs server
example client               Runs client
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(usage)
		os.Exit(1)
	}

	tlsKeyPair, err := tls.LoadX509KeyPair("testdata/selfsigned.pem", "testdata/selfsigned-key.pem")
	if err != nil {
		panic(err)
	}

	if args[0] == "server" {
		server(tlsKeyPair)
		return
	}

	if args[0] == "client" {
		client(tlsKeyPair)
		return
	}

	log.Fatalf(usage)
	os.Exit(1)
}

type Service struct{}

// Hola prints greeting message
func (s *Service) Hola(ctx context.Context, r *grpcutil.HolaRequest) (*grpcutil.HolaResponse, error) {
	return &grpcutil.HolaResponse{Greeting: "Hola from gRPC service!"}, nil
}

func registerService(binding grpcutil.ServiceBinding) error {
	grpcutil.RegisterTestServer(binding.GRPCServer, new(Service))
	return grpcutil.RegisterTestHandler(context.Background(), binding.GRPCGatewayMuxer, binding.GRPCGatewayClient)
}

func server(cert tls.Certificate) {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hola from HTTP handler!")
	})

	handler := logger.Handler(mux)
	options := []grpcutil.Option{
		grpcutil.WithTLSCert(&cert),
		grpcutil.WithPort("8080"),
		grpcutil.WithServices([]grpcutil.ServiceRegisterFn{registerService}),
	}
	handler = grpcutil.Handler(handler, options...)

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"},
		},
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if srv != nil {
				srv.Close()
			}
		}
	}()

	done := make(chan bool)
	fmt.Printf("Starting server in %q... ", srv.Addr)
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			done <- true
			if err != http.ErrServerClosed {
				panic(err)
			}
		} else {
			done <- true
		}
	}()
	fmt.Println("done")
	<-done
}

func client(cert tls.Certificate) {
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(x509Cert)

	clientCreds := credentials.NewClientTLSFromCert(certPool, "")
	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(clientCreds),
	}

	clientConn, err := grpc.Dial("localhost:8080", clientOpts...)
	if err != nil {
		panic(err)
	}

	defer clientConn.Close()

	test := grpcutil.NewTestClient(clientConn)
	res, err := test.Hola(context.Background(), &grpcutil.HolaRequest{})
	if err != nil {
		panic(err)
	}

	log.Println(res.Greeting)
}
