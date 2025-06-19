package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/internal/testutils"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateChild(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		inputChild := models.Child{
			FirstName:   "Test",
			LastName:    "Child",
			Birthdate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:      "male",
			Parent1Name: testutils.StringPtr("Parent One"),
			Parent2Name: testutils.StringPtr("Parent Two"),
			GroupID:     testutils.IntPtr(1),
		}

		returnedChild := models.Child{
			ID:          1,
			FirstName:   "Test",
			LastName:    "Child",
			Birthdate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender: "male",
			Parent1Name: testutils.StringPtr("Parent One"),
			Parent2Name: testutils.StringPtr("Parent Two"),
			GroupID:     testutils.IntPtr(1),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockChildService.On("CreateChild", mock.AnythingOfType("*models.Child")).Return(&returnedChild, nil).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPost, "/children", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateChild(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)

		assert.Equal(t, float64(1), responseBody["id"])
		assert.Equal(t, "Test", responseBody["first_name"])
		assert.Equal(t, "Child", responseBody["last_name"])
		assert.Equal(t, "male", responseBody["gender"])
		assert.Equal(t, "Parent One", responseBody["parent1_name"])
		assert.Equal(t, "Parent Two", responseBody["parent2_name"])
		assert.Equal(t, float64(1), responseBody["group_id"])

		expectedTime, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
		actualTime, _ := time.Parse(time.RFC3339, responseBody["birthdate"].(string))
		assert.True(t, expectedTime.Equal(actualTime), "Birthdate mismatch")

		assert.Contains(t, responseBody, "created_at")
		assert.Contains(t, responseBody, "updated_at")

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		inputChild := models.Child{
			FirstName: "",
		}
		mockChildService.On("CreateChild", mock.AnythingOfType("*models.Child")).Return(nil, services.ErrInvalidInput).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPost, "/children", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Invalid child data provided\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		inputChild := models.Child{
			FirstName: "Error",
			Birthdate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "Female",
		}
		mockChildService.On("CreateChild", mock.AnythingOfType("*models.Child")).Return(nil, errors.New("database error")).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPost, "/children", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateChild(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		req := httptest.NewRequest(http.MethodPost, "/children", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		handler.CreateChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockChildService.AssertExpectations(t)
	})
}

func TestGetAllChildren(t *testing.T) {
	t.Run("Successful Retrieval", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		mockChildService.On("GetAllChildren").Return([]models.Child{
			{ID: 1, FirstName: "Child A", Birthdate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), Gender: "Male"},
			{ID: 2, FirstName: "Child B", Birthdate: time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC), Gender: "Female"},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/children", nil)
		rr := httptest.NewRecorder()

		handler.GetAllChildren(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody []models.Child
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, []models.Child{
			{ID: 1, FirstName: "Child A", Birthdate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), Gender: "Male"},
			{ID: 2, FirstName: "Child B", Birthdate: time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC), Gender: "Female"},
		}, responseBody)

		mockChildService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		mockChildService.On("GetAllChildren").Return([]models.Child{}, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/children", nil)
		rr := httptest.NewRecorder()

		handler.GetAllChildren(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")

		mockChildService.AssertExpectations(t)
	})
}

func TestGetChildByID(t *testing.T) {
	t.Run("Successful Retrieval", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		mockChildService.On("GetChildByID", 1).Return(&models.Child{
			ID:        1,
			FirstName: "Test Child",
			Birthdate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "Male",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/children/1", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "1")
		rr := httptest.NewRecorder()

		handler.GetChildByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)

		assert.Equal(t, float64(1), responseBody["id"])
		assert.Equal(t, "Test Child", responseBody["first_name"])
		assert.Equal(t, "Male", responseBody["gender"])

		expectedTime, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
		actualTime, _ := time.Parse(time.RFC3339, responseBody["birthdate"].(string))
		assert.True(t, expectedTime.Equal(actualTime), "Birthdate mismatch")

		mockChildService.AssertExpectations(t)
	})

	t.Run("Child Not Found", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		mockChildService.On("GetChildByID", 99).Return(nil, services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/children/99", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "99")
		rr := httptest.NewRecorder()

		handler.GetChildByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "Child not found\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		mockChildService.On("GetChildByID", 1).Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/children/1", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "1")
		rr := httptest.NewRecorder()

		handler.GetChildByID(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid Child ID in Path", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		req := httptest.NewRequest(http.MethodGet, "/children/abc", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "abc")
		rr := httptest.NewRecorder()

		handler.GetChildByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid child ID")
		mockChildService.AssertExpectations(t)
	})
}

