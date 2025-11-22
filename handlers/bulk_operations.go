package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/xuri/excelize/v2"
)

// BulkOperationsHandler handles bulk operations HTTP requests.
type BulkOperationsHandler struct {
	ChildService services.ChildService
}

// NewBulkOperationsHandler creates a new BulkOperationsHandler.
func NewBulkOperationsHandler(childService services.ChildService) *BulkOperationsHandler {
	return &BulkOperationsHandler{ChildService: childService}
}

// ImportChildren handles bulk import of children from an XLSX file.
func (bulkOperationsHandler *BulkOperationsHandler) ImportChildren(writer http.ResponseWriter, request *http.Request) {
	log := logger.GetLoggerFromContext(request.Context())

	// Parse the multipart form data
	err := request.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		log.Errorf("Failed to parse multipart form: %v", err)
		http.Error(writer, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, _, err := request.FormFile("file")
	if err != nil {
		log.Errorf("Failed to get file from form: %v", err)
		http.Error(writer, "Failed to get file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Errorf("Failed to close file: %v", closeErr)
		}
	}()

	// Open the XLSX file
	f, err := excelize.OpenReader(file)
	if err != nil {
		log.Errorf("Failed to open XLSX file: %v", err)
		http.Error(writer, "Failed to open XLSX file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all the rows from the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		log.Error("No sheet found in the XLSX file")
		http.Error(writer, "No sheet found in the XLSX file", http.StatusBadRequest)
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Errorf("Failed to get rows from sheet %s: %v", sheetName, err)
		http.Error(writer, "Failed to get rows from sheet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	headerRow := rows[0]
	dataRows := rows[1:]

	// Define the mapping from German headers to Child struct fields
	headerMapping := map[string]string{
		"Vorname":          "FirstName",
		"Nachname":         "LastName",
		"Geburtsdatum":     "Birthdate",
		"Aufnahmedatum":    "AdmissionDate",
		"Entlassungsdatum": "ExpectedSchoolEnrollment",
	}

	// Build a map from column index to Child struct field name
	colIndexToField := make(map[int]string)
	for i, header := range headerRow {
		trimmedHeader := strings.TrimSpace(header)
		if fieldName, ok := headerMapping[trimmedHeader]; ok {
			colIndexToField[i] = fieldName
		}
	}

	var importedChildren []*models.Child
	var importErrors []map[string]string

	for i, row := range dataRows {
		var createdChild *models.Child
		var err error
		child := &models.Child{}
		childName := "" // To store child's name for error reporting

		// Populate child struct from row data
		for colIndex, cellValue := range row {
			fieldName, ok := colIndexToField[colIndex]
			if !ok {
				continue // Skip columns not in our mapping
			}

			trimmedCellValue := strings.TrimSpace(cellValue)

			switch fieldName {
			case "FirstName":
				child.FirstName = trimmedCellValue
				childName = trimmedCellValue // Use first name as part of childName
			case "LastName":
				child.LastName = trimmedCellValue
				if childName != "" {
					childName = fmt.Sprintf("%s %s", childName, trimmedCellValue)
				} else {
					childName = trimmedCellValue
				}
			case "Birthdate":
				// Assuming date format DD.MM.YYYY
				birthdate, err := time.Parse("02.01.2006", trimmedCellValue)
				if err != nil {
					importErrors = append(importErrors, map[string]string{
						"child_name": childName,
						"error":      fmt.Sprintf("Reihe %d: Ungültiges Format für Geburtsdatum '%s'. Ein Datum im Format 02.01.2006 wird erwartet.", i+1, trimmedCellValue),
					})
					log.Warnf("Row %d: Invalid Birthdate format for child %s: %v", i+1, childName, err)
					goto nextRow // Skip to the next row
				}
				child.Birthdate = birthdate
			case "AdmissionDate":
				// Assuming date format DD.MM.YYYY
				admissionDate, err := time.Parse("02.01.2006", trimmedCellValue)
				if err != nil {
					importErrors = append(importErrors, map[string]string{
						"child_name": childName,
						"error":      fmt.Sprintf("Reihe %d: Ungültiges Format für Aufnahmedatum '%s'. Ein Datum im Format 02.01.2006 wird erwartet.", i+1, trimmedCellValue),
					})
					log.Warnf("Row %d: Invalid AdmissionDate format for child %s: %v", i+1, childName, err)
					goto nextRow // Skip to the next row
				}
				child.AdmissionDate = &admissionDate
			case "ExpectedSchoolEnrollment":
				// Assuming date format DD.MM.YYYY
				enrollmentDate, err := time.Parse("02.01.2006", trimmedCellValue)
				if err != nil {
					importErrors = append(importErrors, map[string]string{
						"child_name": childName,
						"error":      fmt.Sprintf("Reihe %d: Ungültiges Format für Entlassungsdatum '%s'. Ein Datum im Format 02.01.2006 wird erwartet.", i+1, trimmedCellValue),
					})
					log.Warnf("Row %d: Invalid ExpectedSchoolEnrollment format for child %s: %v", i+1, childName, err)
					goto nextRow // Skip to the next row
				}
				child.ExpectedSchoolEnrollment = &enrollmentDate
			}
		}

		// Validate the child struct before creation
		if err = models.ValidateChild(*child); err != nil {
			importErrors = append(importErrors, map[string]string{
				"child_name": childName,
				"error":      fmt.Sprintf("Reihe %d: Kind %s konnte nicht erfolgreich importiert werden: %v", i+1, childName, err),
			})
			log.Warnf("Row %d: Child validation failed for child %s: %v", i+1, childName, err)
			goto nextRow // Skip to the next row
		}

		// Set CreatedAt and UpdatedAt
		child.CreatedAt = time.Now()
		child.UpdatedAt = time.Now()

		createdChild, err = bulkOperationsHandler.ChildService.CreateChild(child)
		if err != nil {
			importErrors = append(importErrors, map[string]string{
				"child_name": childName,
				"error":      fmt.Sprintf("Reihe %d: Kind %s konnte nicht erstellt werden: %v", i+1, childName, err),
			})
			log.Errorf("Row %d: Failed to create child %s: %v", i+1, childName, err)
			goto nextRow // Skip to the next row
		}
		importedChildren = append(importedChildren, createdChild)

	nextRow:
		continue
	}

	if len(importErrors) > 0 {
		writer.WriteHeader(http.StatusPartialContent)
		if err := json.NewEncoder(writer).Encode(map[string]interface{}{
			"message":        "Massenimport mit Fehlern abgeschlossen.",
			"imported_count": len(importedChildren),
			"errors":         importErrors,
		}); err != nil {
			log.Errorf("Failed to encode response with partial content: %v", err)
			http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{
		"message":        "Massenimport erfolgreich abgeschlossen",
		"imported_count": len(importedChildren),
		"children":       importedChildren,
	}); err != nil {
		log.Errorf("Failed to encode success response: %v", err)
		http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
