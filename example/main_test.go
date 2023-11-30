package main

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/autom8ter/protoc-gen-ratelimit/example/gen/example"
)

func Test(t *testing.T) {
	go func() {
		if err := runServer(); err != nil {
			panic(err)
		}
	}()
	// wait for server to start
	time.Sleep(1 * time.Second)
	conn, err := grpc.Dial(":10042", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := example.NewExampleServiceClient(conn)
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer 1234567890")
	// test 100 per minute
	for i := 0; i < 100; i++ {
		if _, err := client.Allow100PerMinute(ctx, &emptypb.Empty{}); err != nil {
			t.Fatalf("failed to call Allow100PerMinute: %v", err)
		}
	}

	for i := 0; i < 100; i++ {
		_, err := client.Allow1PerMinute(ctx, &emptypb.Empty{})
		if err == nil && i > 0 {
			t.Fatalf("expected error but got nil")
		}
		if err != nil && i == 0 {
			t.Fatalf("expected nil but got error: %v", err)
		}
	}
}
