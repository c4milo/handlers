package grpcutil_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"

	"github.com/c4milo/handlers/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Example_client() {
	tlsKeyPair, err := tls.LoadX509KeyPair("testdata/selfsigned.pem", "testdata/selfsigned-key.pem")
	if err != nil {
		panic(err)
	}

	x509Cert, err := x509.ParseCertificate(tlsKeyPair.Certificate[0])
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
