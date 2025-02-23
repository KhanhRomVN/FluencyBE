package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig         DBConfig
	Server           ServerConfig
	JWTSecret        string
	RedisConfig      RedisConfig
	OpenSearchConfig OpenSearchConfig
}

type DBConfig struct {
	URL          string
	MaxPoolSize  int
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host             string
	Port             string
	Username         string
	Password         string
	DB               int
	SSL              string
	AllowInsecureTLS bool
}

type OpenSearchConfig struct {
	Host               string
	Port               string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

type ServerConfig struct {
	Port string
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := parseInt(value)
	if err != nil {
		log.Printf("Error parsing %s as int, using default value: %v", key, err)
		return defaultValue
	}
	return intValue
}

func parseInt(value string) (int, error) {
	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	return intValue, err
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.DBConfig.URL == "" {
		return errors.New("database URL is required")
	}
	if c.Server.Port == "" {
		return errors.New("server port is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT secret is required")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		DBConfig: DBConfig{
			URL:          getEnvWithDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/fluency?sslmode=disable"),
			MaxPoolSize:  getEnvAsInt("DB_MAX_POOL_SIZE", 10),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		Server: ServerConfig{
			Port: getEnvWithDefault("SERVER_PORT", "8080"),
		},
		JWTSecret: getEnvWithDefault("JWT_SECRET", "default-development-secret"),
		RedisConfig: RedisConfig{
			Host:     getEnvWithDefault("REDIS_HOST", "localhost"),
			Port:     getEnvWithDefault("REDIS_PORT", "6379"),
			Username: getEnvWithDefault("REDIS_USERNAME", ""),
			Password: getEnvWithDefault("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			SSL:      getEnvWithDefault("REDIS_SSL", ""),
		},
		OpenSearchConfig: OpenSearchConfig{
			Host:               getEnvWithDefault("OPENSEARCH_HOST", "localhost"),
			Port:               getEnvWithDefault("OPENSEARCH_PORT", "9200"),
			Username:           getEnvWithDefault("OPENSEARCH_USERNAME", "admin"),
			Password:           getEnvWithDefault("OPENSEARCH_PASSWORD", "admin"),
			InsecureSkipVerify: os.Getenv("OPENSEARCH_INSECURE_SKIP_VERIFY") == "true",
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}
