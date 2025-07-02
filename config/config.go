package config

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds all application configuration settings.
type Config struct {
	Environment string `mapstructure:"environment"`
	Server      struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
		JWTSecret    string        `mapstructure:"jwt_secret"`
	} `mapstructure:"server"`
	Database struct {
		DSN string `mapstructure:"dsn"` // Data Source Name for SQLite
	} `mapstructure:"database"`
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"` // "text" or "json"
	} `mapstructure:"log"`
	FileStorage struct {
		UploadDir    string   `mapstructure:"upload_dir"`
		MaxSizeMB    int      `mapstructure:"max_size_mb"`
		AllowedTypes []string `mapstructure:"allowed_types"`
	} `mapstructure:"file_storage"`
	AudioProcServiceURL string `mapstructure:"audio_proc_service_url"`
}

// LoadConfig loads configuration from file and environment variables.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("environment", "dev")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 5*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)
	v.SetDefault("database.dsn", "file:kindergarten.db?_foreign_keys=on")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json") // Default to JSON format
	v.SetDefault("file_storage.upload_dir", "uploads")
	v.SetDefault("file_storage.max_size_mb", 10)
	v.SetDefault("file_storage.allowed_types", []string{"audio/mpeg", "audio/wav", "audio/ogg"})
	v.SetDefault("audio_proc_service_url", "http://localhost:8000/analyze-audio")

	// Set config file name and path
	v.SetConfigName("config")   // name of config file (without extension)
	v.AddConfigPath("./config") // path to look for the config file in the current directory
	v.AddConfigPath(".")        // optionally look for config in the working directory

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error and proceed with defaults/env vars
			fmt.Println("No config file found, using defaults and environment variables.")
		} else {
			// Config file was found but another error was encountered
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Automatically read environment variables that match
	v.SetEnvPrefix("KINDERGARTEN") // prefix for environment variables (e.g., KINDERGARTEN_SERVER_PORT)
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validateConfig ensures all necessary settings are present and valid.
func validateConfig(cfg *Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server port cannot be 0")
	}
	if cfg.Server.JWTSecret == "" {
		return fmt.Errorf("server JWT secret cannot be empty")
	}
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database DSN cannot be empty")
	}
	if cfg.FileStorage.UploadDir == "" {
		return fmt.Errorf("file storage upload directory cannot be empty")
	}
	if cfg.FileStorage.MaxSizeMB <= 0 {
		return fmt.Errorf("file storage max size must be greater than 0")
	}
	if len(cfg.FileStorage.AllowedTypes) == 0 {
		return fmt.Errorf("file storage allowed types cannot be empty")
	}

	// Ensure upload directory exists
	if _, err := os.Stat(cfg.FileStorage.UploadDir); os.IsNotExist(err) {
		logrus.Infof("Creating upload directory: %s", cfg.FileStorage.UploadDir)
		if err := os.MkdirAll(cfg.FileStorage.UploadDir, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory %s: %w", cfg.FileStorage.UploadDir, err)
		}
	}

	return nil
}
