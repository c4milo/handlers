# gRPC server handler

A convenient handler to initialize a gRPC server and OpenAPI proxy on an existing HTTP server.

## Requirements

1. Install protobuf compiler

```shell
brew install --devel protobuf
```

2. Install gRPC and gRPC gateway code

```shell
go get -u -v google.golang.org/grpc
go get -u -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get -u -v github.com/golang/protobuf/protoc-gen-go
```

3. Run `go generate`

4. Run server `go run cmd/example.go server`

5. Run client `go run cmd/example.go client`