func TestUpdateChild(t *testing.T) {
	t.Run("Successful Update", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 1
		inputChild := models.Child{
			FirstName:   "Updated",
			LastName:    "Child",
			Birthdate:   time.Date(2019, 5, 10, 0, 0, 0, 0, time.UTC),
			Gender:      "Female",
			Parent1Name: testutils.StringPtr("New Parent"),
		}
		mockChildService.On("UpdateChild", mock.AnythingOfType("*models.Child")).Return(nil).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPut, "/children/"+strconv.Itoa(childID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))

		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, map[string]interface{}{"message": "Child updated successfully"}, responseBody)

		mockChildService.AssertExpectations(t)
	})

	t.Run("Child Not Found", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 99
		inputChild := models.Child{
			FirstName: "Non Existent",
		}
		mockChildService.On("UpdateChild", mock.AnythingOfType("*models.Child")).Return(services.ErrNotFound).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPut, "/children/"+strconv.Itoa(childID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))

		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "Child not found\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 1
		inputChild := models.Child{
			FirstName: "",
		}
		mockChildService.On("UpdateChild", mock.AnythingOfType("*models.Child")).Return(services.ErrInvalidInput).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPut, "/children/"+strconv.Itoa(childID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))

		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Invalid child data provided\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 1
		inputChild := models.Child{
			FirstName: "Error Child",
		}
		mockChildService.On("UpdateChild", mock.AnythingOfType("*models.Child")).Return(errors.New("database error")).Once()

		body, _ := json.Marshal(inputChild)
		req := httptest.NewRequest(http.MethodPut, "/children/"+strconv.Itoa(childID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))

		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid Child ID in Path", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		req := httptest.NewRequest(http.MethodPut, "/children/abc", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "abc")
		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid child ID")
		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		req := httptest.NewRequest(http.MethodPut, "/children/1", strings.NewReader("invalid json"))
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "1")
		rr := httptest.NewRecorder()

		handler.UpdateChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockChildService.AssertExpectations(t)
	})
}

func TestDeleteChild(t *testing.T) {
	t.Run("Successful Deletion", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 1
		mockChildService.On("DeleteChild", childID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/children/"+strconv.Itoa(childID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))
		rr := httptest.NewRecorder()

		handler.DeleteChild(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, map[string]interface{}{"message": "Child deleted successfully"}, responseBody)

		mockChildService.AssertExpectations(t)
	})

	t.Run("Child Not Found", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 99
		mockChildService.On("DeleteChild", childID).Return(services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/children/"+strconv.Itoa(childID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))
		rr := httptest.NewRecorder()

		handler.DeleteChild(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "Child not found\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		childID := 1
		mockChildService.On("DeleteChild", childID).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/children/"+strconv.Itoa(childID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", strconv.Itoa(childID))
		rr := httptest.NewRecorder()

		handler.DeleteChild(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockChildService.AssertExpectations(t)
	})

	t.Run("Invalid Child ID in Path", func(t *testing.T) {
		mockChildService := new(mocks.MockChildService)
		handler := NewChildHandler(mockChildService)

		req := httptest.NewRequest(http.MethodDelete, "/children/abc", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("child_id", "abc")
		rr := httptest.NewRecorder()

		handler.DeleteChild(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid child ID")
		mockChildService.AssertExpectations(t)
	})
}
