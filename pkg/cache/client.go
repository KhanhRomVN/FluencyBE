package cache

import (
	"context"
	"crypto/tls"
	"fluencybe/internal/core/config"
	constants "fluencybe/internal/core/constants"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
	metrics *CacheMetrics
}

func NewRedisClient(cfg config.RedisConfig, log *logger.PrettyLogger) (*RedisClient, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       constants.DefaultRedisDB,

		DialTimeout:  constants.HTTPDialTimeout,
		ReadTimeout:  constants.HTTPReadTimeout,
		WriteTimeout: constants.HTTPWriteTimeout,

		PoolSize:     constants.DefaultRedisPoolSize,
		MinIdleConns: constants.DefaultRedisMinIdleConns,
		MaxRetries:   constants.DefaultRedisMaxRetries,

		MinRetryBackoff: constants.RateLimitMinBackoff,
		MaxRetryBackoff: constants.RateLimitMaxBackoff,
	}

	if cfg.SSL == "true" {
		opts.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.AllowInsecureTLS,
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Critical("REDIS_CONNECTION_ERROR", map[string]interface{}{
			"error": err.Error(),
			"host":  cfg.Host,
			"port":  cfg.Port,
		}, "Failed to connect to Redis, system will continue without caching")
		return &RedisClient{Client: client}, nil
	}

	log.Info("REDIS_CONNECTION_SUCCESS", map[string]interface{}{
		"host": cfg.Host,
		"port": cfg.Port,
	}, "Redis connection established successfully")

	return &RedisClient{Client: client}, nil
}

func (c *RedisClient) Info(ctx context.Context, sections ...string) *redis.StringCmd {
	return c.Client.Info(ctx, sections...)
}
