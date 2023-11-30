package inmem_test

import (
	"context"
	"testing"

	"github.com/autom8ter/proto/gen/ratelimit"
	"google.golang.org/grpc/metadata"

	"github.com/autom8ter/protoc-gen-ratelimit/inmem"
)

type fixture struct {
	name      string
	meta      metadata.MD
	opts      map[string][]*ratelimit.RateLimitOptions
	test      func(ctx context.Context, m *inmem.InMemoryLimiter) error
	expectErr bool
}

func Test(t *testing.T) {
	var fixtures = []fixture{
		{
			name: "TestLimitMethod(pass)",
			meta: map[string][]string{
				"authorization": {"1234"},
			},
			opts: map[string][]*ratelimit.RateLimitOptions{
				"ExampleService_Allow100PerMinute_FullMethodName": {
					{
						Limit:       100,
						MetadataKey: "authorization",
						Message:     "You have exceeded your rate limit (100 per minute)",
					},
				},
			},
			test: func(ctx context.Context, m *inmem.InMemoryLimiter) error {
				count := 0
				for i := 0; i < 100; i++ {
					if err := m.LimitMethod(ctx, "ExampleService_Allow100PerMinute_FullMethodName"); err != nil {
						return err
					}
					count++
				}
				return nil
			},
		},
		{
			name: "TestLimitMethod(fail)",
			meta: map[string][]string{
				"authorization": {"1234"},
			},
			opts: map[string][]*ratelimit.RateLimitOptions{
				"ExampleService_Allow100PerMinute_FullMethodName": {
					{
						Limit:       100,
						MetadataKey: "authorization",
						Message:     "You have exceeded your rate limit (100 per minute)",
					},
				},
			},
			test: func(ctx context.Context, m *inmem.InMemoryLimiter) error {
				count := 0
				for i := 0; i < 200; i++ {
					if err := m.LimitMethod(ctx, "ExampleService_Allow100PerMinute_FullMethodName"); err != nil {
						return err
					}
					count++
				}
				return nil
			},
			expectErr: true,
		},
		{
			name: "TestLimitMethod(multiple keys)(pass)",
			meta: map[string][]string{
				"x-api-key": {"1234"},
			},
			opts: map[string][]*ratelimit.RateLimitOptions{
				"ExampleService_Allow100PerMinute_FullMethodName": {
					{
						Limit:       100,
						MetadataKey: "authorization",
						Message:     "You have exceeded your rate limit (100 per minute)",
					},
					{
						Limit:       100,
						MetadataKey: "x-api-key",
						Message:     "You have exceeded your rate limit (100 per minute)",
					},
				},
			},
			test: func(ctx context.Context, m *inmem.InMemoryLimiter) error {
				count := 0
				for i := 0; i < 100; i++ {
					if err := m.LimitMethod(ctx, "ExampleService_Allow100PerMinute_FullMethodName"); err != nil {
						return err
					}
					count++
				}
				return nil
			},
			expectErr: false,
		},
		{
			name: "TestLimitMethod(no metakey match)(pass)",
			meta: map[string][]string{
				"authorization": {"1234"},
			},
			opts: map[string][]*ratelimit.RateLimitOptions{
				"ExampleService_Allow100PerMinute_FullMethodName": {
					{
						Limit:       100,
						MetadataKey: "x-api-key",
						Message:     "You have exceeded your rate limit (100 per minute)",
					},
				},
			},
			test: func(ctx context.Context, m *inmem.InMemoryLimiter) error {
				count := 0
				for i := 0; i < 100; i++ {
					if err := m.LimitMethod(ctx, "ExampleService_Allow100PerMinute_FullMethodName"); err != nil {
						return err
					}
					count++
				}
				return nil
			},
		},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), f.meta)
			m := inmem.NewLimiter(f.opts)
			if err := f.test(ctx, m); err != nil {
				if !f.expectErr {
					t.Fatal(err)
				}
			}
		})
	}
}
