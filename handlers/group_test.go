package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewGroupHandler(t *testing.T) {
	mockService := new(mocks.MockGroupService)
	handler := NewGroupHandler(mockService)
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.GroupService)
}

func TestCreateGroup(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Creation", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			Name: "Test Group",
		}

		createdTime := time.Now()
		mockService.On("CreateGroup", mock.AnythingOfType("*models.Group")).Return(&models.Group{
			ID:        1,
			Name:      "Test Group",
			CreatedAt: createdTime,
			UpdatedAt: createdTime,
		}, nil).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/groups", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateGroup(recorder, req)

		assert.Equal(t, http.StatusCreated, recorder.Code)

		var actualGroup models.Group
		err := json.Unmarshal(recorder.Body.Bytes(), &actualGroup)
		assert.NoError(t, err)
		assert.Equal(t, 1, actualGroup.ID)
		assert.Equal(t, "Test Group", actualGroup.Name)
		assert.WithinDuration(t, createdTime, actualGroup.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, createdTime, actualGroup.UpdatedAt, 5*time.Second)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(`{"name": 123}`))
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid request payload\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrInvalidInput", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			Name: "",
		}

		mockService.On("CreateGroup", mock.AnythingOfType("*models.Group")).Return(nil, services.ErrInvalidInput).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/groups", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid group data provided\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrAlreadyExists", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			Name: "Existing Group",
		}

		mockService.On("CreateGroup", mock.AnythingOfType("*models.Group")).Return(nil, services.ErrAlreadyExists).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/groups", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateGroup(recorder, req)

		assert.Equal(t, http.StatusConflict, recorder.Code)
		assert.Equal(t, "Group already exists\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			Name: "Test Group",
		}

		mockService.On("CreateGroup", mock.AnythingOfType("*models.Group")).Return(nil, errors.New("database error")).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPost, "/groups", &reqBody)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.CreateGroup(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestGetAllGroups(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Fetch", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("GetAllGroups").Return([]models.Group{
			{ID: 1, Name: "Group A"},
			{ID: 2, Name: "Group B"},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/groups", nil)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))

		recorder := httptest.NewRecorder()
		handler.GetAllGroups(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `[{"id":1,"name":"Group A","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":2,"name":"Group B","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Error", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("GetAllGroups").Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/groups", nil)
		req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))

		recorder := httptest.NewRecorder()
		handler.GetAllGroups(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestGetGroupByID(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Fetch", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("GetGroupByID", 1).Return(&models.Group{ID: 1, Name: "Test Group"}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/groups/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetGroupByID(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"id":1,"name":"Test Group","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Group ID", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/groups/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "abc")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetGroupByID(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid group ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("GetGroupByID", 99).Return(nil, services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/groups/99", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "99")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetGroupByID(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Group not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("GetGroupByID", 1).Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/groups/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GetGroupByID(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestUpdateGroup(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Update", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			ID:   1,
			Name: "Updated Group",
		}

		mockService.On("UpdateGroup", mock.AnythingOfType("*models.Group")).Return(nil).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/groups/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"message":"Group updated successfully"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Group ID", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		req := httptest.NewRequest(http.MethodPut, "/groups/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "abc")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid group ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		req := httptest.NewRequest(http.MethodPut, "/groups/1", strings.NewReader(`{"name": 123}`))
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid request payload\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			ID: 99, Name: "NonExistent",
		}

		mockService.On("UpdateGroup", mock.AnythingOfType("*models.Group")).Return(services.ErrNotFound).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/groups/99", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "99")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Group not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrInvalidInput", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			ID: 1, Name: "",
		}

		mockService.On("UpdateGroup", mock.AnythingOfType("*models.Group")).Return(services.ErrInvalidInput).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/groups/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid group data provided\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrAlreadyExists", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			ID: 1, Name: "Existing Group",
		}

		mockService.On("UpdateGroup", mock.AnythingOfType("*models.Group")).Return(services.ErrAlreadyExists).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/groups/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusConflict, recorder.Code)
		assert.Equal(t, "Group with the same name already exists\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		inputGroup := models.Group{
			ID: 1, Name: "Test Group",
		}

		mockService.On("UpdateGroup", mock.AnythingOfType("*models.Group")).Return(errors.New("database error")).Once()

		var reqBody bytes.Buffer
		json.NewEncoder(&reqBody).Encode(inputGroup) //nolint:errcheck

		req := httptest.NewRequest(http.MethodPut, "/groups/1", &reqBody)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		handler.UpdateGroup(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}

func TestDeleteGroup(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Deletion", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("DeleteGroup", 1).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/groups/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteGroup(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, `{"message":"Group deleted successfully"}`+"\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Group ID", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		req := httptest.NewRequest(http.MethodDelete, "/groups/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "abc")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteGroup(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid group ID\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrNotFound", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("DeleteGroup", 99).Return(services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/groups/99", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "99")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteGroup(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "Group not found\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockGroupService)
		handler := NewGroupHandler(mockService)

		mockService.On("DeleteGroup", 1).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/groups/1", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.DeleteGroup(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())

		mockService.AssertExpectations(t)
	})
}
