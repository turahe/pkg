package redis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turahe/pkg/config"
)

func resetRedisState() {
	rdb = nil
	rdbCluster = nil
	isCluster = false
}

func readRESPCommand(br *bufio.Reader) ([]string, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("expected array, got %q", line)
	}
	n, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		hlen, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		hlen = strings.TrimSpace(hlen)
		if !strings.HasPrefix(hlen, "$") {
			return nil, fmt.Errorf("expected bulk, got %q", hlen)
		}
		size, err := strconv.Atoi(hlen[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2) // data + \r\n
		if _, err := io.ReadFull(br, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
}

func clusterSlotsReply(host, port string) string {
	portNum, _ := strconv.Atoi(port)
	var b strings.Builder
	b.WriteString("*1\r\n*3\r\n:0\r\n:16383\r\n*2\r\n")
	b.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(host), host))
	b.WriteString(fmt.Sprintf(":%d\r\n", portNum))
	return b.String()
}

func startFakeRedisOK(t *testing.T) (host, port string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	host, port, err = net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveFakeRedisOK(conn, host, port)
		}
	}()
	return host, port
}

func helloReply(mode string) string {
	// RESP3 map expected by go-redis v9 protocol negotiation.
	return "%7\r\n" +
		"$6\r\nserver\r\n$5\r\nredis\r\n" +
		"$7\r\nversion\r\n$5\r\n7.2.0\r\n" +
		"$5\r\nproto\r\n:3\r\n" +
		"$2\r\nid\r\n:1\r\n" +
		"$4\r\nmode\r\n$" + strconv.Itoa(len(mode)) + "\r\n" + mode + "\r\n" +
		"$4\r\nrole\r\n$6\r\nmaster\r\n" +
		"$7\r\nmodules\r\n*0\r\n"
}

func serveFakeRedisOK(c net.Conn, host, port string) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	br := bufio.NewReader(c)
	for {
		cmd, err := readRESPCommand(br)
		if err != nil {
			return
		}
		if len(cmd) == 0 {
			continue
		}
		name := strings.ToUpper(cmd[0])
		var reply string
		switch name {
		case "HELLO":
			reply = helloReply("standalone")
		case "PING":
			reply = "+PONG\r\n"
		case "SELECT", "AUTH", "CLIENT", "READONLY":
			reply = "+OK\r\n"
		case "COMMAND":
			// go-redis cluster probes command info; empty array is fine.
			reply = "*0\r\n"
		case "CLUSTER":
			if len(cmd) > 1 && strings.EqualFold(cmd[1], "SLOTS") {
				reply = clusterSlotsReply(host, port)
			} else {
				reply = "+OK\r\n"
			}
		default:
			reply = "+OK\r\n"
		}
		if _, err := io.WriteString(c, reply); err != nil {
			return
		}
	}
}

func startFakeRedisClusterSelectErr(t *testing.T) (host, port string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				br := bufio.NewReader(c)
				for {
					cmd, err := readRESPCommand(br)
					if err != nil {
						return
					}
					if len(cmd) == 0 {
						continue
					}
					name := strings.ToUpper(cmd[0])
					var reply string
					switch name {
					case "HELLO":
						reply = helloReply("standalone")
					case "SELECT":
						reply = "-ERR SELECT is not allowed in cluster mode\r\n"
					case "PING":
						reply = "+PONG\r\n"
					default:
						reply = "+OK\r\n"
					}
					if _, err := io.WriteString(c, reply); err != nil {
						return
					}
				}
			}(conn)
		}
	}()

	host, port, err = net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	return host, port
}

func TestSetup_RedisDisabled(t *testing.T) {
	resetRedisState()
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{Enabled: false},
	}
	if err := Setup(); err != nil {
		t.Errorf("Setup with Redis disabled: %v", err)
	}
}

func TestSetup_StandardSuccess(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	host, port := startFakeRedisOK(t)
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled:         true,
			Host:            host,
			Port:            port,
			PoolSize:        10,
			MinIdleConns:    2,
			ReadTimeoutSec:  1,
			WriteTimeoutSec: 1,
		},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if isCluster {
		t.Fatal("expected standard mode")
	}
	if GetRedis() == nil {
		t.Fatal("GetRedis nil")
	}
	if GetUniversalClient() == nil {
		t.Fatal("GetUniversalClient nil")
	}
	if !IsAlive() {
		t.Error("IsAlive should be true")
	}
	if err := Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestSetup_StandardConnectError(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    "1", // privileged / refused
		},
	}
	err := Setup()
	if err == nil {
		t.Fatal("expected connect error")
	}
	if !strings.Contains(err.Error(), "failed to connect to Redis") {
		t.Errorf("error = %v", err)
	}
}

