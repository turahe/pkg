package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/turahe/pkg/config"
)

// integrationBackends returns Redis and Valkey backends to try. When REDIS_HOST and/or
// VALKEY_HOST are set (e.g. in docker-compose.test), uses those; otherwise uses localhost.
func integrationBackends() []struct{ name, host, port string } {
	defaultBackends := []struct{ name, host, port string }{
		{"Redis", "127.0.0.1", "6379"},
		{"Valkey", "127.0.0.1", "6380"},
	}
	redisHost := os.Getenv("REDIS_HOST")
	valkeyHost := os.Getenv("VALKEY_HOST")
	if redisHost == "" && valkeyHost == "" {
		return defaultBackends
	}
	var backends []struct{ name, host, port string }
	if redisHost != "" {
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		backends = append(backends, struct{ name, host, port string }{"Redis", redisHost, port})
	}
	if valkeyHost != "" {
		port := os.Getenv("VALKEY_PORT")
		if port == "" {
			port = "6379"
		}
		backends = append(backends, struct{ name, host, port string }{"Valkey", valkeyHost, port})
	}
	if len(backends) == 0 {
		return defaultBackends
	}
	return backends
}

const integrationTimeout = 500 * time.Millisecond
const keyPrefix = "pkg:integration:"

// setupIntegrationBackend sets config and calls Setup() using the first available
// backend (Redis or Valkey). Skips the test if none is reachable so go test ./... passes without Redis.
func setupIntegrationBackend(t *testing.T) {
	t.Helper()
	for _, b := range integrationBackends() {
		if !Available(b.host, b.port, integrationTimeout) {
			continue
		}
		config.Config = &config.Configuration{
			Redis: config.RedisConfiguration{
				Enabled: true,
				Host:    b.host,
				Port:    b.port,
				DB:      0,
			},
		}
		if err := Setup(); err != nil {
			t.Fatalf("%s at %s:%s: Setup failed: %v", b.name, b.host, b.port, err)
		}
		t.Logf("Using %s at %s:%s", b.name, b.host, b.port)
		return
	}
	t.Skip("No Redis or Valkey available. Start one with: docker compose up -d (Redis:6379 or Valkey:6380)")
}

// setupIntegrationBackendOrSkip sets config and calls Setup() using the first available
// backend. Skips the test if no backend is reachable (so go test ./... passes without Redis).
func setupIntegrationBackendOrSkip(t *testing.T) {
	setupIntegrationBackend(t)
}

