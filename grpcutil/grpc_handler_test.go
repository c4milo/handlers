package grpcutil

//go:generate protoc -I. -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. --grpc-gateway_out=logtostderr=true:. --swagger_out=logtostderr=true:. hola.proto

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hooklift/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Service implements the Hola service defined in hola.proto
type Service struct{}

func (s *Service) Hola(ctx context.Context, r *HolaRequest) (*HolaResponse, error) {
	return &HolaResponse{
		Greeting: "Hola!",
	}, nil
}

// TestHandler runs a series of tests against our gRPC server handler.
func TestHandler(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello HTTP handler")
	})

	cert, err := tls.LoadX509KeyPair("testdata/cert.pem", "testdata/key.pem")
	assert.Ok(t, err)

	serverOpts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}

	grpcServer := grpc.NewServer(serverOpts...)
	RegisterTestServer(grpcServer, new(Service))

	grpcHandler := Handler(requestHandler, grpcServer)
	server := httptest.NewUnstartedServer(grpcHandler)
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2"},
	}
	server.StartTLS()
	defer server.Close()

	// Prepare gRPC client connection
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	assert.Ok(t, err)
	certPool := x509.NewCertPool()
	certPool.AddCert(x509Cert)

	u, err := url.Parse(server.URL)
	assert.Ok(t, err)

	// Tests gRPC service
	clientCreds := credentials.NewClientTLSFromCert(certPool, "")
	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(clientCreds),
	}

	clientConn, err := grpc.Dial("localhost:"+u.Port(), clientOpts...)
	assert.Ok(t, err)
	defer clientConn.Close()

	test := NewTestClient(clientConn)
	res, err := test.Hola(context.Background(), &HolaRequest{})
	assert.Ok(t, err)
	assert.Equals(t, "Hola!", res.Greeting)

	// Test HTTP handler
	// configure a client to use trust those certificates
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
	}
	res2, err := client.Get("https://localhost:" + u.Port())
	assert.Ok(t, err)
	data, err := ioutil.ReadAll(res2.Body)
	assert.Ok(t, err)
	err = res2.Body.Close()
	assert.Ok(t, err)

	assert.Equals(t, "Hello HTTP handler", string(data[:]))
}
