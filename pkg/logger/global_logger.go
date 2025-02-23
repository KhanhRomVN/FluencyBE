package logger

import (
	"io"
	"os"
	"strings"
	"sync"
)

var (
	defaultLogger *PrettyLogger
	defaultOnce   sync.Once
)

// GlobalLoggerOptions contains configuration options for the global logger
type GlobalLoggerOptions struct {
	Level    LogLevel
	Output   io.Writer
	Service  string
	Colorful bool
}

// DefaultLoggerOptions returns the default options for the global logger
func DefaultLoggerOptions() GlobalLoggerOptions {
	return GlobalLoggerOptions{
		Level:    LevelInfo,
		Output:   os.Stdout,
		Service:  "MAIN",
		Colorful: true,
	}
}

// InitGlobalLogger initializes the global logger with custom options
func InitGlobalLogger(opts GlobalLoggerOptions) *PrettyLogger {
	defaultOnce.Do(func() {
		defaultLogger = &PrettyLogger{
			level:        opts.Level,
			output:       opts.Output,
			service:      opts.Service,
			colorful:     opts.Colorful,
			activeLevels: make(map[LogLevel]bool),
		}

		// Parse LOG_LEVELS environment variable
		logLevelsStr := os.Getenv("LOG_LEVELS")
		if logLevelsStr != "" {
			levels := strings.Split(strings.ToUpper(logLevelsStr), ",")
			for _, level := range levels {
				level = strings.TrimSpace(level)
				switch level {
				case "DEBUG":
					defaultLogger.activeLevels[LevelDebug] = true
				case "INFO":
					defaultLogger.activeLevels[LevelInfo] = true
				case "WARNING":
					defaultLogger.activeLevels[LevelWarning] = true
				case "ERROR":
					defaultLogger.activeLevels[LevelError] = true
				case "CRITICAL":
					defaultLogger.activeLevels[LevelCritical] = true
				}
			}
		} else {
			// If no specific levels are set, enable all levels up to the configured level
			for l := LevelDebug; l <= opts.Level; l++ {
				defaultLogger.activeLevels[l] = true
			}
		}
	})
	return defaultLogger
}

// GetGlobalLogger returns the global logger instance, initializing it with default options if not already initialized
func GetGlobalLogger() *PrettyLogger {
	if defaultLogger == nil {
		return InitGlobalLogger(DefaultLoggerOptions())
	}
	return defaultLogger
}

// WithService creates a new logger instance with a different service name but sharing the same configuration
func WithService(serviceName string) *PrettyLogger {
	parent := GetGlobalLogger()
	return &PrettyLogger{
		level:    parent.level,
		output:   parent.output,
		service:  serviceName,
		colorful: parent.colorful,
	}
}
