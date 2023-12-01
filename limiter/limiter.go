package limiter

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
)

// Limiter is a function that returns a context and error
type Limiter func(ctx context.Context) error

// UnaryServerInterceptor returns a new unary server interceptor that performs rate limiting on the request.
// If no matchers are provided, the limiter will apply to all requests.
// If matchers are provided, the limiter will only apply to requests that match at least one of the matchers.
func UnaryServerInterceptor(limiter Limiter, matchers ...selector.Matcher) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(matchers) == 0 {
			return unaryServerInterceptor(limiter)(ctx, req, info, handler)
		}
		meta := interceptors.NewServerCallMeta(info.FullMethod, nil, req)
		for _, matcher := range matchers {
			if matcher.Match(ctx, meta) {
				return unaryServerInterceptor(limiter)(ctx, req, info, handler)
			}
		}
		return handler(ctx, req)
	}
}

func unaryServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		err := limiter(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that performs rate limiting on the request.
// If no matchers are provided, the limiter will apply to all requests.
// If matchers are provided, the limiter will only apply to requests that match at least one of the matchers.
func StreamServerInterceptor(limiter Limiter, matchers ...selector.Matcher) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if len(matchers) == 0 {
			return streamServerInterceptor(limiter)(srv, ss, info, handler)
		}
		meta := interceptors.NewServerCallMeta(info.FullMethod, info, nil)
		for _, matcher := range matchers {
			if matcher.Match(ss.Context(), meta) {
				return streamServerInterceptor(limiter)(srv, ss, info, handler)
			}
		}
		return handler(srv, ss)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that performs rate limiting on the request.
func streamServerInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
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
