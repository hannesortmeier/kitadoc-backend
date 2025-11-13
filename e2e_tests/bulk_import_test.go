package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestBulkImportChildrenFromXLSX(t *testing.T) {
	setupTest(t)

	// Path to the test file
	filePath := filepath.Join("test_uploads", "Kindliste.xlsx")

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close() // nolint:errcheck

	// Create a buffer to store our request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form-data header with the provided file name
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy the file content to the form part
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content to form part: %v", err)
	}

	// Close the multipart writer to set the terminating boundary
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Create a new request with the body and correct content type
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/bulk/import-children", body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+adminAuthToken)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		responseBody := readResponseBody(t, resp)
		t.Fatalf("Expected status %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(responseBody))
	}

	// Check the response body for success message
	responseBody := readResponseBody(t, resp)
	if !bytes.Contains(responseBody, []byte("Bulk import completed successfully")) {
		t.Errorf("Expected bulk import success message, got %s", responseBody)
	}

	// Verify that the children were actually created
	t.Run("Verify Children Creation", func(t *testing.T) {
		resp := makeAuthenticatedRequest(t, http.MethodGet, "/api/v1/children", authToken, nil, "application/json")
		defer resp.Body.Close() // nolint:errcheck
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to get children: %s", readResponseBody(t, resp))
		}

		body := readResponseBody(t, resp)
		// Parse the response to list of children
		// Response looks like this: "[{\"id\":1,\"first_name\":\"l\",\"last_name\":\"k\",\"birthdate\":\"2022-11-18T00:00:00Z\",\"gender\":\"DUMMY_GENDER\",\"family_language\":\"DUMMY_LANGUAGE\",\"migration_background\":false,\"admission_date\":\"1900-01-01T00:00:00Z\",\"expected_school_enrollment\":\"2029-07-31T00:00:00Z\",\"address\":\"DUMMY_ADDRESS\",\"parent1_name\":\"k\",\"parent2_name\":\"k\",\"created_at\":\"2025-11-13T20:06:20Z\",\"updated_at\":\"2025-11-13T20:06:20Z\"},{\"id\":2,\"first_name\":\"a\",\"last_name\":\"a\",\"birthdate\":\"2020-11-10T00:00:00Z\",\"gender\":\"DUMMY_GENDER\",\"family_language\":...+4997 more"
		var children []struct {
			FirstName                string `json:"first_name"`
			LastName                 string `json:"last_name"`
			Birthdate                string `json:"birthdate"`
			ExpectedSchoolEnrollment string `json:"expected_school_enrollment"`
			Parent1Name              string `json:"parent1_name"`
			Parent2Name              string `json:"parent2_name"`
		}

		err := json.Unmarshal(body, &children)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		// Check that the expected children are in the list
		expectedChildren := []struct {
			FirstName                string
			LastName                 string
			Birthdate                string
			ExpectedSchoolEnrollment string
			Parent1Name              string
			Parent2Name              string
		}{
			{"l", "k", "18.11.2022", "31.07.2029", "k", "k"},
			{"a", "a", "10.11.2020", "31.07.2027", "a", "a"},
		}

		for _, expected := range expectedChildren {
			found := false
			for _, child := range children {
				if child.FirstName == expected.FirstName && child.LastName == expected.LastName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected child %+v to be in the list of children", expected)
			}
		}
	})
}
