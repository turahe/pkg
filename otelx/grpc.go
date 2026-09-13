package otelx

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

// NewGRPCServerHandler returns an OpenTelemetry gRPC stats.Handler for servers.
// Call after Init so spans use the configured TracerProvider.
func NewGRPCServerHandler(opts ...otelgrpc.Option) stats.Handler {
	opts = append([]otelgrpc.Option{
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	}, opts...)
	return otelgrpc.NewServerHandler(opts...)
}

// NewGRPCClientHandler returns an OpenTelemetry gRPC stats.Handler for clients.
// Call after Init so spans use the configured TracerProvider.
func NewGRPCClientHandler(opts ...otelgrpc.Option) stats.Handler {
	opts = append([]otelgrpc.Option{
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	}, opts...)
	return otelgrpc.NewClientHandler(opts...)
}

// GRPCServerOption returns grpc.StatsHandler(NewGRPCServerHandler(opts...)).
func GRPCServerOption(opts ...otelgrpc.Option) grpc.ServerOption {
	return grpc.StatsHandler(NewGRPCServerHandler(opts...))
}

// GRPCDialOption returns grpc.WithStatsHandler(NewGRPCClientHandler(opts...)).
func GRPCDialOption(opts ...otelgrpc.Option) grpc.DialOption {
	return grpc.WithStatsHandler(NewGRPCClientHandler(opts...))
}
