package di

import (
	"database/sql"
	"fluencybe/internal/core/config"
	"fluencybe/internal/core/status"
	"fluencybe/internal/infrastructure/discord"
	"fluencybe/internal/infrastructure/metrics"
	"fluencybe/internal/infrastructure/router"
	"fluencybe/pkg/cache"
	"fluencybe/pkg/db"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/search"

	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go/v2"
	"gorm.io/gorm"
)

func initializeDB(cfg *config.Config, log *logger.PrettyLogger) (*sql.DB, error) {
	return db.NewDBConnection(cfg.DBConfig, log)
}

func initializeGormDB(cfg *config.Config, log *logger.PrettyLogger) (*gorm.DB, error) {
	return db.NewGormDBConnection(cfg.DBConfig, log)
}

func initializeRedis(cfg *config.Config, log *logger.PrettyLogger) (*cache.RedisClient, error) {
	return cache.NewRedisClient(cfg.RedisConfig, log)
}

func initializeOpenSearch(cfg *config.Config, log *logger.PrettyLogger) (*opensearch.Client, error) {
	return search.NewOpenSearchClient(cfg.OpenSearchConfig, log)
}

func initializeDiscordBot(log *logger.PrettyLogger) (*discord.Bot, error) {
	bot, err := discord.NewBot(log)
	if err != nil {
		return nil, err
	}

	if err := bot.Start(); err != nil {
		return nil, err
	}

	return bot, nil
}

func initializeMetrics(log *logger.PrettyLogger, redisClient *cache.RedisClient) *metrics.Metrics {
	metricsCollector := metrics.NewMetrics(log, redisClient)
	metricsCollector.StartMetricsCollection()
	return metricsCollector
}

func initializeRouter(dbConn *sql.DB) *gin.Engine {
	r := router.NewRouter(dbConn)
	return r.Engine
}

func initializeHealthCheck(redisClient *cache.RedisClient, openSearchClient *opensearch.Client, log *logger.PrettyLogger) {
	status.InitConnectionStatus(log)
	status.StartHealthCheck(redisClient, openSearchClient)
}
