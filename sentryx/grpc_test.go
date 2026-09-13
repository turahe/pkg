package sentryx

import (
	"testing"

	"google.golang.org/grpc"
)

func TestServerOptions(t *testing.T) {
	opts := ServerOptions()
	if len(opts) != 2 {
		t.Fatalf("got %d options, want 2", len(opts))
	}
	_ = grpc.NewServer(opts...)
}

func TestDialOptions(t *testing.T) {
	opts := DialOptions()
	if len(opts) != 2 {
		t.Fatalf("got %d options, want 2", len(opts))
	}
}

func TestUnaryInterceptors(t *testing.T) {
	if UnaryServerInterceptor(GRPCServerOptions{Repanic: true}) == nil {
		t.Fatal("expected unary server interceptor")
	}
	if StreamServerInterceptor(GRPCServerOptions{Repanic: true}) == nil {
		t.Fatal("expected stream server interceptor")
	}
	if UnaryClientInterceptor() == nil {
		t.Fatal("expected unary client interceptor")
	}
	if StreamClientInterceptor() == nil {
		t.Fatal("expected stream client interceptor")
	}
}
