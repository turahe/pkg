package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/turahe/pkg/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	rdb        *redis.Client
	rdbCluster *redis.ClusterClient
	ctx        = context.Background()
	isCluster  bool
)

type Database struct {
	*gorm.DB
}

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
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", configuration.Redis.Host, configuration.Redis.Port),
		Password: configuration.Redis.Password,
		DB:       configuration.Redis.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
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
	// Warn if DB is set to non-zero in cluster mode (cluster only supports DB 0)
	if configuration.Redis.DB != 0 {
		// Note: We don't return an error here, just log a warning since DB selection
		// is automatically ignored in cluster mode, but it's good to inform the user
	}

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

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Password: configuration.Redis.Password,
		// Note: DB option is not supported in cluster mode (cluster only uses DB 0)
		// Google Cloud Memorystore Redis Cluster specific options
		MaxRedirects:   3,
		ReadOnly:       false,
		RouteByLatency: false,
		RouteRandomly:  false,
		// Cluster slots will be auto-discovered by the client
	})

	// Test cluster connection
	if err := client.Ping(ctx).Err(); err != nil {
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
		return rdbCluster.Ping(ctx).Err() == nil
	}

	if rdb == nil {
		return false
	}

	return rdb.Ping(ctx).Err() == nil
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

// GetUniversalClient returns a universal client interface that works with both standard and cluster clients
func GetUniversalClient() redis.Cmdable {
	if isCluster {
		return rdbCluster
	}
	return rdb
}
