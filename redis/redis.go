package redis

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

type Database struct {
	*gorm.DB
}

func Setup() error {
	var client *redis.Client
	configuration := config.GetConfig()

	if configuration.Redis.Enabled {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", configuration.Redis.Host, configuration.Redis.Port),
			Password: configuration.Redis.Password,
			DB:       configuration.Redis.DB,
		})

		if err := client.Ping(ctx).Err(); err != nil {
			return err
		}
	}

	rdb = client

	return nil
}

func IsAlive() bool {
	if rdb == nil {
		return false
	}

	return rdb.Ping(ctx).Err() == nil
}

func GetRedis() *redis.Client {
	if rdb == nil {
		panic("Redis client is not initialized. Call Setup() first.")
	}

	return rdb
}
