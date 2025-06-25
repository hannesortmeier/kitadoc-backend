package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
)

func TestMain(m *testing.M) {
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
			DSN: ":memory:", // Use in-memory database for testing
		},
		FileStorage: struct {
			UploadDir    string   `mapstructure:"upload_dir"`
			MaxSizeMB    int      `mapstructure:"max_size_mb"`
			AllowedTypes []string `mapstructure:"allowed_types"`
		}{
			MaxSizeMB:    10, // Set a small limit for testing
			AllowedTypes: []string{"audio/mpeg", "audio/wav", "audio/ogg", "application/octet-stream"},
			UploadDir:    "test_uploads", // Use a test directory for uploads
		},
	}

	logLevel, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(logLevel, &logrus.TextFormatter{FullTimestamp: true})

	// Initialize the database connection directly
	db, err := sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}
	defer db.Close() //nolint:errcheck

	// Read and execute data_model.sql
	sqlContent, err := os.ReadFile("../database/data_model.sql")
	if err != nil {
		panic(fmt.Sprintf("failed to read data_model.sql: %v", err))
	}

	_, err = db.Exec(string(sqlContent))
	if err != nil {
		panic(fmt.Sprintf("failed to execute data_model.sql: %v", err))
	}

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

// Helper function to make authenticated requests
func makeAuthenticatedRequest(t *testing.T, method, url, token string, body interface{}, contentType string) *http.Response {
	reqBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(method, ts.URL+url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	return resp
}

// Helper function to make unauthenticated requests
func makeUnauthenticatedRequest(t *testing.T, method, url string, body interface{}, contentType string) *http.Response {
	reqBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(method, ts.URL+url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	return resp
}

// Helper function to read response body
func readResponseBody(t *testing.T, resp *http.Response) []byte {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return body
}

func TestPublicRoutes(t *testing.T) {
	// Test POST /api/v1/auth/register
	t.Run("Register User", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
			"username": "testuser",
			"password": "password123",
			"role":     "teacher",
		}, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, resp.StatusCode, string(body))
		} else {
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("testuser")) {
				t.Errorf("Expected success message, got %s", body)
			}
		}
	})

	t.Run("Register Admin User", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
			"username": "adminuser",
			"password": "password123",
			"role":     "admin",
		}, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, resp.StatusCode, string(body))
		} else {
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("adminuser")) {
				t.Errorf("Expected success message, got %s", body)
			}
		}
	})

	// Test POST /api/v1/auth/login
	var authToken string
	t.Run("Login User", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"username": "testuser", // Added username field to match registration
			"email":    "testuser@example.com",
			"password": "password123",
		}, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		} else {
			var loginResp struct {
				Token string `json:"token"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &loginResp); err != nil {
				t.Fatalf("Failed to unmarshal login response: %v. Body: %s", err, string(body))
			}
			authToken = loginResp.Token
			if authToken == "" {
				t.Error("Expected auth token, got empty string")
			}
		}
	})

	var adminAuthToken string
	t.Run("Login Admin User", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"username": "adminuser",
			"password": "password123",
		}, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		} else {
			var loginResp struct {
				Token string `json:"token"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &loginResp); err != nil {
				t.Fatalf("Failed to unmarshal login response: %v. Body: %s", err, string(body))
			}
			adminAuthToken = loginResp.Token
			if adminAuthToken == "" {
				t.Error("Expected auth token, got empty string")
			}
		}
	})

	// Test GET /health
	t.Run("Health Check", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodGet, "/health", nil, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
		body := readResponseBody(t, resp)
		// Update to check for {"status":"ok"} format instead of just "OK"
		if !bytes.Contains(body, []byte("status")) || !bytes.Contains(body, []byte("ok")) {
			t.Errorf("Expected status ok in response, got %s", body)
		}
	})

	// Skip the rest of the tests if we don't have a valid auth token
	if authToken == "" {
		t.Skip("Skipping authenticated tests due to login failure")
	}

	// Auth Endpoints
	t.Run("Auth Endpoints", func(t *testing.T) {
		// Test GET /api/v1/auth/me
		t.Run("Get Current User", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/auth/me", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("testuser")) {
					t.Errorf("Expected user information in response, got %s", body)
				}
			}
		})

		// Test POST /api/v1/auth/logout
		t.Run("Logout User", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/auth/logout", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("success")) && !bytes.Contains(body, []byte("logged out")) {
					t.Errorf("Expected logout success message, got %s", body)
				}
			}
		})
	})

	// Need to login again after logout
	t.Run("Login Again After Logout", func(t *testing.T) {
		resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"username": "testuser",
			"email":    "testuser@example.com",
			"password": "password123",
		}, "application/json")
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode == http.StatusOK {
			var loginResp struct {
				Token string `json:"token"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &loginResp); err == nil {
				authToken = loginResp.Token
			}
		}

		if authToken == "" {
			t.Skip("Skipping remaining tests due to relogin failure")
		}
	})

	// Children Management Endpoints
	t.Run("Children Management Endpoints", func(t *testing.T) {
		var childID int
		// Test POST /api/v1/children
		t.Run("Create Child", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
				"birthdate":  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				"gender":     "female",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck

			body := readResponseBody(t, resp)
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, resp.StatusCode, string(body))
				t.Skip("Skipping dependent child tests")
			}

			var childResp struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(body, &childResp); err != nil {
				t.Fatalf("Failed to unmarshal child creation response: %v. Body: %s", err, string(body))
				t.Skip("Skipping dependent child tests")
			}
			childID = childResp.ID
			if childID == 0 {
				t.Error("Expected child ID, got empty string")
				t.Skip("Skipping dependent child tests")
			}
		})

		// Only run dependent tests if we have a child ID
		if childID == 0 {
			t.Skip("Skipping child tests because child creation failed")
		}

		// Test GET /api/v1/children
		t.Run("Get All Children", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/children", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("John")) {
					t.Errorf("Expected child name in response, got %s", body)
				}
			}
		})

		// Test GET /api/v1/children/{child_id}
		t.Run("Get Child By ID", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/children/%d", childID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("John")) {
					t.Errorf("Expected child name in response, got %s", body)
				}
			}
		})

		// Test PUT /api/v1/children/{child_id}
		t.Run("Update Child", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/children/%d", childID), authToken, map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
				"birthdate":  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				"gender":     "female",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("Child updated successfully")) {
					t.Errorf("Expected updated child name in response, got %s", body)
				}
			}
		})

		// Test DELETE /api/v1/children/{child_id}
		t.Run("Delete Child", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/children/%d", childID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
			} else {
				body := readResponseBody(t, resp)
				if !bytes.Contains(body, []byte("success")) && !bytes.Contains(body, []byte("deleted")) {
					t.Errorf("Expected delete success message, got %s", body)
				}
			}
		})
	})

	// Teachers Management Endpoints
	t.Run("Teachers Management Endpoints", func(t *testing.T) {
		var teacherID int
		// Test POST /api/v1/teachers
		t.Run("Create Teacher", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/teachers", adminAuthToken, map[string]string{
				"first_name": "Alice",
				"last_name":  "Smith",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
			}
			var teacherResp struct {
				ID int `json:"id"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &teacherResp); err != nil {
				t.Fatalf("Failed to unmarshal teacher creation response: %v", err)
			}

			teacherID = teacherResp.ID
			if teacherID == 0 {
				t.Error("Expected teacher ID, got 0")
			}
		})

		// Test GET /api/v1/teachers
		t.Run("Get All Teachers", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/teachers", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Alice")) {
				t.Errorf("Expected teacher name in response, got %s", body)
			}
		})

		// Test GET /api/v1/teachers/{teacher_id}
		t.Run("Get Teacher By ID", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/teachers/%d", teacherID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Alice")) {
				t.Errorf("Expected teacher name in response, got %s", body)
			}
		})

		// Test PUT /api/v1/teachers/{teacher_id}
		t.Run("Update Teacher", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/teachers/%d", teacherID), adminAuthToken, map[string]string{
				"first_name": "Alicia",
				"last_name":  "Smith",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Teacher updated successfully")) {
				t.Errorf("Expected updated teacher name in response, got %s", body)
			}
		})

		// Test DELETE /api/v1/teachers/{teacher_id}
		t.Run("Delete Teacher", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/teachers/%d", teacherID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Teacher deleted successfully")) {
				t.Errorf("Expected delete success message, got %s", body)
			}
		})
	})

	// Groups Management Endpoints
	t.Run("Groups Management Endpoints", func(t *testing.T) {
		var groupID int
		// Test POST /api/v1/groups
		t.Run("Create Group", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/groups", adminAuthToken, map[string]string{
				"name": "Group A",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
			}
			var groupResp struct {
				ID int `json:"id"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &groupResp); err != nil {
				t.Fatalf("Failed to unmarshal group creation response: %v", err)
			}
			groupID = groupResp.ID
			if groupID == 0 {
				t.Error("Expected group ID, got 0")
			}
		})

		// Test GET /api/v1/groups
		t.Run("Get All Groups", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/groups", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Group A")) {
				t.Errorf("Expected group name in response, got %s", body)
			}
		})

		// Test GET /api/v1/groups/{group_id}
		t.Run("Get Group By ID", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", groupID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Group A")) {
				t.Errorf("Expected group name in response, got %s", body)
			}
		})

		// Test PUT /api/v1/groups/{group_id}
		t.Run("Update Group", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/groups/%d", groupID), adminAuthToken, map[string]string{
				"name":        "Group B",
				"description": "Afternoon class",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Group updated successfully")) {
				t.Errorf("Expected updated group name in response, got %s", body)
			}
		})

		// Test DELETE /api/v1/groups/{group_id}
		t.Run("Delete Group", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/groups/%d", groupID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Group deleted successfully")) {
				t.Errorf("Expected delete success message, got %s", body)
			}
		})
	})

	// Categories Management Endpoints
	t.Run("Categories Management Endpoints", func(t *testing.T) {
		var categoryID int
		// Test POST /api/v1/categories
		t.Run("Create Category", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/categories", adminAuthToken, map[string]string{
				"name":        "Art",
				"description": "Creative activities",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
			}
			var categoryResp struct {
				ID int `json:"id"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &categoryResp); err != nil {
				t.Fatalf("Failed to unmarshal category creation response: %v", err)
			}
			categoryID = categoryResp.ID
			if categoryID == 0 {
				t.Error("Expected category ID, got empty string")
			}
		})

		// Test GET /api/v1/categories
		t.Run("Get All Categories", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/categories", authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Art")) {
				t.Errorf("Expected category name in response, got %s", body)
			}
		})

		// Test PUT /api/v1/categories/{category_id}
		t.Run("Update Category", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/categories/%d", categoryID), adminAuthToken, map[string]string{
				"name":        "Music",
				"description": "Musical activities",
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Category updated successfully")) {
				t.Errorf("Expected updated category name in response, got %s", body)
			}
		})

		// Test DELETE /api/v1/categories/{category_id}
		t.Run("Delete Category", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", categoryID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Category deleted successfully")) {
				t.Errorf("Expected delete success message, got %s", body)
			}
		})
	})

	// Child-Teacher Assignments Endpoints
	t.Run("Child-Teacher Assignments Endpoints", func(t *testing.T) {
		// First, create a child and a teacher for assignment
		var childID, teacherID int
		t.Run("Setup Child and Teacher for Assignment", func(t *testing.T) {
			respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
				"first_name": "AssignChild",
				"last_name":  "Test",
				"birthdate":  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				"gender":     "other",
			}, "application/json")
			defer respChild.Body.Close() //nolint:errcheck
			var childResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respChild), &childResp) //nolint:errcheck
			childID = childResp.ID

			respTeacher := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/teachers", adminAuthToken, map[string]string{
				"first_name": "AssignTeacher",
				"last_name":  "Test",
			}, "application/json")
			defer respTeacher.Body.Close() //nolint:errcheck
			var teacherResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respTeacher), &teacherResp) //nolint:errcheck
			teacherID = teacherResp.ID
		})

		var assignmentID int
		// Test POST /api/v1/assignments
		t.Run("Create Assignment", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/assignments", authToken, map[string]interface{}{
				"child_id":   childID,
				"teacher_id": teacherID,
				"start_date": time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
			}
			var assignmentResp struct {
				ID int `json:"id"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &assignmentResp); err != nil {
				t.Fatalf("Failed to unmarshal assignment creation response: %v", err)
			}
			assignmentID = assignmentResp.ID
			if assignmentID == 0 {
				t.Error("Expected assignment ID, got empty string")
			}
		})

		// Test GET /api/v1/assignments/child/{child_id}
		t.Run("Get Assignments by Child ID", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/assignments/child/%d", childID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte(strconv.Itoa(teacherID))) {
				t.Errorf("Expected teacher ID in assignment response, got %s", body)
			}
		})

		// Test PUT /api/v1/assignments/{assignment_id}
		t.Run("Update Assignment", func(t *testing.T) {
			// Create another teacher to reassign
			var newTeacherID int
			respNewTeacher := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/teachers", adminAuthToken, map[string]string{
				"first_name": "NewAssignTeacher",
				"last_name":  "Test",
			}, "application/json")
			defer respNewTeacher.Body.Close() //nolint:errcheck
			var newTeacherResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respNewTeacher), &newTeacherResp) //nolint:errcheck
			newTeacherID = newTeacherResp.ID

			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/assignments/%d", assignmentID), authToken, map[string]interface{}{
				"child_id":   childID,
				"teacher_id": newTeacherID,
				"start_date": time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Assignment updated successfully")) {
				t.Errorf("Expected updated teacher ID in assignment response, got %s", body)
			}
		})

		// Test DELETE /api/v1/assignments/{assignment_id}
		t.Run("Delete Assignment", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/assignments/%d", assignmentID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Assignment deleted successfully")) {
				t.Errorf("Expected delete success message, got %s", body)
			}
		})
	})

	// Audio Recordings Endpoints
	t.Run("Audio Recordings Endpoints", func(t *testing.T) {
		// Test POST /api/v1/audio/upload
		t.Run("Upload Audio Recording", func(t *testing.T) {
			// Create a buffer to hold the multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add a test audio file to the form
			part, err := writer.CreateFormFile("audio", "test_audio.mp3")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			// Create a small dummy audio content
			dummyAudioContent := []byte("This is a test audio file content")
			if _, err := part.Write(dummyAudioContent); err != nil {
				t.Fatalf("Failed to write to form file: %v", err)
			}

			// Close the multipart writer
			if err := writer.Close(); err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

			// Create the request manually
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/audio/upload", body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+authToken)

			// Send the request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close() //nolint:errcheck

			// Verify response
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d or %d, got %d", http.StatusOK, http.StatusCreated, resp.StatusCode)
			}
			responseBody := readResponseBody(t, resp)
			if !bytes.Contains(responseBody, []byte("Audio uploaded successfully")) {
				t.Errorf("Expected upload success message, got %s", responseBody)
			}
		})
	})

	// Document Generation Endpoints
	t.Run("Document Generation Endpoints", func(t *testing.T) {
		// First, create a child for document generation
		var childID int
		t.Run("Setup Child for Document Generation", func(t *testing.T) {
			respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
				"first_name": "ReportChild",
				"last_name":  "Test",
				"birthdate":  time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				"gender":     "other",
			}, "application/json")
			defer respChild.Body.Close() //nolint:errcheck
			var childResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respChild), &childResp) //nolint:errcheck
			childID = childResp.ID
		})

		// Test GET /api/v1/documents/child-report/{child_id}
		t.Run("Generate Child Report", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/documents/child-report/%d", childID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			// For document generation, we might expect a specific content type (e.g., application/pdf)
			// and some non-empty body. A simple check for non-empty body is sufficient for happy path.
			body := readResponseBody(t, resp)
			if len(body) == 0 {
				t.Error("Expected non-empty report content, got empty")
			}
		})
	})

	// Documentation Entries Endpoints
	t.Run("Documentation Entries Endpoints", func(t *testing.T) {
		// First, create a child for documentation entry
		var childID int
		t.Run("Setup Child for Documentation", func(t *testing.T) {
			respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
				"first_name": "DocChild",
				"last_name":  "Test",
				"birthdate":  time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				"gender":     "male",
			}, "application/json")
			defer respChild.Body.Close() //nolint:errcheck
			var childResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respChild), &childResp) //nolint:errcheck
			childID = childResp.ID
		})

		// Create a teacher for documentation entry
		var teacherID int
		t.Run("Setup Teacher for Documentation", func(t *testing.T) {
			respTeacher := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/teachers", adminAuthToken, map[string]string{
				"first_name": "DocTeacher",
				"last_name":  "Test",
			}, "application/json")
			defer respTeacher.Body.Close() //nolint:errcheck
			var teacherResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respTeacher), &teacherResp) //nolint:errcheck
			teacherID = teacherResp.ID
		})

		// Create a category for documentation entry
		var categoryID int
		t.Run("Setup Category for Documentation", func(t *testing.T) {
			respCategory := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/categories", adminAuthToken, map[string]string{
				"name": "DocCategory",
			}, "application/json")
			defer respCategory.Body.Close() //nolint:errcheck
			var categoryResp struct {
				ID int `json:"id"`
			}
			json.Unmarshal(readResponseBody(t, respCategory), &categoryResp) //nolint:errcheck
			categoryID = categoryResp.ID
		})

		var entryID int
		// Test POST /api/v1/documentation
		t.Run("Create Documentation Entry", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/documentation", authToken, map[string]interface{}{
				"child_id":                childID,
				"teacher_id":              teacherID,
				"category_id":             categoryID,
				"observation_description": "Child showed great progress today.",
				"observation_date":        time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
			}
			var entryResp struct {
				ID int `json:"id"`
			}
			body := readResponseBody(t, resp)
			if err := json.Unmarshal(body, &entryResp); err != nil {
				t.Fatalf("Failed to unmarshal documentation entry creation response: %v", err)
			}
			entryID = entryResp.ID
			if entryID == 0 {
				t.Error("Expected entry ID, got 0")
			}
		})

		// Test GET /api/v1/documentation/child/{child_id}
		t.Run("Get Documentation Entries by Child ID", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/documentation/child/%d", childID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Child showed great progress today.")) {
				t.Errorf("Expected documentation title in response, got %s", body)
			}
		})

		// Test PUT /api/v1/documentation/{entry_id}
		t.Run("Update Documentation Entry", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/documentation/%d", entryID), authToken, map[string]interface{}{
				"child_id":                childID,
				"teacher_id":              teacherID,
				"category_id":             categoryID,
				"observation_description": "Child showed even greater progress today.",
				"observation_date":        time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			}, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Documentation entry updated successfully")) {
				t.Errorf("Expected updated documentation title in response, got %s", body)
			}
		})

		// Test PUT /api/v1/documentation/{entry_id}/approve
		t.Run("Approve Documentation Entry", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/documentation/%d/approve", entryID), authToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Documentation entry approved successfully")) {
				t.Errorf("Expected approval message, got %s", body)
			}
		})

		// Test DELETE /api/v1/documentation/{entry_id}
		t.Run("Delete Documentation Entry", func(t *testing.T) {
			resp := makeAuthenticatedRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/documentation/%d", entryID), adminAuthToken, nil, "application/json")
			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			body := readResponseBody(t, resp)
			if !bytes.Contains(body, []byte("Documentation entry deleted successfully")) {
				t.Errorf("Expected delete success message, got %s", body)
			}
		})
	})

	// Bulk Operations Endpoints
	t.Run("Bulk Operations Endpoints", func(t *testing.T) {
		// Test POST /api/v1/bulk/import-children
		t.Run("Import Children in Bulk", func(t *testing.T) {
			csvContent := `first_name,last_name,birth_date,gender
BulkChild1,Test,2023-01-01,male
BulkChild2,Test,2024-01-01,other`

			csvBuffer := bytes.NewBufferString(csvContent)
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Create form file field
			part, err := writer.CreateFormFile("children_csv", "test_data.csv")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			// Write CSV content to the form file
			if _, err := io.Copy(part, csvBuffer); err != nil {
				t.Fatalf("Failed to copy CSV to form file: %v", err)
			}

			// Close the multipart writer
			if err := writer.Close(); err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/bulk/import-children", body)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+adminAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}

			defer resp.Body.Close() //nolint:errcheck
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
			responseBody := readResponseBody(t, resp)
			if !bytes.Contains(responseBody, []byte("Bulk import completed successfully")) {
				t.Errorf("Expected bulk import success message, got %s", responseBody)
			}
		})
	})
}
