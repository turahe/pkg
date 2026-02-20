package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/pkg/config"

	"github.com/redis/go-redis/v9"
)

var (
	rdb        *redis.Client
	rdbCluster *redis.ClusterClient
	isCluster  bool
)

func Setup() error {
	configuration := config.GetConfig()

	if !configuration.Redis.Enabled {
		return nil
	}

	// Check if cluster mode is enabled
	if configuration.Redis.ClusterMode {
		return setupClusterClient(configuration)
	}

	// Setup standard Redis client
	return setupStandardClient(configuration)
}

func setupStandardClient(configuration *config.Configuration) error {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", configuration.Redis.Host, configuration.Redis.Port),
		Password: configuration.Redis.Password,
		DB:       configuration.Redis.DB,
	}
	if configuration.Redis.PoolSize > 0 {
		opts.PoolSize = configuration.Redis.PoolSize
	}
	if configuration.Redis.MinIdleConns > 0 {
		opts.MinIdleConns = configuration.Redis.MinIdleConns
	}
	if configuration.Redis.ReadTimeoutSec > 0 {
		opts.ReadTimeout = time.Duration(configuration.Redis.ReadTimeoutSec) * time.Second
	}
	if configuration.Redis.WriteTimeoutSec > 0 {
		opts.WriteTimeout = time.Duration(configuration.Redis.WriteTimeoutSec) * time.Second
	}
	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		// Check if the error is related to cluster mode
		if strings.Contains(err.Error(), "SELECT is not allowed in cluster mode") {
			return fmt.Errorf("Redis server is in cluster mode, but REDIS_CLUSTER_MODE is not enabled. Please set REDIS_CLUSTER_MODE=true in your configuration: %w", err)
		}
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	rdb = client
	isCluster = false
	return nil
}

func setupClusterClient(configuration *config.Configuration) error {
	var addrs []string

	if configuration.Redis.ClusterNodes != "" {
		// Parse comma-separated cluster nodes
		nodes := strings.Split(configuration.Redis.ClusterNodes, ",")
		for _, node := range nodes {
			trimmed := strings.TrimSpace(node)
			if trimmed != "" {
				// Ensure node has port, use default if not specified
				if !strings.Contains(trimmed, ":") {
					trimmed = fmt.Sprintf("%s:%s", trimmed, configuration.Redis.Port)
				}
				addrs = append(addrs, trimmed)
			}
		}
	} else {
		// Fallback to single host:port if cluster nodes not specified
		addrs = []string{fmt.Sprintf("%s:%s", configuration.Redis.Host, configuration.Redis.Port)}
	}

	if len(addrs) == 0 {
		return fmt.Errorf("no Redis cluster nodes configured")
	}

	clusterOpts := &redis.ClusterOptions{
		Addrs:        addrs,
		Password:     configuration.Redis.Password,
		MaxRedirects: 3,
	}
	if configuration.Redis.PoolSize > 0 {
		clusterOpts.PoolSize = configuration.Redis.PoolSize
	}
	if configuration.Redis.MinIdleConns > 0 {
		clusterOpts.MinIdleConns = configuration.Redis.MinIdleConns
	}
	if configuration.Redis.ReadTimeoutSec > 0 {
		clusterOpts.ReadTimeout = time.Duration(configuration.Redis.ReadTimeoutSec) * time.Second
	}
	if configuration.Redis.WriteTimeoutSec > 0 {
		clusterOpts.WriteTimeout = time.Duration(configuration.Redis.WriteTimeoutSec) * time.Second
	}
	client := redis.NewClusterClient(clusterOpts)

	// Test cluster connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	rdbCluster = client
	isCluster = true
	return nil
}

func IsAlive() bool {
	if isCluster {
		if rdbCluster == nil {
			return false
		}
		return rdbCluster.Ping(context.Background()).Err() == nil
	}

	if rdb == nil {
		return false
	}

	return rdb.Ping(context.Background()).Err() == nil
}

func GetRedis() *redis.Client {
	if isCluster {
		panic("Redis is in cluster mode. Use GetRedisCluster() instead.")
	}

	if rdb == nil {
		panic("Redis client is not initialized. Call Setup() first.")
	}

	return rdb
}

func GetRedisCluster() *redis.ClusterClient {
	if !isCluster {
		panic("Redis is not in cluster mode. Use GetRedis() instead.")
	}

	if rdbCluster == nil {
		panic("Redis cluster client is not initialized. Call Setup() first.")
	}

	return rdbCluster
}

// GetUniversalClient returns a universal client interface that works with both standard and cluster clients.
func GetUniversalClient() redis.Cmdable {
	if isCluster {
		return rdbCluster
	}
	return rdb
}

// Close closes the active Redis client and releases all connections.
// Safe to call even if Redis was never set up (no-op when not enabled).
func Close() error {
	if isCluster && rdbCluster != nil {
		err := rdbCluster.Close()
		rdbCluster = nil
		return err
	}
	if rdb != nil {
		err := rdb.Close()
		rdb = nil
		return err
	}
	return nil
}
