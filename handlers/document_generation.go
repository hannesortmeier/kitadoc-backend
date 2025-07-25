package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"kitadoc-backend/middleware"
	"kitadoc-backend/services"
)

// DocumentGenerationHandler handles document generation and download HTTP requests.
type DocumentGenerationHandler struct {
	DocumentationEntryService services.DocumentationEntryService
	AssignmentService		  services.AssignmentService
}

// NewDocumentGenerationHandler creates a new DocumentGenerationHandler.
func NewDocumentGenerationHandler(
	documentationEntryService services.DocumentationEntryService,
	assignmentService services.AssignmentService,
) *DocumentGenerationHandler {
	return &DocumentGenerationHandler{
		DocumentationEntryService: documentationEntryService,
		AssignmentService:         assignmentService,
	}
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

	assignments, err := handler.AssignmentService.GetAssignmentHistoryForChild(childID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			logger.WithField("child_id", childID).WithError(err).Warn("No assignments found for child")
		}
		logger.WithField("child_id", childID).WithError(err).Error("Internal server error during assignment retrieval")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	reportBytes, err := handler.DocumentationEntryService.GenerateChildReport(logger, ctx, childID, assignments)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			logger.WithField("child_id", childID).WithError(err).Warn("Child not found for report generation")
			http.Error(writer, "Child not found", http.StatusNotFound)
			return
		}
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
	documentName, err := handler.DocumentationEntryService.GetDocumentName(ctx, childID)
	if err != nil {
		logger.WithField("child_id", childID).WithError(err).Error("Failed to retrieve child details for report")
		http.Error(writer, "Failed to retrieve child details", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", documentName))
	if _, err := writer.Write(reportBytes); err != nil {
		logger.WithField("child_id", childID).WithError(err).Error("Failed to write report bytes to response")
		http.Error(writer, "Failed to write report", http.StatusInternalServerError)
		return
	}
}