package grpcutil_test

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/c4milo/handlers/grpcutil"
)

func ExampleHandler() {
	tlsKeyPair, err := tls.LoadX509KeyPair("testdata/cert.pem", "testdata/key.pem")
	if err != nil {
		panic(err)
	}

	serverOpts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&tlsKeyPair)),
	}

	grpcServer := grpc.NewServer(serverOpts...)
	// Register your gRPC services with the grpcServer

	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hola!")
	})

	rack := grpcutil.Handler(mux, grpcServer)

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: rack,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsKeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}
