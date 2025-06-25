package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// BulkOperationsHandler handles bulk operations HTTP requests.
type BulkOperationsHandler struct {
	ChildService services.ChildService
}

// NewBulkOperationsHandler creates a new BulkOperationsHandler.
func NewBulkOperationsHandler(childService services.ChildService) *BulkOperationsHandler {
	return &BulkOperationsHandler{ChildService: childService}
}

// ImportChildren handles bulk import of children from a CSV file.
func (bulkOperationsHandler *BulkOperationsHandler) ImportChildren(writer http.ResponseWriter, request *http.Request) {
	// 1. Parse multipart form data
	err := request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(writer, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Get the CSV file from the form
	file, _, err := request.FormFile("children_csv")
	if err != nil {
		http.Error(writer, "Error retrieving CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close() //nolint:errcheck

	// 3. Read the CSV data
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(writer, "Failed to read CSV data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		http.Error(writer, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Assuming the first row is the header
	// Expected headers: FirstName,LastName,DateOfBirth,Gender
	// You might want more robust header validation.

	var importedChildren []*models.Child
	var importErrors []string

	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}

		if len(record) != 4 { // Expecting 4 columns
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid number of columns. Expected 4, got %d", i+1, len(record)))
			continue
		}

		dob, err := time.Parse("2006-01-02", record[2]) // Assuming YYYY-MM-DD format
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid DateOfBirth format: %v", i+1, err))
			continue
		}

		child := &models.Child{
			FirstName: record[0],
			LastName:  record[1],
			Birthdate: dob,
			Gender:    record[3],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		createdChild, err := bulkOperationsHandler.ChildService.CreateChild(child)
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Failed to create child: %v", i, err))
			continue
		}
		importedChildren = append(importedChildren, createdChild)
	}

	if len(importErrors) > 0 {
		writer.WriteHeader(http.StatusPartialContent) // Some items failed
		if err := json.NewEncoder(writer).Encode(map[string]interface{}{
			"message":        "Bulk import completed with errors",
			"imported_count": len(importedChildren),
			"errors":         importErrors,
		}); err != nil {
			http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			// Log the error but do not return here, we still want to return the successful imports
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"message":        "Bulk import completed successfully",
		"imported_count": len(importedChildren),
		"children":       importedChildren,
	}); err != nil {
		http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
