package inmem

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/autom8ter/proto/gen/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	metadata2 "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// InMemoryLimiter is a ratelimiter that uses an in-memory map to store counters
type InMemoryLimiter struct {
	opts     map[string][]*ratelimit.RateLimitOptions
	counters map[string]int
	mu       sync.RWMutex
}

// NewLimiter returns a new ratelimiter that uses an in-memory map to store counters
func NewLimiter(opts map[string][]*ratelimit.RateLimitOptions) *InMemoryLimiter {
	i := &InMemoryLimiter{
		opts:     opts,
		counters: map[string]int{},
	}
	// cleanup old counters
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				i.mu.Lock()
				for k := range i.counters {
					split := strings.Split(k, ":")
					if len(split) != 3 {
						continue
					}
					t, err := time.Parse(time.RFC3339, split[0])
					if err != nil {
						continue
					}
					if time.Since(t) > time.Minute {
						delete(i.counters, k)
					}
				}
				i.mu.Unlock()
			}
		}
	}()
	return i
}

// LimitMethod limits the request based on the method and metadata key
func (i *InMemoryLimiter) LimitMethod(ctx context.Context, method string) error {
	opts, ok := i.opts[method]
	if !ok || len(opts) == 0 {
		return nil
	}
	metadata, ok := metadata2.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "no metadata found in context")
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, opt := range opts {
		metaKey, ok := metadata[opt.MetadataKey]
		if !ok {
			continue
		}
		counterKey := fmt.Sprintf("%s:%s:%s:%s", time.Now().Truncate(time.Minute).Format(time.RFC3339), method, opt.MetadataKey, strings.Join(metaKey, ""))
		if i.counters[counterKey] >= int(opt.Limit) {
			if opt.Message != "" {
				return status.Errorf(codes.ResourceExhausted, opt.Message)
			}
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		i.counters[counterKey]++
	}

	return nil
}

// Debug Unsafe - returns the counters map for debugging purposes
func (i *InMemoryLimiter) Debug() map[string]int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.counters
}

// Limit limits the request based on the method and metadata key
func (i *InMemoryLimiter) Limit(ctx context.Context) error {
	method, ok := grpc.Method(ctx)
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "no method found in context")
	}
	return i.LimitMethod(ctx, method)
}
