package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"kitadoc-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestGetAllAssignments(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/assignments", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add authentication token to the request
	req.Header.Set("Authorization", "Bearer "+authToken)

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
