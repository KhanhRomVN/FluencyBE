package main

import (
	"context"
	"fluencybe/internal/core/config"
	"fluencybe/internal/core/constants"
	"fluencybe/internal/infrastructure/di"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create dependency injection container
	container, err := di.NewContainer(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      container.Router,
		WriteTimeout: constants.HTTPWriteTimeout,
		ReadTimeout:  constants.HTTPReadTimeout,
		IdleTimeout:  constants.HTTPIdleTimeout,
	}

	// Start server
	go func() {
		container.Logger.Info("SERVER_START", map[string]interface{}{
			"port":          cfg.Server.Port,
			"write_timeout": server.WriteTimeout,
			"read_timeout":  server.ReadTimeout,
		}, "Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			container.Logger.Critical("SERVER_ERROR", map[string]interface{}{
				"error": err.Error(),
				"port":  cfg.Server.Port,
			}, "Server failed to start")
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	container.Logger.Info("SHUTDOWN_SIGNAL", nil, "Received shutdown signal")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	container.Logger.Info("SERVER_SHUTDOWN", map[string]interface{}{
		"timeout": constants.ShutdownTimeout.String(),
	}, "Initiating graceful shutdown")

	if err := server.Shutdown(ctx); err != nil {
		container.Logger.Critical("SERVER_SHUTDOWN_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Server forced to shutdown")
		os.Exit(1)
	}

	container.Logger.Info("SERVER_SHUTDOWN", map[string]interface{}{
		"status": "completed",
	}, "Server shutdown successfully")
}
