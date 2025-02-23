package status

import (
	"context"
	"sync"
	"time"

	constants "fluencybe/internal/core/constants"

	"fluencybe/pkg/cache"
	"fluencybe/pkg/logger"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

var (
	RedisStatus      bool
	OpenSearchStatus bool
	mu               sync.RWMutex
	log              *logger.PrettyLogger
)

func InitConnectionStatus(logger *logger.PrettyLogger) {
	log = logger
}

func SetRedisStatus(status bool) {
	mu.Lock()
	defer mu.Unlock()
	RedisStatus = status
}

func SetOpenSearchStatus(status bool) {
	mu.Lock()
	defer mu.Unlock()
	OpenSearchStatus = status
}

func GetRedisStatus() bool {
	mu.RLock()
	defer mu.RUnlock()
	return RedisStatus
}

func GetOpenSearchStatus() bool {
	mu.RLock()
	defer mu.RUnlock()
	return OpenSearchStatus
}

func StartHealthCheck(redisClient cache.Cache, openSearchClient *opensearch.Client) {
	go checkRedisHealth(redisClient)
	go checkOpenSearchHealth(openSearchClient)
}

func checkRedisHealth(client cache.Cache) {
	ticker := time.NewTicker(constants.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.Get(ctx, "health_check")

		cancel()

		isHealthy := err == nil || err.Error() == "redis: nil"
		SetRedisStatus(isHealthy)

		if !isHealthy {
			log.Error("redis_health_check", map[string]interface{}{
				"error": err.Error(),
			}, "Redis health check failed")
		}
	}
}

func checkOpenSearchHealth(client *opensearch.Client) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		req := opensearchapi.ClusterHealthRequest{}
		res, err := req.Do(context.Background(), client)

		isHealthy := err == nil && res != nil && res.StatusCode == 200
		SetOpenSearchStatus(isHealthy)

		if !isHealthy {
			log.Error("opensearch_health_check", map[string]interface{}{
				"error": err.Error(),
			}, "OpenSearch health check failed")
		}
		if res != nil {
			res.Body.Close()
		}
	}
}
