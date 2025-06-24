package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// DocumentationEntryHandler handles documentation entry-related HTTP requests.
type DocumentationEntryHandler struct {
	DocumentationEntryService services.DocumentationEntryService
}

// NewDocumentationEntryHandler creates a new DocumentationEntryHandler.
func NewDocumentationEntryHandler(documentationEntryService services.DocumentationEntryService) *DocumentationEntryHandler {
	return &DocumentationEntryHandler{DocumentationEntryService: documentationEntryService}
}

// CreateDocumentationEntry handles creating a new documentation entry.
func (handler *DocumentationEntryHandler) CreateDocumentationEntry(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	var entry models.DocumentationEntry
	if err := json.NewDecoder(request.Body).Decode(&entry); err != nil {
		logger.WithError(err).Warn("Invalid request payload for CreateDocumentationEntry")
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	createdEntry, err := handler.DocumentationEntryService.CreateDocumentationEntry(logger, request.Context(), &entry)
	if err != nil {
		if err == services.ErrInvalidInput {
			logger.WithError(err).Warn("Invalid documentation entry data provided for creation")
			http.Error(writer, "Invalid documentation entry data provided", http.StatusBadRequest)
			return
		}
		logger.WithError(err).Error("Internal server error during documentation entry creation")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(createdEntry)
}

// GetDocumentationEntriesByChildID handles fetching documentation entries by child ID.
func (handler *DocumentationEntryHandler) GetDocumentationEntriesByChildID(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	childIDStr := request.PathValue("child_id")
	childID, err := strconv.Atoi(childIDStr)
	if err != nil {
		logger.WithField("child_id_str", childIDStr).WithError(err).Warn("Invalid child ID format for GetDocumentationEntriesByChildID")
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	entries, err := handler.DocumentationEntryService.GetAllDocumentationForChild(logger, request.Context(), childID)
	if err != nil {
		logger.WithError(err).WithField("child_id", childID).Error("Internal server error fetching documentation entries for child")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(writer).Encode(entries)
}

// UpdateDocumentationEntry handles updating an existing documentation entry.
func (handler *DocumentationEntryHandler) UpdateDocumentationEntry(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	entryIDStr := request.PathValue("entry_id")
	entryID, err := strconv.Atoi(entryIDStr)
	if err != nil {
		logger.WithField("entry_id_str", entryIDStr).WithError(err).Warn("Invalid entry ID format for UpdateDocumentationEntry")
		http.Error(writer, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	var entry models.DocumentationEntry
	if err := json.NewDecoder(request.Body).Decode(&entry); err != nil {
		logger.WithError(err).Warn("Invalid request payload for UpdateDocumentationEntry")
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	entry.ID = entryID
	entry.UpdatedAt = time.Now()

	err = handler.DocumentationEntryService.UpdateDocumentationEntry(logger, request.Context(), &entry)
	if err != nil {
		if err == services.ErrNotFound {
			logger.WithField("entry_id", entryID).Warn("Documentation entry not found for update")
			http.Error(writer, "Documentation entry not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			logger.WithError(err).Warn("Invalid documentation entry data provided for update")
			http.Error(writer, "Invalid documentation entry data provided", http.StatusBadRequest)
			return
		}
		logger.WithError(err).WithField("entry_id", entryID).Error("Internal server error during documentation entry update")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Documentation entry updated successfully"})
}

// DeleteDocumentationEntry handles deleting a documentation entry.
func (handler *DocumentationEntryHandler) DeleteDocumentationEntry(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	entryIDStr := request.PathValue("entry_id")
	entryID, err := strconv.Atoi(entryIDStr)
	if err != nil {
		logger.WithField("entry_id_str", entryIDStr).WithError(err).Warn("Invalid entry ID format for DeleteDocumentationEntry")
		http.Error(writer, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	err = handler.DocumentationEntryService.DeleteDocumentationEntry(logger, request.Context(), entryID)
	if err != nil {
		if err == services.ErrNotFound {
			logger.WithField("entry_id", entryID).Warn("Documentation entry not found for deletion")
			http.Error(writer, "Documentation entry not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithField("entry_id", entryID).Error("Internal server error during documentation entry deletion")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Documentation entry deleted successfully"})
}

// ApproveDocumentationEntry handles approving a documentation entry.
func (handler *DocumentationEntryHandler) ApproveDocumentationEntry(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	entryIDStr := request.PathValue("entry_id")
	entryID, err := strconv.Atoi(entryIDStr)
	if err != nil {
		logger.WithField("entry_id_str", entryIDStr).WithError(err).Warn("Invalid entry ID format for ApproveDocumentationEntry")
		http.Error(writer, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	// TODO(hannes): In a real application, the approvedByUserID would come from the authenticated user's context.
	// For now, we'll use a placeholder.
	approvedByUserID := 1 // Placeholder for the approving user's ID
	err = handler.DocumentationEntryService.ApproveDocumentationEntry(logger, request.Context(), entryID, approvedByUserID)
	if err != nil {
		if err == services.ErrNotFound {
			logger.WithField("entry_id", entryID).Warn("Documentation entry not found for approval")
			http.Error(writer, "Documentation entry not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithField("entry_id", entryID).Error("Internal server error during documentation entry approval")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Documentation entry approved successfully"})
}