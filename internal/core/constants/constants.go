package constants

import (
	"errors"
	"time"
)

const (
	ServiceName     = "MAIN"
	ShutdownTimeout = 15 * time.Second

	// Environment variables
	EnvDatabaseURL = "DATABASE_URL"
	EnvServerPort  = "SERVER_PORT"
	EnvJWTSecret   = "JWT_SECRET"
	EnvRedisHost   = "REDIS_HOST"

	// JWT settings
	JWTAccessTokenTTL  = 7 * 24 * time.Hour // Token hết hạn sau 24h
	JWTRefreshTokenTTL = 7 * 24 * time.Hour // Refresh token hết hạn sau 7 ngày
	JWTIssuer          = "FluencyBE"
	JWTAlgorithm       = "HS256"

	// Configuration defaults
	DefaultDBMaxPoolSize     = 10
	DefaultServerPort        = "8080"
	DefaultDBMaxOpenConns    = 10
	DefaultDBMaxIdleConns    = 5
	DefaultRedisDB           = 0
	DefaultRedisPoolSize     = 10
	DefaultRedisMinIdleConns = 5
	DefaultRedisMaxRetries   = 1

	// Cache TTLs
	CacheDefaultTTL = 24 * time.Hour
	CacheShortTTL   = 5 * time.Minute
	CacheVersionTTL = 5 * time.Minute
	CacheSearchTTL  = 1 * time.Minute

	// HTTP timeouts
	HTTPWriteTimeout = 15 * time.Second
	HTTPReadTimeout  = 15 * time.Second
	HTTPIdleTimeout  = 60 * time.Second
	HTTPDialTimeout  = 2 * time.Second

	// Rate limiting
	RateLimitRequests   = 1000
	RateLimitDuration   = time.Minute
	RateLimitMinBackoff = 8 * time.Millisecond
	RateLimitMaxBackoff = 512 * time.Millisecond

	// Log fields
	LogComponentKey = "component"
	LogVersionKey   = "version"

	// User roles
	RoleUser      = "user"
	RoleDeveloper = "developer"

	// Health check intervals
	HealthCheckInterval = 10 * time.Second
	HealthCheckTimeout  = 5 * time.Second

	// Metrics collection interval
	MetricsCollectionInterval = 10 * time.Second

	// OpenSearch settings
	OpenSearchBulkSize = 1000
	OpenSearchTimeout  = 30 * time.Second
	OpenSearchRetries  = 3

	// Password settings
	MinPasswordLength = 8
	MaxPasswordLength = 100

	// Username settings
	MinUsernameLength = 3
	MaxUsernameLength = 50

	// Field length limits - Common
	MaxTopicLength       = 100
	MaxInstructionLength = 1000
	MaxTranscriptLength  = 5000
	MaxAudioURLs         = 10
	MaxImageURLs         = 10
	MinMaxTime           = 30   // 30 seconds
	MaxMaxTime           = 3600 // 1 hour

	// Field length limits - Reading specific
	MaxTitleLength       = 200
	MaxPassages          = 10
	MaxPassageLength     = 5000
	MaxQuestionLength    = 500
	MaxAnswerLength      = 500
	MaxOptionsLength     = 500
	MaxExplanationLength = 1000

	// Field length limits - Listening specific
	MaxListeningQuestionLength    = 500
	MaxListeningAnswerLength      = 500
	MaxListeningOptionsLength     = 500
	MaxListeningExplanationLength = 1000
	MaxMapLabellingQuestions      = 20
	MaxMatchingPairs              = 10
)

type ContextKey string

const (
	GinContextKey ContextKey = "GinContextKey"
)

// Authentication errors
var (
	ErrAuthHeaderRequired = errors.New("authorization header is required")
	ErrInvalidAuthFormat  = errors.New("invalid authorization format")
)
