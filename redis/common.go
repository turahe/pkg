package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// getClient returns the appropriate Redis client (standard or cluster)
func getClient() redis.Cmdable {
	return GetUniversalClient()
}

// String

func Get(key string) (string, error) {
	val, err := getClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
func Set(key string, value interface{}, expiration time.Duration) error {
	return getClient().Set(ctx, key, value, expiration).Err()
}
func Delete(key string) error {
	return getClient().Del(ctx, key).Err()
}
func MGet(keys ...string) ([]string, error) {
	result, err := getClient().MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	values := make([]string, len(result))
	for i, v := range result {
		if v != nil {
			values[i] = v.(string)
		}
	}
	return values, nil
}
func MSet(pairs map[string]interface{}) error {
	return getClient().MSet(ctx, pairs).Err()
}

// Hash

func HGet(key, field string) (string, error) {
	return getClient().HGet(ctx, key, field).Result()
}
func HGetAll(key string) (map[string]string, error) {
	return getClient().HGetAll(ctx, key).Result()
}
func HSet(key string, field string, value interface{}) error {
	return getClient().HSet(ctx, key, field, value).Err()
}
func HSetMap(key string, fields map[string]interface{}) error {
	return getClient().HSet(ctx, key, fields).Err()
}

// List

func LPush(key string, values ...interface{}) error {
	return getClient().LPush(ctx, key, values...).Err()
}
func RPop(key string) (string, error) {
	return getClient().RPop(ctx, key).Result()
}
func LRange(key string, start, stop int64) ([]string, error) {
	return getClient().LRange(ctx, key, start, stop).Result()
}

// Set

func SAdd(key string, members ...interface{}) error {
	return getClient().SAdd(ctx, key, members...).Err()
}
func SMembers(key string) ([]string, error) {
	return getClient().SMembers(ctx, key).Result()
}
func SRem(key string, members ...interface{}) error {
	return getClient().SRem(ctx, key, members...).Err()
}

// Lock

func AcquireLock(key string, value interface{}, expiration time.Duration) (bool, error) {
	return getClient().SetNX(ctx, key, value, expiration).Result()
}
func ExtendLock(key string, expiration time.Duration) error {
	return getClient().Expire(ctx, key, expiration).Err()
}
func ReleaseLock(key string) error {
	return getClient().Del(ctx, key).Err()
}

// Pipeline

func Pipeline(f func(pipe redis.Pipeliner)) error {
	var pipe redis.Pipeliner
	if isCluster {
		pipe = rdbCluster.Pipeline()
	} else {
		pipe = rdb.Pipeline()
	}
	f(pipe)
	_, err := pipe.Exec(ctx)
	return err
}
func PipelineSet(keyValues map[string]interface{}, expiration time.Duration) error {
	return Pipeline(func(pipe redis.Pipeliner) {
		for key, value := range keyValues {
			pipe.Set(ctx, key, value, expiration)
		}
	})
}

// Publish & Subscribe

func PublishMessage(channel, message string) error {
	return getClient().Publish(ctx, channel, message).Err()
}
func SubscribeToChannel(channel string, handler func(message string)) error {
	var sub *redis.PubSub
	if isCluster {
		sub = rdbCluster.Subscribe(ctx, channel)
	} else {
		sub = rdb.Subscribe(ctx, channel)
	}
	defer sub.Close()

	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			return err
		}

		handler(msg.Payload)
	}
}

// Scan

func ScanKeys(pattern string, count int64) ([]string, error) {
	if isCluster {
		// For cluster mode, scan all nodes
		return scanClusterKeys(pattern, count)
	}

	// Standard mode
	cursor := uint64(0)
	var keys []string

	for {
		var newKeys []string
		var err error

		newKeys, cursor, err = rdb.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, newKeys...)
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func scanClusterKeys(pattern string, count int64) ([]string, error) {
	var allKeys []string
	err := rdbCluster.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		cursor := uint64(0)
		for {
			var keys []string
			var err error
			keys, cursor, err = shard.Scan(ctx, cursor, pattern, count).Result()
			if err != nil {
				return err
			}
			allKeys = append(allKeys, keys...)
			if cursor == 0 {
				break
			}
		}
		return nil
	})
	return allKeys, err
}

// Save

func Save() error {
	return getClient().Save(ctx).Err()
}
func BGSave() error {
	return getClient().BgSave(ctx).Err()
}
