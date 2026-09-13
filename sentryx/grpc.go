package sentryx

import (
	"time"

	sentrygrpc "github.com/getsentry/sentry-go/grpc"
	"google.golang.org/grpc"
)

// GRPCServerOptions configures Sentry gRPC server interceptors.
type GRPCServerOptions struct {
	// Repanic re-panics after recovery so other interceptors can still see the panic.
	// Defaults to true when using ServerInterceptorOptions / default helpers.
	Repanic bool
	// WaitForDelivery blocks until the panic event is flushed to Sentry.
	WaitForDelivery bool
	// Timeout is the max wait for WaitForDelivery. Zero uses Sentry's default flush timeout.
	Timeout time.Duration
}

func (o GRPCServerOptions) toSentry() sentrygrpc.ServerOptions {
	return sentrygrpc.ServerOptions{
		Repanic:         o.Repanic,
		WaitForDelivery: o.WaitForDelivery,
		Timeout:         o.Timeout,
	}
}

// UnaryServerInterceptor returns a Sentry unary server interceptor.
// Prefer ServerOptions() when wiring a new grpc.Server.
func UnaryServerInterceptor(opts GRPCServerOptions) grpc.UnaryServerInterceptor {
	return sentrygrpc.UnaryServerInterceptor(opts.toSentry())
}

// StreamServerInterceptor returns a Sentry stream server interceptor.
func StreamServerInterceptor(opts GRPCServerOptions) grpc.StreamServerInterceptor {
	return sentrygrpc.StreamServerInterceptor(opts.toSentry())
}

// UnaryClientInterceptor returns a Sentry unary client interceptor.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return sentrygrpc.UnaryClientInterceptor()
}

// StreamClientInterceptor returns a Sentry stream client interceptor.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return sentrygrpc.StreamClientInterceptor()
}

// ServerOptions returns unary + stream Sentry interceptors for grpc.NewServer.
// Defaults Repanic to true so panics still abort the RPC after capture.
func ServerOptions(opts ...GRPCServerOptions) []grpc.ServerOption {
	o := GRPCServerOptions{Repanic: true}
	if len(opts) > 0 {
		o = opts[0]
	}
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(UnaryServerInterceptor(o)),
		grpc.ChainStreamInterceptor(StreamServerInterceptor(o)),
	}
}

// DialOptions returns unary + stream Sentry interceptors for grpc.NewClient / Dial.
func DialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(StreamClientInterceptor()),
	}
}
