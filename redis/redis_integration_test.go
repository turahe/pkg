package redis

import (
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestSetup_RedisEnabled_Integration(t *testing.T) {
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    "6379",
			DB:      0,
		},
	}
	defer func() { rdb = nil }()

	err := Setup()
	if err != nil {
		t.Skipf("Redis not available (start with docker compose): %v", err)
	}

	if !IsAlive() {
		t.Error("IsAlive should be true when Redis is connected")
	}

	// Test basic operations
	key := "pkg:integration:test"
	if err := Set(key, "value", 10*time.Second); err != nil {
		t.Errorf("Set: %v", err)
	}
	val, err := Get(key)
	if err != nil {
		t.Errorf("Get: %v", err)
	}
	if val != "value" {
		t.Errorf("Get: got %q, want %q", val, "value")
	}
	if err := Delete(key); err != nil {
		t.Errorf("Delete: %v", err)
	}
	val, _ = Get(key)
	if val != "" {
		t.Errorf("Get after Delete: got %q, want empty", val)
	}
}
