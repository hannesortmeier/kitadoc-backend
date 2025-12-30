package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// Config holds all application configuration settings.
type Config struct {
	Environment string `mapstructure:"environment"`
	AdminUser   struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"admin_user"`
	NormalUser struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"normal_user"`
	Server struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
		JWTSecret    string        `mapstructure:"jwt_secret"`
	} `mapstructure:"server"`
	Database struct {
		DSN           string `mapstructure:"dsn"` // Data Source Name for SQLite
		EncryptionKey string `mapstructure:"encryption_key"`
	} `mapstructure:"database"`
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"` // "text" or "json"
	} `mapstructure:"log"`
	FileStorage struct {
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
	v.SetDefault("server.port", 8070)
	v.SetDefault("server.read_timeout", 5*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)
	v.SetDefault("database.dsn", "file:test.db?_pragma=foreign_keys(1)")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json") // Default to JSON format
	v.SetDefault("file_storage.upload_dir", "uploads")
	v.SetDefault("file_storage.max_size_mb", 10)
	v.SetDefault("file_storage.allowed_types", []string{"audio/mpeg", "audio/wav"})
	v.SetDefault("audio_proc_service_url", "http://127.0.0.1:8000/analyze-audio")

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
	if err := v.BindEnv("server.port", "KINDERGARTEN_SERVER_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_SERVER_PORT: %w", err)
	}
	if err := v.BindEnv("server.read_timeout", "KINDERGARTEN_SERVER_READ_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_SERVER_READ_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("server.write_timeout", "KINDERGARTEN_SERVER_WRITE_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_SERVER_WRITE_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("server.idle_timeout", "KINDERGARTEN_SERVER_IDLE_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_SERVER_IDLE_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("server.jwt_secret", "KINDERGARTEN_SERVER_JWT_SECRET"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_SERVER_JWT_SECRET: %w", err)
	}
	if err := v.BindEnv("database.dsn", "KINDERGARTEN_DATABASE_DSN"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_DATABASE_DSN: %w", err)
	}
	if err := v.BindEnv("database.encryption_key", "KINDERGARTEN_DATABASE_ENCRYPTION_KEY"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_DATABASE_ENCRYPTION_KEY: %w", err)
	}
	if err := v.BindEnv("log.level", "KINDERGARTEN_LOG_LEVEL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_LOG_LEVEL: %w", err)
	}
	if err := v.BindEnv("log.format", "KINDERGARTEN_LOG_FORMAT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_LOG_FORMAT: %w", err)
	}
	if err := v.BindEnv("file_storage.upload_dir", "KINDERGARTEN_FILE_STORAGE_UPLOAD_DIR"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_FILE_STORAGE_UPLOAD_DIR: %w", err)
	}
	if err := v.BindEnv("file_storage.max_size_mb", "KINDERGARTEN_FILE_STORAGE_MAX_SIZE_MB"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_FILE_STORAGE_MAX_SIZE_MB: %w", err)
	}
	if err := v.BindEnv("file_storage.allowed_types", "KINDERGARTEN_FILE_STORAGE_ALLOWED_TYPES"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_FILE_STORAGE_ALLOWED_TYPES: %w", err)
	}
	if err := v.BindEnv("audio_proc_service_url", "KINDERGARTEN_AUDIO_PROC_SERVICE_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_AUDIO_PROC_SERVICE_URL: %w", err)
	}
	if err := v.BindEnv("admin_user.username", "KINDERGARTEN_ADMIN_USERNAME"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_ADMIN_USERNAME: %w", err)
	}
	if err := v.BindEnv("admin_user.password", "KINDERGARTEN_ADMIN_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_ADMIN_PASSWORD: %w", err)
	}
	if err := v.BindEnv("normal_user.username", "KINDERGARTEN_NORMAL_USERNAME"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_NORMAL_USERNAME: %w", err)
	}
	if err := v.BindEnv("normal_user.password", "KINDERGARTEN_NORMAL_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind env var KINDERGARTEN_NORMAL_PASSWORD: %w", err)
	}

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
	if cfg.Database.EncryptionKey == "" {
		return fmt.Errorf("database encryption key cannot be empty")
	}
	if len(cfg.Database.EncryptionKey) != 32 {
		return fmt.Errorf("database encryption key must be 32 bytes long")
	}
	if cfg.FileStorage.MaxSizeMB <= 0 {
		return fmt.Errorf("file storage max size must be greater than 0")
	}
	if len(cfg.FileStorage.AllowedTypes) == 0 {
		return fmt.Errorf("file storage allowed types cannot be empty")
	}

	return nil
}
