package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	authToken      string
	adminAuthToken string
)

func setupTest(t *testing.T) {
	// Disable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("failed to disable foreign keys: %v", err)
	}

	// Drop all tables to clean up the database
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name != 'sqlite_sequence';")
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	for _, table := range tables {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Fatalf("failed to drop table %s: %v", table, err)
		}
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Read and execute data_model.sql
	sqlContent, err := os.ReadFile("../database/data_model.sql")
	if err != nil {
		t.Fatalf("failed to read data_model.sql: %v", err)
	}

	_, err = db.Exec(string(sqlContent))
	if err != nil {
		t.Fatalf("failed to execute data_model.sql: %v", err)
	}

	// Read and execute sample_data.sql
	sqlContent, err = os.ReadFile("../database/sample_data.sql")
	if err != nil {
		t.Fatalf("failed to read sample_data.sql: %v", err)
	}

	_, err = db.Exec(string(sqlContent))
	if err != nil {
		t.Fatalf("failed to execute sample_data.sql: %v", err)
	}

	// Create a regular user and get token
	resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "testuser",
		"password": "password123",
		"role":     "teacher",
	}, "application/json")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to register user: %s", readResponseBody(t, resp))
	}

	resp = makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "testuser",
		"password": "password123",
	}, "application/json")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to login user: %s", readResponseBody(t, resp))
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(readResponseBody(t, resp), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	authToken = loginResp.Token

	// Create an admin user and get token
	resp = makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "adminuser",
		"password": "password123",
		"role":     "admin",
	}, "application/json")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to register admin user: %s", readResponseBody(t, resp))
	}

	resp = makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "adminuser",
		"password": "password123",
	}, "application/json")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to login admin user: %s", readResponseBody(t, resp))
	}
	if err := json.Unmarshal(readResponseBody(t, resp), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal admin login response: %v", err)
	}
	adminAuthToken = loginResp.Token
}

func teardownTest(t *testing.T) {
	// Disable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("failed to disable foreign keys: %v", err)
	}

	// Drop all tables to clean up the database
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name != 'sqlite_sequence';")
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	for _, table := range tables {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Fatalf("failed to drop table %s: %v", table, err)
		}
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
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
	setupTest(t)
	defer teardownTest(t)

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
}

	func TestAuthEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

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
}

	func TestChildrenManagementEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	var childID int
	// Test POST /api/v1/children
	t.Run("Create Child", func(t *testing.T) {
		resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
			"first_name":                 "John",
			"last_name":                  "Test",
			"birthdate":                  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			"gender":                     "female",
			"migration_background":       true,
			"family_language":            "Niederländisch",
			"admission_date":             time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			"expected_school_enrollment": time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			"address":                    "123 Test St, Test City, TC 12345",
			"parent1_name":               "Parent One",
			"parent2_name":               "Parent Two",
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
			"first_name":                 "Jane",
			"last_name":                  "Doe",
			"birthdate":                  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			"gender":                     "female",
			"migration_background":       true,
			"family_language":            "Niederländisch",
			"admission_date":             time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			"expected_school_enrollment": time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			"address":                    "123 Test St, Test City, TC 12345",
			"parent1_name":               "Parent One",
			"parent2_name":               "Parent Two",
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
}

	func TestTeachersManagementEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	var teacherID int
	// Test POST /api/v1/teachers
	t.Run("Create Teacher", func(t *testing.T) {
		resp := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/teachers", authToken, map[string]string{
			"first_name": "Alice",
			"last_name":  "Smith",
			"username":   "alicesmith",
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
			"username":   "aliciasmith",
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
}

	func TestCategoriesManagementEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

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
}

	func TestChildTeacherAssignmentsEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// First, create a child and a teacher for assignment
	var childID, teacherID int
	t.Run("Setup Child and Teacher for Assignment", func(t *testing.T) {
		respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
			"first_name":                 "AssignChild",
			"last_name":                  "Test",
			"birthdate":                  time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			"gender":                     "other",
			"migration_background":       true,
			"family_language":            "English",
			"admission_date":             time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			"expected_school_enrollment": time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			"address":                    "123 Test St, Test City, TC 12345",
			"parent1_name":               "Parent One",
			"parent2_name":               "Parent Two",
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
			"username":   "assignteacher",
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
			"username":   "newassignteacher",
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
}

	func TestAudioRecordingsEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// Test POST /api/v1/audio/upload
	// Create a category for documentation entry
	t.Run("Setup Category for Documentation", func(t *testing.T) {
		respCategory := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/categories", adminAuthToken, map[string]string{
			"name": "DocCategory",
		}, "application/json")
		defer respCategory.Body.Close() //nolint:errcheck
	})

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

		// Add teacher_id and timestamp fields
		if err := writer.WriteField("teacher_id", "2"); err != nil {
			t.Fatalf("Failed to write field: %v", err)
		}
		if err := writer.WriteField("timestamp", time.Now().Format(time.RFC3339)); err != nil {
			t.Fatalf("Failed to write field: %v", err)
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
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

	func TestDocumentGenerationEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// First, create a child for document generation
	var childID int
	t.Run("Setup Child for Document Generation", func(t *testing.T) {
		respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
			"first_name":                 "ReportChild",
			"last_name":                  "Test",
			"birthdate":                  time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			"gender":                     "other",
			"family_language":            "Deutsch",
			"migration_background":       false,
			"parent1_name":               "Parent One",
			"parent2_name":               "Parent Two",
			"address":                    "123 Main St, City, Country",
			"expected_school_enrollment": time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
			"admission_date":             time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
		// save the report to a file for manual inspection if needed
		if err := os.WriteFile("child_report.docx", body, 0644); err != nil {
			t.Fatalf("Failed to write report to file: %v", err)
		}
	})
}

	func TestDocumentationEntriesEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// First, create a child for documentation entry
	var childID int
	t.Run("Setup Child for Documentation", func(t *testing.T) {
		respChild := makeAuthenticatedRequest(t, http.MethodPost, "/api/v1/children", authToken, map[string]interface{}{
			"first_name":                 "DocChild",
			"last_name":                  "Test",
			"birthdate":                  time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			"gender":                     "other",
			"family_language":            "Deutsch",
			"migration_background":       false,
			"parent1_name":               "Parent One",
			"parent2_name":               "Parent Two",
			"address":                    "123 Main St, City, Country",
			"expected_school_enrollment": time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			"admission_date":             time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
			"username":   "doctestteacher",
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
			"name": "DocCategory2",
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
			"child_id":    childID,
			"teacher_id":  teacherID,
			"category_id": categoryID,

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
		reqBody := map[string]interface{}{
			"approvedByTeacherId": teacherID,
		}
		resp := makeAuthenticatedRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/documentation/%d/approve", entryID), authToken, reqBody, "application/json")
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
}

	func TestBulkOperationsEndpoints(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

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
		//if resp.StatusCode != http.StatusOK {
		//	t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		//}
		//responseBody := readResponseBody(t, resp)
		//if !bytes.Contains(responseBody, []byte("Bulk import completed successfully")) {
		//	t.Errorf("Expected bulk import success message, got %s", responseBody)
		//}
	})
}
