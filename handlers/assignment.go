package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// AssignmentHandler handles assignment-related HTTP requests.
type AssignmentHandler struct {
	AssignmentService services.AssignmentService
}

// NewAssignmentHandler creates a new AssignmentHandler.
func NewAssignmentHandler(assignmentService services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{AssignmentService: assignmentService}
}

// CreateAssignment handles creating a new assignment.
func (assignmentHandler *AssignmentHandler) CreateAssignment(writer http.ResponseWriter, request *http.Request) {
	var assignment models.Assignment
	if err := json.NewDecoder(request.Body).Decode(&assignment); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	assignment.StartDate = time.Now()
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	createdAssignment, err := assignmentHandler.AssignmentService.CreateAssignment(&assignment)
	if err != nil {
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid assignment data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(createdAssignment)
}

// GetAssignmentsByChildID handles fetching assignments by child ID.
func (assignmentHandler *AssignmentHandler) GetAssignmentsByChildID(writer http.ResponseWriter, request *http.Request) {
	childIDStr := request.PathValue("child_id")
	childID, err := strconv.Atoi(childIDStr)
	if err != nil {
		http.Error(writer, "Invalid child ID", http.StatusBadRequest)
		return
	}

	assignments, err := assignmentHandler.AssignmentService.GetAssignmentHistoryForChild(childID)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(writer).Encode(assignments)
}

// UpdateAssignment handles updating an existing assignment.
func (assignmentHandler *AssignmentHandler) UpdateAssignment(writer http.ResponseWriter, request *http.Request) {
	assignmentIDStr := request.PathValue("assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		http.Error(writer, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	var assignment models.Assignment
	if err := json.NewDecoder(request.Body).Decode(&assignment); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	assignment.ID = assignmentID
	assignment.UpdatedAt = time.Now()

	err = assignmentHandler.AssignmentService.UpdateAssignment(&assignment)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Assignment not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid assignment data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Assignment updated successfully"})
}

// DeleteAssignment handles deleting an assignment.
func (assignmentHandler *AssignmentHandler) DeleteAssignment(writer http.ResponseWriter, request *http.Request) {
	assignmentIDStr := request.PathValue("assignment_id")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		http.Error(writer, "Invalid assignment ID", http.StatusBadRequest)
		return
	}

	err = assignmentHandler.AssignmentService.DeleteAssignment(assignmentID)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Assignment not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Assignment deleted successfully"})
}
