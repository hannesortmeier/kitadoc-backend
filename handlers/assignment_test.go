package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/handlers/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAssignment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignment := models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now(),
		}
		mockService.On("CreateAssignment", mock.AnythingOfType("*models.Assignment")).Return(&assignment, nil).Once()

		body, _ := json.Marshal(assignment)
		req := httptest.NewRequest(http.MethodPost, "/assignments", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateAssignment(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var createdAssignment models.Assignment
		json.NewDecoder(rr.Body).Decode(&createdAssignment)
		assert.Equal(t, assignment.ChildID, createdAssignment.ChildID)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/assignments", bytes.NewBuffer([]byte("invalid json")))
		rr := httptest.NewRecorder()

		handler.CreateAssignment(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockService.AssertNotCalled(t, "CreateAssignment", mock.Anything)
	})

	t.Run("service returns invalid input error", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignment := models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now(),
		}
		mockService.On("CreateAssignment", mock.AnythingOfType("*models.Assignment")).Return(nil, services.ErrInvalidInput).Once()

		body, _ := json.Marshal(assignment)
		req := httptest.NewRequest(http.MethodPost, "/assignments", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateAssignment(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid assignment data provided")
		mockService.AssertExpectations(t)
	})

	t.Run("service returns internal server error", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignment := models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now(),
		}
		mockService.On("CreateAssignment", mock.AnythingOfType("*models.Assignment")).Return(nil, errors.New("db error")).Once()

		body, _ := json.Marshal(assignment)
		req := httptest.NewRequest(http.MethodPost, "/assignments", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateAssignment(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestGetAssignmentsByChildID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		childID := 1
		assignments := []models.Assignment{
			{ID: 1, ChildID: childID, StartDate: time.Now()},
			{ID: 2, ChildID: childID, StartDate: time.Now()},
		}
		mockService.On("GetAssignmentHistoryForChild", childID).Return(assignments, nil).Once()

		router := http.NewServeMux()
		router.HandleFunc("GET /assignments/child/{child_id}", handler.GetAssignmentsByChildID)

		req := httptest.NewRequest(http.MethodGet, "/assignments/child/"+strconv.Itoa(childID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var fetchedAssignments []models.Assignment
		json.NewDecoder(rr.Body).Decode(&fetchedAssignments)
		assert.Len(t, fetchedAssignments, 2)
		assert.Equal(t, assignments[0].ID, fetchedAssignments[0].ID)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid child ID", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		router := http.NewServeMux()
		router.HandleFunc("GET /assignments/child/{child_id}", handler.GetAssignmentsByChildID)

		req := httptest.NewRequest(http.MethodGet, "/assignments/child/abc", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid child ID")
		mockService.AssertNotCalled(t, "GetAssignmentHistoryForChild", mock.Anything)
	})

	t.Run("service returns error", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		childID := 1
		mockService.On("GetAssignmentHistoryForChild", childID).Return(nil, errors.New("db error")).Once()

		router := http.NewServeMux()
		router.HandleFunc("GET /assignments/child/{child_id}", handler.GetAssignmentsByChildID)

		req := httptest.NewRequest(http.MethodGet, "/assignments/child/"+strconv.Itoa(childID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestUpdateAssignment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		assignment := models.Assignment{
			ID:          assignmentID,
			ChildID:     1,
			TeacherID:   1,
			StartDate:   time.Now(),
		}
		mockService.On("UpdateAssignment", mock.AnythingOfType("*models.Assignment")).Return(nil).Once()

		body, _ := json.Marshal(assignment)
		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/"+strconv.Itoa(assignmentID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Assignment updated successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid assignment ID", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/abc", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid assignment ID")
		mockService.AssertNotCalled(t, "UpdateAssignment", mock.Anything)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/"+strconv.Itoa(assignmentID), bytes.NewBuffer([]byte("invalid json")))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockService.AssertNotCalled(t, "UpdateAssignment", mock.Anything)
	})

	t.Run("assignment not found", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		assignment := models.Assignment{
			ID:          assignmentID,
			ChildID:     1,
			TeacherID:   1,
			StartDate:   time.Now(),
		}
		mockService.On("UpdateAssignment", mock.AnythingOfType("*models.Assignment")).Return(services.ErrNotFound).Once()

		body, _ := json.Marshal(assignment)
		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/"+strconv.Itoa(assignmentID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Assignment not found")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid assignment data provided", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		assignment := models.Assignment{
			ID:          assignmentID,
			ChildID:     1,
			TeacherID:   1,
			StartDate:   time.Now(),
		}
		mockService.On("UpdateAssignment", mock.AnythingOfType("*models.Assignment")).Return(services.ErrInvalidInput).Once()

		body, _ := json.Marshal(assignment)
		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/"+strconv.Itoa(assignmentID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid assignment data provided")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		assignment := models.Assignment{
			ID:          assignmentID,
			ChildID:     1,
			TeacherID:   1,
			StartDate:   time.Now(),
		}
		mockService.On("UpdateAssignment", mock.AnythingOfType("*models.Assignment")).Return(errors.New("db error")).Once()

		body, _ := json.Marshal(assignment)
		router := http.NewServeMux()
		router.HandleFunc("PUT /assignments/{assignment_id}", handler.UpdateAssignment)

		req := httptest.NewRequest(http.MethodPut, "/assignments/"+strconv.Itoa(assignmentID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestDeleteAssignment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		mockService.On("DeleteAssignment", assignmentID).Return(nil).Once()

		router := http.NewServeMux()
		router.HandleFunc("DELETE /assignments/{assignment_id}", handler.DeleteAssignment)

		req := httptest.NewRequest(http.MethodDelete, "/assignments/"+strconv.Itoa(assignmentID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Assignment deleted successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid assignment ID", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		router := http.NewServeMux()
		router.HandleFunc("DELETE /assignments/{assignment_id}", handler.DeleteAssignment)

		req := httptest.NewRequest(http.MethodDelete, "/assignments/abc", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid assignment ID")
		mockService.AssertNotCalled(t, "DeleteAssignment", mock.Anything)
	})

	t.Run("assignment not found", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)

		assignmentID := 1
		mockService.On("DeleteAssignment", assignmentID).Return(services.ErrNotFound).Once()

		router := http.NewServeMux()
		router.HandleFunc("DELETE /assignments/{assignment_id}", handler.DeleteAssignment)

		req := httptest.NewRequest(http.MethodDelete, "/assignments/"+strconv.Itoa(assignmentID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Assignment not found")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.AssignmentService)
		handler := NewAssignmentHandler(mockService)
		
		assignmentID := 1
		mockService.On("DeleteAssignment", assignmentID).Return(errors.New("db error")).Once()

		router := http.NewServeMux()
		router.HandleFunc("DELETE /assignments/{assignment_id}", handler.DeleteAssignment)

		req := httptest.NewRequest(http.MethodDelete, "/assignments/"+strconv.Itoa(assignmentID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}