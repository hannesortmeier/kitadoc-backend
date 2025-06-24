package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"kitadoc-backend/app"
	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up structured logging
	logLevel, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logrus.Fatalf("Invalid log level in configuration: %v", err)
	}

	var logFormatter logrus.Formatter
	switch cfg.Log.Format {
	case "json":
		logFormatter = &logrus.JSONFormatter{}
	case "text":
		logFormatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}
	default:
		logrus.Fatalf("Unsupported log format: %s. Must be 'json' or 'text'.", cfg.Log.Format)
	}
	logger.InitGlobalLogger(logLevel, logFormatter)

	logger.GetGlobalLogger().Infof("Application starting in %s environment...", cfg.Environment)

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		logger.GetGlobalLogger().Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.GetGlobalLogger().Errorf("Failed to close database connection: %v", err)
		}
	}()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		logger.GetGlobalLogger().Fatalf("Failed to connect to database: %v", err)
	}
	logger.GetGlobalLogger().Info("Successfully connected to the database!")

	// Initialize DAL
	dal := data.NewDAL(db)

	// Initialize App
	application := app.NewApplication(*cfg, dal)

	// Set up routes
	application.Routes()

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      application.GetRouter(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.GetGlobalLogger().Infof("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.GetGlobalLogger().Fatalf("Could not listen on %s: %v", server.Addr, err)
		}
	}()

	<-done
	logger.GetGlobalLogger().Info("Attempting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.GetGlobalLogger().Fatalf("Server shutdown failed: %v", err)
	}
	logger.GetGlobalLogger().Info("Server gracefully shut down.")
}
