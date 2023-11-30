package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/autom8ter/proto/gen/ratelimit"
	"github.com/go-redis/redis_rate/v10"
	redis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	metadata2 "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type RedisLimiter struct {
	limiter *redis_rate.Limiter
	opts    map[string][]*ratelimit.RateLimitOptions
}

func NewLimiter(client *redis.Client, opts map[string][]*ratelimit.RateLimitOptions) *RedisLimiter {
	return &RedisLimiter{
		limiter: redis_rate.NewLimiter(client),
		opts:    opts,
	}
}

func (r *RedisLimiter) LimitMethod(ctx context.Context, method string) error {
	opts, ok := r.opts[method]
	if !ok {
		return nil
	}
	metadata, ok := metadata2.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "no metadata found in context")
	}
	for _, opt := range opts {
		header, ok := metadata[opt.MetadataKey]
		if !ok {
			continue
		}
		redisKey := fmt.Sprintf("%s:%s:%s", method, opt.MetadataKey, strings.Join(header, ""))
		res, err := r.limiter.Allow(ctx, redisKey, redis_rate.PerMinute(int(opt.Limit)))
		if err != nil {
			return status.Errorf(codes.Internal, "failed to rate limit: %v", err)
		}
		if res.Allowed == 0 {
			if opt.Message != "" {
				return status.Errorf(codes.ResourceExhausted, opt.Message)
			}
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
	}

	return nil
}

func (r *RedisLimiter) Limit(ctx context.Context) error {
	method, ok := grpc.Method(ctx)
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "no method found in context")
	}
	return r.LimitMethod(ctx, method)
}