func TestIntegration_GetSetDelete(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	key := keyPrefix + "getset"
	defer func() { _ = Delete(ctx, key) }()

	if err := Set(ctx, key, "value", 10*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, err := Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "value" {
		t.Errorf("Get: got %q, want %q", val, "value")
	}
	if err := Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	val, _ = Get(ctx, key)
	if val != "" {
		t.Errorf("Get after Delete: got %q, want empty", val)
	}
}

func TestIntegration_GetMissingKey(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	val, err := Get(context.Background(), keyPrefix+"nonexistent")
	if err != nil {
		t.Fatalf("Get missing key: %v", err)
	}
	if val != "" {
		t.Errorf("Get missing key: got %q, want empty", val)
	}
}

func TestIntegration_IsAlive(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	if !IsAlive() {
		t.Error("IsAlive should be true when connected")
	}
}

func TestIntegration_Hash(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	key := keyPrefix + "hash"
	defer func() { _ = Delete(ctx, key) }()

	if err := HSet(ctx, key, "f1", "v1"); err != nil {
		t.Fatalf("HSet: %v", err)
	}
	if err := HSet(ctx, key, "f2", "v2"); err != nil {
		t.Fatalf("HSet f2: %v", err)
	}
	v, err := HGet(ctx, key, "f1")
	if err != nil {
		t.Fatalf("HGet: %v", err)
	}
	if v != "v1" {
		t.Errorf("HGet f1: got %q, want v1", v)
	}
	all, err := HGetAll(ctx, key)
	if err != nil {
		t.Fatalf("HGetAll: %v", err)
	}
	if len(all) != 2 || all["f1"] != "v1" || all["f2"] != "v2" {
		t.Errorf("HGetAll: got %v", all)
	}
}

func TestIntegration_List(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	key := keyPrefix + "list"
	defer func() { _ = Delete(ctx, key) }()

	if err := LPush(ctx, key, "a", "b", "c"); err != nil {
		t.Fatalf("LPush: %v", err)
	}
	list, err := LRange(ctx, key, 0, -1)
	if err != nil {
		t.Fatalf("LRange: %v", err)
	}
	if len(list) != 3 || list[0] != "c" || list[2] != "a" {
		t.Errorf("LRange: got %v", list)
	}
	popped, err := RPop(ctx, key)
	if err != nil {
		t.Fatalf("RPop: %v", err)
	}
	if popped != "a" {
		t.Errorf("RPop: got %q, want a", popped)
	}
}

func TestIntegration_Set(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	key := keyPrefix + "set"
	defer func() { _ = Delete(ctx, key) }()

	if err := SAdd(ctx, key, "x", "y", "z"); err != nil {
		t.Fatalf("SAdd: %v", err)
	}
	members, err := SMembers(ctx, key)
	if err != nil {
		t.Fatalf("SMembers: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("SMembers: got %v", members)
	}
	if err := SRem(ctx, key, "y"); err != nil {
		t.Fatalf("SRem: %v", err)
	}
	members, _ = SMembers(ctx, key)
	if len(members) != 2 {
		t.Errorf("SMembers after SRem: got %v", members)
	}
}

func TestIntegration_Lock(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	key := keyPrefix + "lock"
	defer func() { _ = ReleaseLock(ctx, key) }()

	ok, err := AcquireLock(ctx, key, "owner1", 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	if !ok {
		t.Error("AcquireLock: expected true")
	}
	ok2, _ := AcquireLock(ctx, key, "owner2", 5*time.Second)
	if ok2 {
		t.Error("AcquireLock second time should fail (lock held)")
	}
	if err := ReleaseLock(ctx, key); err != nil {
		t.Fatalf("ReleaseLock: %v", err)
	}
	ok3, err := AcquireLock(ctx, key, "owner3", 5*time.Second)
	if err != nil || !ok3 {
		t.Errorf("AcquireLock after release: ok=%v err=%v", ok3, err)
	}
}

func TestIntegration_Pipeline(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	k1, k2 := keyPrefix+"pipe1", keyPrefix+"pipe2"
	defer func() { _ = Delete(ctx, k1); _ = Delete(ctx, k2) }()

	err := PipelineSet(ctx, map[string]interface{}{k1: "v1", k2: "v2"}, 10*time.Second)
	if err != nil {
		t.Fatalf("PipelineSet: %v", err)
	}
	vals, err := MGet(ctx, k1, k2)
	if err != nil {
		t.Fatalf("MGet: %v", err)
	}
	if len(vals) != 2 || vals[0] != "v1" || vals[1] != "v2" {
		t.Errorf("MGet after PipelineSet: got %v", vals)
	}
}

func TestIntegration_MSetMGet(t *testing.T) {
	setupIntegrationBackend(t)
	defer func() { rdb = nil; rdbCluster = nil }()

	ctx := context.Background()
	k1, k2 := keyPrefix + "mset1", keyPrefix + "mset2"
	defer func() { _ = Delete(ctx, k1); _ = Delete(ctx, k2) }()

	if err := MSet(ctx, map[string]interface{}{k1: "a", k2: "b"}); err != nil {
		t.Fatalf("MSet: %v", err)
	}
	vals, err := MGet(ctx, k1, k2)
	if err != nil {
		t.Fatalf("MGet: %v", err)
	}
	if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
		t.Errorf("MGet: got %v", vals)
	}
}

func TestIntegration_Close(t *testing.T) {
	setupIntegrationBackend(t)
	// Close() clears rdb/rdbCluster; no need to nil them in defer
	err := Close()
	if err != nil {
		t.Errorf("Close: %v", err)
	}
	if IsAlive() {
		t.Error("IsAlive should be false after Close")
	}
}
