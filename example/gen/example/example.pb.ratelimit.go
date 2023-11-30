package example

import (
	"github.com/autom8ter/proto/gen/ratelimit"

	"github.com/autom8ter/protoc-gen-ratelimit/inmem"
	"github.com/autom8ter/protoc-gen-ratelimit/limiter"
)

// NewRateLimiter returns a new inmemory ratelimiter
func NewRateLimiter() (limiter.Limiter, error) {
	limit := inmem.NewLimiter(map[string][]*ratelimit.RateLimitOptions{
		ExampleService_Allow100PerMinute_FullMethodName: {
			{
				Limit:       100,
				MetadataKey: "authorization",
				Message:     "You have exceeded your rate limit (100 per minute)",
			},
			{
				Limit:       1000,
				MetadataKey: "x-api-key",
				Message:     "You have exceeded your rate limit (100 per minute)",
			},
		},
		ExampleService_Allow1PerMinute_FullMethodName: {
			{
				Limit:       1,
				MetadataKey: "authorization",
				Message:     "You have exceeded your rate limit (1 per minute)",
			},
		},
	})
	return limit.Limit, nil
}
