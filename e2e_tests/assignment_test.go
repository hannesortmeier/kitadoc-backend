package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"kitadoc-backend/models"

	"github.com/stretchr/testify/assert"
)

func getAuthToken(t *testing.T) (string, error) {
	// This is a simplified token retrieval. In a real scenario, you would
	// likely have a more robust way to log in and get a token.
	resp := makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "testuser",
		"password": "password123",
	}, "application/json")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", assert.AnError
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

func TestGetAllAssignments(t *testing.T) {
	// Register a user
	makeUnauthenticatedRequest(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "testuser",
		"password": "password123",
		"role":     "teacher",
	}, "application/json")

	// Log in to get a token
	token, err := getAuthToken(t)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/assignments", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add authentication token to the request
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check the response body
	var assignments []models.Assignment
	err = json.NewDecoder(resp.Body).Decode(&assignments)
	assert.NoError(t, err)
	assert.NotNil(t, assignments)
}
