package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/internal/testutils"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTeacherHandler(t *testing.T) {
	mockService := new(mocks.MockTeacherService)
	handler := NewTeacherHandler(mockService)
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.TeacherService)
}

func TestCreateTeacher(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Creation", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
		}

		createdTime := time.Now()
		mockService.On("CreateTeacher", mock.AnythingOfType("*models.Teacher")).Return(&models.Teacher{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
			CreatedAt: createdTime,
			UpdatedAt: createdTime,
		}, nil).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/teachers", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateTeacher(recorder, req)

		assert.Equal(t, http.StatusCreated, recorder.Code)

		var actualTeacher models.Teacher
		err := json.Unmarshal(recorder.Body.Bytes(), &actualTeacher)
		assert.NoError(t, err)
		assert.Equal(t, 1, actualTeacher.ID)
		assert.Equal(t, "John", actualTeacher.FirstName)
		assert.Equal(t, "Doe", actualTeacher.LastName)
		assert.Equal(t, "johndoe", actualTeacher.Username)
		assert.WithinDuration(t, createdTime, actualTeacher.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, createdTime, actualTeacher.UpdatedAt, 5*time.Second)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/teachers", bytes.NewBufferString(`{"first_name": 123}`))
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid request payload\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrInvalidInput", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			FirstName: "John",
			Username:  "johndoe",
		}

		mockService.On("CreateTeacher", mock.AnythingOfType("*models.Teacher")).Return(nil, services.ErrInvalidInput).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/teachers", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid teacher data provided\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
		}

		mockService.On("CreateTeacher", mock.AnythingOfType("*models.Teacher")).Return(nil, errors.New("database error")).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/teachers", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateTeacher(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestGetAllTeachers(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Fetch", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("GetAllTeachers").Return([]models.Teacher{
			{ID: 1, FirstName: "Jane", LastName: "Smith", Username: "janesmith"},
			{ID: 2, FirstName: "Peter", LastName: "Jones", Username: "peterjones"},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))

		recorder := httptest.NewRecorder()
		handler.GetAllTeachers(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `[{"id":1,"first_name":"Jane","last_name":"Smith","username":"janesmith","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":2,"first_name":"Peter","last_name":"Jones","username":"peterjones","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Error", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("GetAllTeachers").Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/teachers", nil)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))

		recorder := httptest.NewRecorder()
		handler.GetAllTeachers(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestGetTeacherByID(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Fetch", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("GetTeacherByID", 1).Return(&models.Teacher{ID: 1, FirstName: "John", Username: "johndoe"}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetTeacherByID(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"id":1,"first_name":"John","last_name":"","username":"johndoe","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Teacher ID", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/teachers/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "abc")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetTeacherByID(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid teacher ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("GetTeacherByID", 99).Return(nil, services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/teachers/99", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "99")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetTeacherByID(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Teacher not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("GetTeacherByID", 1).Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/teachers/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetTeacherByID(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestUpdateTeacher(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Update", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			ID:        1,
			FirstName: "Updated",
			LastName:  "Teacher",
		}

		mockService.On("UpdateTeacher", mock.AnythingOfType("*models.Teacher")).Return(nil).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/teachers/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"message":"Teacher updated successfully"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Teacher ID", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		req := httptest.NewRequest(http.MethodPut, "/teachers/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "abc")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid teacher ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		req := httptest.NewRequest(http.MethodPut, "/teachers/1", bytes.NewBufferString(`{"first_name": 123}`))
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid request payload\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			ID: 99, FirstName: "NonExistent", LastName: "Teacher",
		}

		mockService.On("UpdateTeacher", mock.AnythingOfType("*models.Teacher")).Return(services.ErrNotFound).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/teachers/99", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "99")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Teacher not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrInvalidInput", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			ID: 1, FirstName: "", LastName: "Teacher",
		}

		mockService.On("UpdateTeacher", mock.AnythingOfType("*models.Teacher")).Return(services.ErrInvalidInput).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/teachers/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid teacher data provided\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		inputTeacher := models.Teacher{
			ID: 1, FirstName: "Test", LastName: "Teacher",
		}

		mockService.On("UpdateTeacher", mock.AnythingOfType("*models.Teacher")).Return(errors.New("database error")).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputTeacher) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/teachers/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateTeacher(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestDeleteTeacher(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Deletion", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("DeleteTeacher", 1).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/teachers/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteTeacher(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"message":"Teacher deleted successfully"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Teacher ID", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		req := httptest.NewRequest(http.MethodDelete, "/teachers/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "abc")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteTeacher(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid teacher ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("DeleteTeacher", 99).Return(services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/teachers/99", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "99")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteTeacher(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Teacher not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("DeleteTeacher", 1).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/teachers/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteTeacher(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrForeignKeyConstraint", func(t *testing.T) {
		mockService := new(mocks.MockTeacherService)
		handler := NewTeacherHandler(mockService)

		mockService.On("DeleteTeacher", 2).Return(services.ErrForeignKeyConstraint).Once()

		req := httptest.NewRequest(http.MethodDelete, "/teachers/2", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("teacher_id", "2")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteTeacher(recorder, req)

		assert.Equal(t, http.StatusConflict, recorder.Code)
		assert.Equal(t, "Cannot delete teacher: foreign key constraint violation\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}
