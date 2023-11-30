package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/autom8ter/protoc-gen-ratelimit/example/gen/example"
	"github.com/autom8ter/protoc-gen-ratelimit/example/server"
	"github.com/autom8ter/protoc-gen-ratelimit/limiter"
)

func runServer() error {
	// create a new ratelimiter from the generated function(protoc-gen-ratelimit)
	limit, err := example.NewRateLimiter()
	if err != nil {
		return err
	}
	// create a new grpc server with the ratelimitr interceptors
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			limiter.UnaryServerInterceptor(limit),
		),
		grpc.StreamInterceptor(
			limiter.StreamServerInterceptor(limit),
		),
	)
	// register the example service
	example.RegisterExampleServiceServer(srv, server.NewExampleServer())
	lis, err := net.Listen("tcp", ":10042")
	if err != nil {
		return err
	}
	defer lis.Close()
	fmt.Println("starting server on :10042")
	// start the server
	if err := srv.Serve(lis); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}
