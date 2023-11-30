package server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/autom8ter/protoc-gen-ratelimit/example/gen/example"
)

type exampleServer struct {
	example.UnimplementedExampleServiceServer
}

func NewExampleServer() example.ExampleServiceServer {
	return &exampleServer{}
}

func (e *exampleServer) Allow100PerMinute(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (e *exampleServer) Allow1PerMinute(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
