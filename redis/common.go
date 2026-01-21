package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// String

func Get(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
func Set(key string, value interface{}, expiration time.Duration) error {
	return rdb.Set(ctx, key, value, expiration).Err()
}
func Delete(key string) error {
	return rdb.Del(ctx, key).Err()
}
func MGet(keys ...string) ([]string, error) {
	result, err := rdb.MGet(ctx, keys...).Result()
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
	return rdb.MSet(ctx, pairs).Err()
}

// Hash

func HGet(key, field string) (string, error) {
	return rdb.HGet(ctx, key, field).Result()
}
func HGetAll(key string) (map[string]string, error) {
	return rdb.HGetAll(ctx, key).Result()
}
func HSet(key string, field string, value interface{}) error {
	return rdb.HSet(ctx, key, field, value).Err()
}
func HSetMap(key string, fields map[string]interface{}) error {
	return rdb.HSet(ctx, key, fields).Err()
}

// List

func LPush(key string, values ...interface{}) error {
	return rdb.LPush(ctx, key, values...).Err()
}
func RPop(key string) (string, error) {
	return rdb.RPop(ctx, key).Result()
}
func LRange(key string, start, stop int64) ([]string, error) {
	return rdb.LRange(ctx, key, start, stop).Result()
}

// Set

func SAdd(key string, members ...interface{}) error {
	return rdb.SAdd(ctx, key, members...).Err()
}
func SMembers(key string) ([]string, error) {
	return rdb.SMembers(ctx, key).Result()
}
func SRem(key string, members ...interface{}) error {
	return rdb.SRem(ctx, key, members...).Err()
}

// Lock

func AcquireLock(key string, value interface{}, expiration time.Duration) (bool, error) {
	return rdb.SetNX(ctx, key, value, expiration).Result()
}
func ExtendLock(key string, expiration time.Duration) error {
	return rdb.Expire(ctx, key, expiration).Err()
}
func ReleaseLock(key string) error {
	return rdb.Del(ctx, key).Err()
}

// Pipeline

func Pipeline(f func(pipe redis.Pipeliner)) error {
	pipe := rdb.Pipeline()
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
	return rdb.Publish(ctx, channel, message).Err()
}
func SubscribeToChannel(channel string, handler func(message string)) error {
	sub := rdb.Subscribe(ctx, channel)
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

// Save

func Save() error {
	return rdb.Save(ctx).Err()
}
func BGSave() error {
	return rdb.BgSave(ctx).Err()
}
