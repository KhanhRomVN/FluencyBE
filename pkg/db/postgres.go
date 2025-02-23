package db

import (
	"database/sql"
	"fluencybe/internal/core/config"
	"fluencybe/pkg/logger"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func enableRequiredExtensions(db *sql.DB) error {
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto")
	return err
}

func NewDBConnection(cfg config.DBConfig, log *logger.PrettyLogger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		log.Critical("DATABASE_CONNECTION_ERROR", map[string]interface{}{
			"error": err.Error(),
			"url":   cfg.URL,
		}, "Failed to connect to database")
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Critical("DATABASE_PING_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to ping database")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable required extensions
	if err := enableRequiredExtensions(db); err != nil {
		log.Critical("DATABASE_EXTENSION_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to enable required database extensions")
		return nil, fmt.Errorf("failed to enable required extensions: %w", err)
	}

	log.Info("DATABASE_CONNECTION_SUCCESS", map[string]interface{}{
		"max_open_conns": cfg.MaxOpenConns,
		"max_idle_conns": cfg.MaxIdleConns,
	}, "Database connection established successfully")

	return db, nil
}

func NewGormDBConnection(cfg config.DBConfig, log *logger.PrettyLogger) (*gorm.DB, error) {
	sqlDB, err := NewDBConnection(cfg, log)
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true, // Disable prepared statement caching
	}), &gorm.Config{
		PrepareStmt: false, // Disable prepared statement globally
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Critical("GORM_INIT_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to initialize GORM")
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	return gormDB, nil
}
