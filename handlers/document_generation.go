package handlers

import (
	"context"
	"net/http"
	"strconv"

	"kitadoc-backend/middleware"
	"kitadoc-backend/services"
)

// DocumentGenerationHandler handles document generation and download HTTP requests.
type DocumentGenerationHandler struct {
	DocumentationEntryService services.DocumentationEntryService
}

// NewDocumentGenerationHandler creates a new DocumentGenerationHandler.
func NewDocumentGenerationHandler(documentationEntryService services.DocumentationEntryService) *DocumentGenerationHandler {
	return &DocumentGenerationHandler{DocumentationEntryService: documentationEntryService}
}

// GenerateChildReport handles generating a child report.
func (handler *DocumentGenerationHandler) GenerateChildReport(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())

	childIDStr := request.PathValue("child_id")
	childID, err := strconv.Atoi(childIDStr)
	if err != nil {
		logger.WithField("child_id_str", childIDStr).WithError(err).Warn("Invalid child ID format for report generation")
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	logger.WithField("child_id", childID).Info("Generating child report")

	// Use context for graceful shutdown and cancellation
	ctx, cancel := context.WithCancel(request.Context())
	defer cancel()

	reportBytes, err := handler.DocumentationEntryService.GenerateChildReport(logger, ctx, childID)
	if err != nil {
		if err == services.ErrChildReportGenerationFailed {
			logger.WithField("child_id", childID).WithError(err).Error("Failed to generate child report in service")
			http.Error(writer, "Failed to generate child report", http.StatusInternalServerError)
			return
		}
		logger.WithField("child_id", childID).WithError(err).Error("Internal server error during child report generation")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.WithField("child_id", childID).Info("Child report generated successfully, sending for download")

	writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	writer.Header().Set("Content-Disposition", "attachment; filename=\"child_report.docx\"")
	writer.Write(reportBytes)
}