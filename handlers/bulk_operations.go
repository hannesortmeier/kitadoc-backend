package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(writer, "Failed to read CSV data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		http.Error(writer, "CSV data is empty", http.StatusBadRequest)
		return
	}

	var importedChildren []*models.Child
	var importErrors []string

	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}

		if len(record) != 11 {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid number of columns. Expected 11, got %d", i+1, len(record)))
			continue
		}

		birthdate, err := time.Parse("2006-01-02", record[2])
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid Birthdate format: %v", i+1, err))
			continue
		}

		migrationBackground, err := strconv.ParseBool(record[5])
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid MigrationBackground format: %v", i+1, err))
			continue
		}

		admissionDate, err := time.Parse("2006-01-02", record[6])
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid AdmissionDate format: %v", i+1, err))
			continue
		}

		expectedSchoolEnrollment, err := time.Parse("2006-01-02", record[7])
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Invalid ExpectedSchoolEnrollment format: %v", i+1, err))
			continue
		}

		child := &models.Child{
			FirstName:                record[0],
			LastName:                 record[1],
			Birthdate:                birthdate,
			Gender:                   record[3],
			FamilyLanguage:           record[4],
			MigrationBackground:      migrationBackground,
			AdmissionDate:            admissionDate,
			ExpectedSchoolEnrollment: expectedSchoolEnrollment,
			Address:                  record[8],
			Parent1Name:              record[9],
			Parent2Name:              record[10],
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		createdChild, err := bulkOperationsHandler.ChildService.CreateChild(child)
		if err != nil {
			importErrors = append(importErrors, fmt.Sprintf("Row %d: Failed to create child: %v", i, err))
			continue
		}
		importedChildren = append(importedChildren, createdChild)
	}

	if len(importErrors) > 0 {
		writer.WriteHeader(http.StatusPartialContent)
		if err := json.NewEncoder(writer).Encode(map[string]interface{}{
			"message":        "Bulk import completed with errors",
			"imported_count": len(importedChildren),
			"errors":         importErrors,
		}); err != nil {
			http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
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
