package otelx

import (
	"testing"

	"google.golang.org/grpc"
)

func TestGRPCServerOption(t *testing.T) {
	opt := GRPCServerOption()
	if opt == nil {
		t.Fatal("expected non-nil server option")
	}
	_ = grpc.NewServer(opt)
}

func TestGRPCDialOption(t *testing.T) {
	opt := GRPCDialOption()
	if opt == nil {
		t.Fatal("expected non-nil dial option")
	}
}

func TestNewGRPCHandlers(t *testing.T) {
	if NewGRPCServerHandler() == nil {
		t.Fatal("expected non-nil server handler")
	}
	if NewGRPCClientHandler() == nil {
		t.Fatal("expected non-nil client handler")
	}
}
