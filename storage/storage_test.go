package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/turahe/pkg/config"
)

func TestSetup_Disabled(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})

	config.Config = &config.Configuration{Storage: config.StorageConfiguration{}}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	_, err := ReadObject(context.Background(), "x")
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("got %v, want ErrNotInitialized", err)
	}
}

func TestSetupContext_DisabledCancelledOK(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})

	config.Config = &config.Configuration{Storage: config.StorageConfiguration{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SetupContext(ctx); err != nil {
		t.Fatalf("SetupContext: %v", err)
	}
}

func TestSetup_GCSRequiresBucket(t *testing.T) {
	t.Cleanup(func() {
		config.Config = nil
		_ = Close()
	})

	config.Config = &config.Configuration{
		Storage: config.StorageConfiguration{Driver: "gcs"},
	}

	if err := Setup(); err == nil {
		t.Fatal("expected error when bucket missing")
	}
}
