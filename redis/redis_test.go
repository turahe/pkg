package redis

import (
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

func TestSetup_RedisDisabled(t *testing.T) {
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{Enabled: false},
	}
	err := Setup()
	if err != nil {
		t.Errorf("Setup with Redis disabled: %v", err)
	}
}

func TestIsAlive_WhenNotSetup(t *testing.T) {
	// Ensure rdb is nil (e.g. after Setup with Redis.Enabled=false or fresh)
	rdb = nil
	if IsAlive() {
		t.Error("IsAlive should be false when Redis is not setup")
	}
}

func TestAvailable(t *testing.T) {
	// Unreachable host:port should return false
	if Available("127.0.0.1", "63999", 10*time.Millisecond) {
		t.Error("Available should be false for unreachable host:port")
	}
	// Invalid port (non-numeric) still dials; use a valid but closed port for "unreachable"
	if Available("192.0.2.1", "6379", 10*time.Millisecond) {
		t.Error("Available should be false for unreachable address (TEST-NET)")
	}
}
