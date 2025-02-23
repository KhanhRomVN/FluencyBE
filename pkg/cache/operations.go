package cache

import (
	"context"
	constants "fluencybe/internal/core/constants"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	BatchSet(ctx context.Context, items map[string]string, expiration time.Duration) error
	Info(ctx context.Context, sections ...string) *redis.StringCmd
	GetMetrics() *CacheMetrics
	Keys(ctx context.Context, pattern string) ([]string, error)
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *RedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if expiration == 0 {
		expiration = constants.CacheDefaultTTL
	}
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisClient) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

func (c *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.Client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *RedisClient) BatchSet(ctx context.Context, items map[string]string, expiration time.Duration) error {
	pipe := c.Client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, key, value, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	iter := c.Client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}
