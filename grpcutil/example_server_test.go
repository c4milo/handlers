package grpcutil_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/c4milo/handlers/grpcutil"
	"github.com/c4milo/handlers/logger"
	"golang.org/x/net/context"
)

type Service struct{}

// Hola prints greeting message
func (s *Service) Hola(ctx context.Context, r *grpcutil.HolaRequest) (*grpcutil.HolaResponse, error) {
	return &grpcutil.HolaResponse{Greeting: "Hola from gRPC service!"}, nil
}

func registerService(binding grpcutil.ServiceBinding) error {
	grpcutil.RegisterTestServer(binding.GRPCServer, new(Service))
	return grpcutil.RegisterTestHandler(context.Background(), binding.GRPCGatewayMuxer, binding.GRPCGatewayClient)
}

func Example_server() {
	tlsKeyPair, err := tls.LoadX509KeyPair("testdata/selfsigned.pem", "testdata/selfsigned-key.pem")
	if err != nil {
		panic(err)
	}

	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hola from HTTP handler!")
	})

	handler := logger.Handler(mux)
	options := []grpcutil.Option{
		grpcutil.WithTLSCert(&tlsKeyPair),
		grpcutil.WithPort("8080"),
		grpcutil.WithServices([]grpcutil.ServiceRegisterFn{registerService}),
	}
	handler = grpcutil.Handler(handler, options...)

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsKeyPair},
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
