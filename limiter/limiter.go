package limiter

import (
	"context"

	"google.golang.org/grpc"
)

// Limiter is a function that returns a context and error
type Limiter func(ctx context.Context) error

// UnaryServerInterceptor returns a new unary server interceptor that performs rate limiting on the request.
func UnaryServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		err := limiter(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that performs rate limiting on the request.
func StreamServerInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := limiter(stream.Context())
		if err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

// Chain returns a Limiter that executes each limiter in sequence.
// If any of the limiters returns an error, the chain stops executing and returns that error.
func Chain(limiters ...Limiter) Limiter {
	return func(ctx context.Context) error {
		for _, limiter := range limiters {
			if err := limiter(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
