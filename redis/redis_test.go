package redis

import (
	"testing"

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