func TestSetup_StandardClusterSelectHint(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	host, port := startFakeRedisClusterSelectErr(t)
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled: true,
			Host:    host,
			Port:    port,
			DB:      1, // triggers SELECT
		},
	}
	err := Setup()
	if err == nil {
		t.Fatal("expected cluster-mode hint error")
	}
	if !strings.Contains(err.Error(), "REDIS_CLUSTER_MODE") {
		t.Errorf("error = %v", err)
	}
}

func TestSetup_ClusterSuccess_WithNodes(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	host, port := startFakeRedisOK(t)
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled:         true,
			ClusterMode:     true,
			ClusterNodes:    fmt.Sprintf("%s:%s, ,%s", host, port, host), // empty entry + host without port
			Port:            port,
			PoolSize:        5,
			MinIdleConns:    1,
			ReadTimeoutSec:  1,
			WriteTimeoutSec: 1,
		},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup cluster: %v", err)
	}
	if !isCluster {
		t.Fatal("expected cluster mode")
	}
	if GetRedisCluster() == nil {
		t.Fatal("GetRedisCluster nil")
	}
	if GetUniversalClient() == nil {
		t.Fatal("GetUniversalClient nil")
	}
	if !IsAlive() {
		t.Error("IsAlive should be true in cluster mode")
	}
	if err := Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestSetup_ClusterSuccess_FallbackHost(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	host, port := startFakeRedisOK(t)
	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled:     true,
			ClusterMode: true,
			Host:        host,
			Port:        port,
		},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup cluster fallback: %v", err)
	}
	_ = Close()
}

func TestSetup_ClusterNoNodes(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled:      true,
			ClusterMode:  true,
			ClusterNodes: " , , ",
			Host:         "",
			Port:         "",
		},
	}
	err := Setup()
	if err == nil || !strings.Contains(err.Error(), "no Redis cluster nodes") {
		t.Fatalf("expected no nodes error, got %v", err)
	}
}

func TestSetup_ClusterConnectError(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	config.Config = &config.Configuration{
		Redis: config.RedisConfiguration{
			Enabled:     true,
			ClusterMode: true,
			Host:        "127.0.0.1",
			Port:        "1",
		},
	}
	err := Setup()
	if err == nil || !strings.Contains(err.Error(), "failed to connect to Redis cluster") {
		t.Fatalf("expected cluster connect error, got %v", err)
	}
}

func TestIsAlive_WhenNotSetup(t *testing.T) {
	resetRedisState()
	if IsAlive() {
		t.Error("IsAlive should be false when Redis is not setup")
	}
}

func TestIsAlive_ClusterNil(t *testing.T) {
	resetRedisState()
	defer resetRedisState()
	isCluster = true
	rdbCluster = nil
	if IsAlive() {
		t.Error("IsAlive should be false when cluster client is nil")
	}
}

func TestIsAlive_PingFails(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	if IsAlive() {
		t.Error("IsAlive should be false when ping fails")
	}

	resetRedisState()
	isCluster = true
	rdbCluster = redis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{"127.0.0.1:1"}})
	if IsAlive() {
		t.Error("IsAlive should be false when cluster ping fails")
	}
}

func TestAvailable(t *testing.T) {
	if Available("127.0.0.1", "63999", 10*time.Millisecond) {
		t.Error("Available should be false for unreachable host:port")
	}
	if Available("192.0.2.1", "6379", 10*time.Millisecond) {
		t.Error("Available should be false for unreachable address (TEST-NET)")
	}

	host, port := startFakeRedisOK(t)
	if !Available(host, port, time.Second) {
		t.Error("Available should be true for listening fake redis")
	}
}

func TestGetRedis_Panics(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	t.Run("cluster mode", func(t *testing.T) {
		isCluster = true
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = GetRedis()
	})

	t.Run("not initialized", func(t *testing.T) {
		isCluster = false
		rdb = nil
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = GetRedis()
	})
}

func TestGetRedisCluster_Panics(t *testing.T) {
	resetRedisState()
	defer resetRedisState()

	t.Run("not cluster mode", func(t *testing.T) {
		isCluster = false
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = GetRedisCluster()
	})

	t.Run("not initialized", func(t *testing.T) {
		isCluster = true
		rdbCluster = nil
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		_ = GetRedisCluster()
	})
}

func TestClose_Noop(t *testing.T) {
	resetRedisState()
	if err := Close(); err != nil {
		t.Errorf("Close noop: %v", err)
	}
}

func TestClose_Standard(t *testing.T) {
	resetRedisState()
	defer resetRedisState()
	host, port := startFakeRedisOK(t)
	rdb = redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port)})
	isCluster = false
	if err := Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	if rdb != nil {
		t.Error("rdb should be nil after Close")
	}
}
