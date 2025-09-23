package main

import (
	"context"
	"database/sql"
	"embed"
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
	"kitadoc-backend/services"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

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
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		}
	default:
		logrus.Fatalf("Unsupported log format: %s. Must be 'json' or 'text'.", cfg.Log.Format)
	}
	logger.InitGlobalLogger(logLevel, logFormatter)

	log := logger.GetGlobalLogger()
	log.Infof("Application starting in %s environment...", cfg.Environment)

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("Failed to close database connection: %v", err)
		}
	}()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Info("Successfully connected to the database!")

	// Check if the database schema is initialized
	err = data.MigrateDB(db, migrationFS)
	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	log.Info("Database schema is up to date.")

	// Initialize DAL
	dal := data.NewDAL(db)

	// Initialize App
	application := app.NewApplication(*cfg, dal)

	// Get UserService for pre-creating users
	userService := application.AuthHandler.UserService

	// Pre-create admin user if environment variables are set
	adminUsername := cfg.AdminUser.Username
	adminPassword := cfg.AdminUser.Password
	if adminUsername != "" && adminPassword != "" {
		_, err := userService.RegisterUser(log.GetLogrusEntry(), adminUsername, adminPassword, "admin")
		if err != nil && !errors.Is(err, services.ErrAlreadyExists) {
			log.Fatalf("Failed to pre-create admin user: %v", err)
		} else if errors.Is(err, services.ErrAlreadyExists) {
			log.Infof("Admin user '%s' already exists.", adminUsername)
		} else {
			log.Infof("Admin user '%s' created successfully.", adminUsername)
		}
	}

	// Pre-create normal user if environment variables are set
	normalUsername := cfg.NormalUser.Username
	normalPassword := cfg.NormalUser.Password
	if normalUsername != "" && normalPassword != "" {
		_, err := userService.RegisterUser(log.GetLogrusEntry(), normalUsername, normalPassword, "teacher")
		if err != nil && !errors.Is(err, services.ErrAlreadyExists) {
			log.Fatalf("Failed to pre-create normal user: %v", err)
		} else if errors.Is(err, services.ErrAlreadyExists) {
			log.Infof("Normal user '%s' already exists.", normalUsername)
		} else {
			log.Infof("Normal user '%s' created successfully.", normalUsername)
		}
	}

	// Set up routes
	routerWithMiddleware := application.Routes()

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      routerWithMiddleware,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Infof("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v", server.Addr, err)
		}
	}()

	<-done
	log.Info("Attempting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Info("Server gracefully shut down.")
}
