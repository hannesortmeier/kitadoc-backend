package e2e_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"kitadoc-backend/app"
	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
)

var (
	application *app.Application
	ts          *httptest.Server
	db          *sql.DB
)

func TestMain(m *testing.M) {
	// Create a mock server for the audio-proc service
	mockAudioProc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"number_of_entries": 1,
			"analysis_results": []map[string]interface{}{
				{
					"child_id":              2,
					"first_name":            "Anna",
					"last_name":             "Müller",
					"transcription_summary": "Anna spielt mit Tom Fangeln im Sandkasten.",
					"analysis_category": map[string]interface{}{
						"category_id":   2,
						"category_name": "Soziale Entwicklung",
					},
				},
			},
		}); err != nil {
			panic(fmt.Sprintf("failed to encode response: %v", err))
		}
	}))
	defer mockAudioProc.Close()

	// Create temporary directory for uploads
	if err := os.MkdirAll("test_uploads", os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create test uploads directory: %v", err))
	}
	// Ensure the test uploads directory is cleaned up after tests
	defer func() {
		if err := os.RemoveAll("test_uploads"); err != nil {
			fmt.Printf("failed to remove test uploads directory: %v\n", err)
		}
	}()
	// Initialize configuration for testing with an in-memory SQLite database
	cfg := config.Config{
		Environment: "test",
		Server: struct {
			Port         int           `mapstructure:"port"`
			ReadTimeout  time.Duration `mapstructure:"read_timeout"`
			WriteTimeout time.Duration `mapstructure:"write_timeout"`
			IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
			JWTSecret    string        `mapstructure:"jwt_secret"`
		}{
			Port:      8080,
			JWTSecret: "test_jwt_secret_very_long_and_secure_key_for_testing_purposes",
		},
		Database: struct {
			DSN string `mapstructure:"dsn"`
		}{
			DSN: ":memory:?_foreign_keys=on", // Use in-memory database for testing
		},
		FileStorage: struct {
			MaxSizeMB    int      `mapstructure:"max_size_mb"`
			AllowedTypes []string `mapstructure:"allowed_types"`
		}{
			MaxSizeMB:    10, // Set a small limit for testing
			AllowedTypes: []string{"audio/mpeg", "audio/wav", "audio/ogg", "application/octet-stream"},
		},
		AudioProcServiceURL: mockAudioProc.URL,
	}

	logLevel, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(logLevel, &logrus.TextFormatter{FullTimestamp: true})

	var err error
	// Initialize the database connection directly
	db, err = sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}
	defer db.Close() //nolint:errcheck

	// Initialize DAL
	dal := data.NewDAL(db)

	// Initialize the application with test configuration and DAL
	application = app.NewApplication(cfg, dal)

	// Set up routes explicitly
	application.Routes()

	// Start the test server with the router that has routes set up
	ts = httptest.NewServer(application.GetRouter())
	defer ts.Close()

	// Run tests
	code := m.Run()

	os.Exit(code)
}
