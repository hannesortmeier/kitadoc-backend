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

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"kitadoc-backend/app"
	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/migrations"
	"kitadoc-backend/models"
)

var (
	application       *app.Application
	ts                *httptest.Server
	db                *sql.DB
	mockTranscription *httptest.Server
	mockLLMAnalysis   *httptest.Server
)

func TestMain(m *testing.M) {
	// Create a mock server for the transcription service
	mockTranscription = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep 1 second to simulate processing time
		time.Sleep(1 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode("Anna spielt mit Tom Fangeln im Sandkasten."); err != nil {
			panic(fmt.Sprintf("failed to encode transcription response: %v", err))
		}
	}))
	defer mockTranscription.Close()

	// Create a mock server for the llm analysis service
	mockLLMAnalysis = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep 1 second to simulate processing time
		time.Sleep(1 * time.Second)
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
			panic(fmt.Sprintf("failed to encode analysis response: %v", err))
		}
	}))
	defer mockLLMAnalysis.Close()

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
	// Create a temporary file for the SQLite database so tests use a real file-backed DB
	tmpDBFile, err := os.CreateTemp("", "kitadoc_test_*.db")
	if err != nil {
		panic(fmt.Sprintf("failed to create temporary test database file: %v", err))
	}
	// Close the file descriptor; SQLite will open it by path.
	tmpDBFile.Close() // nolint:errcheck
	// Ensure the temporary database file is removed after tests
	defer func() {
		if err := os.Remove(tmpDBFile.Name()); err != nil {
			fmt.Printf("failed to remove temporary test database file: %v\n", err)
		}
	}()

	// Initialize configuration for testing with a file-backed SQLite database
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
			DSN           string `mapstructure:"dsn"`
			EncryptionKey string `mapstructure:"encryption_key"`
		}{
			DSN:           "file:" + tmpDBFile.Name() + "?_pragma=foreign_keys(1)", // Use file-backed DB in tmp
			EncryptionKey: "0123456789abcdef0123456789abcdef",
		},
		FileStorage: struct {
			MaxSizeMB    int      `mapstructure:"max_size_mb"`
			AllowedTypes []string `mapstructure:"allowed_types"`
		}{
			MaxSizeMB:    10, // Set a small limit for testing
			AllowedTypes: []string{"audio/mpeg", "audio/wav", "audio/ogg", "application/octet-stream"},
		},
		TranscriptionServiceURL: mockTranscription.URL,
		LLMAnalysisServiceURL:   mockLLMAnalysis.URL,
	}

	logLevel, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(logLevel, &logrus.TextFormatter{FullTimestamp: true})

	// Initialize the database connection directly
	db, err = sql.Open("sqlite", cfg.Database.DSN)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}
	defer db.Close() //nolint:errcheck

	db.SetMaxOpenConns(1)

	// Run migrations
	if err := data.MigrateDB(db, migrations.Files); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	// Initialize DAL
	dal := data.NewDAL(db, []byte(cfg.Database.EncryptionKey))

	// Pre-create users for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	adminUser := &models.User{Username: "admin", PasswordHash: string(hashedPassword), Role: "admin"}
	_, err = dal.Users.Create(adminUser)
	if err != nil {
		panic(fmt.Sprintf("failed to create admin user: %v", err))
	}

	normalUser := &models.User{Username: "testuser", PasswordHash: string(hashedPassword), Role: "teacher"}
	_, err = dal.Users.Create(normalUser)
	if err != nil {
		panic(fmt.Sprintf("failed to create normal user: %v", err))
	}

	// Insert default kita masterdata
	defaultMasterdata := &models.KitaMasterdata{
		Name:        "Test Kita",
		Street:      "Test Str",
		HouseNumber: "1",
		PostalCode:  "12345",
		City:        "Test City",
		PhoneNumber: "123456",
		Email:       "test@example.com",
	}
	if err := dal.KitaMasterdata.Update(defaultMasterdata); err != nil {
		panic(fmt.Sprintf("failed to create default kita masterdata: %v", err))
	}

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
