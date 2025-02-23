package cache

import (
	"context"
	"strconv"
	"strings"
	"time"
)

type CacheMetrics struct {
	Hits        int64
	Misses      int64
	LatencyMs   float64
	MemoryUsage int64
}

func (c *RedisClient) GetMetrics() *CacheMetrics {
	ctx := context.Background()
	info, err := c.Info(ctx).Result()
	if err != nil {
		return nil
	}

	metrics := &CacheMetrics{}

	for _, line := range strings.Split(info, "\r\n") {
		if strings.HasPrefix(line, "keyspace_hits:") {
			if v, err := strconv.ParseInt(strings.TrimPrefix(line, "keyspace_hits:"), 10, 64); err == nil {
				metrics.Hits = v
			}
		} else if strings.HasPrefix(line, "keyspace_misses:") {
			if v, err := strconv.ParseInt(strings.TrimPrefix(line, "keyspace_misses:"), 10, 64); err == nil {
				metrics.Misses = v
			}
		} else if strings.HasPrefix(line, "used_memory:") {
			if v, err := strconv.ParseInt(strings.TrimPrefix(line, "used_memory:"), 10, 64); err == nil {
				metrics.MemoryUsage = v
			}
		}
	}

	start := time.Now()
	c.Ping(ctx)
	metrics.LatencyMs = float64(time.Since(start).Milliseconds())

	return metrics
}
