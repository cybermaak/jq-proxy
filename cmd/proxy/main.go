package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jq-proxy-service/internal/config"
	"jq-proxy-service/internal/proxy"
	"jq-proxy-service/internal/transform"
	"jq-proxy-service/pkg/client"

	"github.com/sirupsen/logrus"
)

func main() {
	var configPath = flag.String("config", "configs/config.json", "Path to configuration file")
	var port = flag.String("port", "", "Port to listen on (overrides config)")
	var logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Set log level
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Invalid log level")
	}
	logger.SetLevel(level)

	logger.WithField("config_path", *configPath).Info("Starting JQ Proxy Service")

	// Initialize configuration provider with environment variable support
	configProvider := config.NewEnvProvider(*configPath)
	proxyConfig, err := configProvider.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	logger.WithFields(logrus.Fields{
		"endpoints": len(proxyConfig.Endpoints),
		"port":      proxyConfig.Server.Port,
	}).Info("Configuration loaded successfully")

	// Override port if specified via command line
	if *port != "" {
		fmt.Sscanf(*port, "%d", &proxyConfig.Server.Port)
		logger.WithField("port", proxyConfig.Server.Port).Info("Port overridden by command line")
	}

	// Initialize HTTP client
	httpClient := client.NewClient(time.Duration(proxyConfig.Server.ReadTimeout) * time.Second)

	// Initialize unified transformer (supports jq)
	transformer := transform.NewUnifiedTransformer()

	// Initialize proxy service
	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)

	// Initialize HTTP handler
	handler := proxy.NewHandler(proxyService, logger)
	router := handler.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", proxyConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(proxyConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(proxyConfig.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", proxyConfig.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}
