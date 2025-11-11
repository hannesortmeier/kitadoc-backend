package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// ChildHandler handles child-related HTTP requests.
type ChildHandler struct {
	ChildService services.ChildService
}

// NewChildHandler creates a new ChildHandler.
func NewChildHandler(childService services.ChildService) *ChildHandler {
	return &ChildHandler{ChildService: childService}
}

// CreateChild handles creating a new child.
func (childHandler *ChildHandler) CreateChild(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	var child models.Child
	if err := json.NewDecoder(request.Body).Decode(&child); err != nil {
		logger.Errorf("Failed to decode request body: %v", err)
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdChild, err := childHandler.ChildService.CreateChild(&child)
	if err != nil {
		if err == services.ErrInvalidInput {
			logger.Errorf("Invalid child data provided: %v", err)
			http.Error(writer, "Invalid child data provided", http.StatusBadRequest)
			return
		}
		logger.Errorf("Failed to create child: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(writer).Encode(createdChild); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllChildren handles fetching all children.
func (childHandler *ChildHandler) GetAllChildren(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	children, err := childHandler.ChildService.GetAllChildren()
	if err != nil {
		logger.Errorf("Failed to get all children: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(children); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetChildByID handles fetching a child by ID.
func (childHandler *ChildHandler) GetChildByID(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	idStr := request.PathValue("child_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Errorf("Invalid child ID: %v", err)
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	child, err := childHandler.ChildService.GetChildByID(id)
	if err != nil {
		if err == services.ErrNotFound {
			logger.Errorf("Child not found: %d", id)
			http.Error(writer, "Child not found", http.StatusNotFound)
			return
		}
		logger.Errorf("Failed to get child: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(child); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateChild handles updating an existing child.
func (childHandler *ChildHandler) UpdateChild(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	idStr := request.PathValue("child_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Errorf("Invalid child ID: %v", err)
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	var child models.Child
	if err := json.NewDecoder(request.Body).Decode(&child); err != nil {
		logger.Errorf("Failed to decode request body: %v", err)
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	child.ID = id

	err = childHandler.ChildService.UpdateChild(&child)
	if err != nil {
		if err == services.ErrNotFound {
			logger.Errorf("Child not found: %d", child.ID)
			http.Error(writer, "Child not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			logger.Errorf("Invalid child data provided: %v", err)
			http.Error(writer, "Invalid child data provided", http.StatusBadRequest)
			return
		}
		logger.Errorf("Failed to update child: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Child updated successfully"}); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteChild handles deleting a child.
func (childHandler *ChildHandler) DeleteChild(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	idStr := request.PathValue("child_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Errorf("Invalid child ID: %v", err)
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	err = childHandler.ChildService.DeleteChild(id)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			logger.Errorf("Child not found: %d", id)
			http.Error(writer, "Child not found", http.StatusNotFound)
			return
		case services.ErrForeignKeyConstraint:
			http.Error(writer, "Cannot delete child: foreign key constraint violation", http.StatusConflict)
			return
		}
		logger.Errorf("Failed to delete child: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Child deleted successfully"}); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
