# protoc-gen-ratelimit 🛡️

![GoDoc](https://godoc.org/github.com/autom8ter/protoc-gen-ratelimit?status.svg)

**protoc-gen-ratelimit** is an innovative protoc plugin and library 🌟 designed to simplify gRPC request
ratelimiting.
It seamlessly integrates ratelimit configuration options directly within your proto files 📝, reducing the need to clutter your
application code with complex code to handle ratelimiting.
Perfect for developers 👨‍💻👩‍💻 looking to streamline their workflows in gRPC applications.
In this README, you'll find easy installation instructions 📥, examples 💡, and all you need to harness the power of
expression-based rules for robust and efficient request handling 💼.

## Features

- [x] Unary and Stream gRPC interceptors
- [x] Highly configurable rate limiting by method
- [x] Supports multiple ratelimit providers (Redis, In-Memory, etc)

## Installation

The plugin can be installed with the following command:

```bash
    go install github.com/autom8ter/protoc-gen-ratelimit
```

## Code Generation

buf.gen.yaml example:

```yaml
version: v1
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: gen
    opt: paths=source_relative
  - plugin: buf.build/grpc/go
    out: gen
    opt:
      - paths=source_relative
  - plugin: ratelimit
    out: gen
    opt:
      - paths=source_relative
      - limiter=inmem # or limiter=redis
```

## Example

```protobuf
import "ratelimit/ratelimit.proto";

// Example service is an example of how to configure method based ratelimits(per minute)
service ExampleService {
  rpc Allow100PerMinute(google.protobuf.Empty) returns (google.protobuf.Empty) {
    // limit by authorization header
    option (ratelimit.options) = {
      limit: 100
      message: "You have exceeded your rate limit (100 per minute)"
      metadata_key: "authorization"
    };
    // limit by api key
    option (ratelimit.options) = {
      limit: 1000
      message: "You have exceeded your rate limit (100 per minute)"
      metadata_key: "x-api-key"
    };
  }
  rpc Allow1PerMinute(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (ratelimit.options) = {
      limit: 1
      message: "You have exceeded your rate limit (1 per minute)"
      metadata_key: "authorization"
    };
  }
}

```

```go
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
```

See [example](example) for the full example.
