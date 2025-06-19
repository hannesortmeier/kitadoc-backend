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

	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryService is a mock implementation of services.CategoryService
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(category *models.Category) (*models.Category, error) {
	args := m.Called(category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryByID(id int) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) UpdateCategory(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryService) DeleteCategory(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCategoryService) GetAllCategories() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func TestCreateCategory(t *testing.T) {
	mockCategoryService := new(MockCategoryService)
	handler := NewCategoryHandler(mockCategoryService)

	tests := []struct {
		name           string
		inputCategory  models.Category
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful Creation",
			inputCategory: models.Category{
				Name:        "Test Category",
				Description: models.StringPtr("A category for testing"),
			},
			setupMocks: func() {
				mockCategoryService.On("CreateCategory", mock.AnythingOfType("*models.Category")).Return(&models.Category{
					ID:          1,
					Name:        "Test Category",
					Description: models.StringPtr("A category for testing"),
				}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"created_at":  "0001-01-01T00:00:00Z",
				"id":          float64(1),
				"name":        "Test Category",
				"description": "A category for testing",
				"updated_at":  "0001-01-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryService = new(MockCategoryService)
			handler = NewCategoryHandler(mockCategoryService)
			tt.setupMocks()

			body, _ := json.Marshal(tt.inputCategory)
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.CreateCategory(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)

			mockCategoryService.AssertExpectations(t)
		})
	}

	t.Run("Internal Server Error", func(t *testing.T) {
		mockCategoryService = new(MockCategoryService)
		handler = NewCategoryHandler(mockCategoryService)

		var category = models.Category{
			Name:        "Another Category",
			Description: models.StringPtr("Description"),
		}

		mockCategoryService.On("CreateCategory", mock.AnythingOfType("*models.Category")).Return(nil, errors.New("database error")).Once()

		body, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateCategory(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())
		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid input", func(t *testing.T) {
		mockCategoryService = new(MockCategoryService)
		handler = NewCategoryHandler(mockCategoryService)

		var category = models.Category{
			Name:        "", // Invalid name
			Description: models.StringPtr("Description"),
		}

		mockCategoryService.On("CreateCategory", mock.AnythingOfType("*models.Category")).Return(nil, services.ErrInvalidInput).Once()

		body, _ := json.Marshal(category)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.CreateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Invalid category data provided\n", rr.Body.String())
		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockCategoryService = new(MockCategoryService)
		handler = NewCategoryHandler(mockCategoryService)

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		handler.CreateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockCategoryService.AssertExpectations(t)
	})
}

func TestGetAllCategories(t *testing.T) {
	mockCategoryService := new(MockCategoryService)
	handler := NewCategoryHandler(mockCategoryService)

	tests := []struct {
		name           string
		setupMocks     func()
		expectedStatus int
		expectedBody   []models.Category
	}{
		{
			name: "Successful Retrieval",
			setupMocks: func() {
				mockCategoryService.On("GetAllCategories").Return([]models.Category{
					{ID: 1, Name: "Category A", Description: models.StringPtr("Desc A")},
					{ID: 2, Name: "Category B", Description: models.StringPtr("Desc B")},
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: []models.Category{
				{ID: 1, Name: "Category A", Description: models.StringPtr("Desc A")},
				{ID: 2, Name: "Category B", Description: models.StringPtr("Desc B")},
			},
		},
		{
			name: "Internal Server Error",
			setupMocks: func() {
				mockCategoryService.On("GetAllCategories").Return([]models.Category{}, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil, // Body will be an error message string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryService = new(MockCategoryService)
			handler = NewCategoryHandler(mockCategoryService)
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
			rr := httptest.NewRecorder()

			handler.GetAllCategories(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var responseBody []models.Category
				json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.Equal(t, tt.expectedBody, responseBody)
			} else {
				assert.Contains(t, rr.Body.String(), "Internal server error")
			}

			mockCategoryService.AssertExpectations(t)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	t.Run("Successful Update", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 1
		inputCategory := models.Category{
			Name:        "Updated Category",
			Description: models.StringPtr("Updated description"),
		}

		mockCategoryService.On("UpdateCategory", mock.AnythingOfType("*models.Category")).Return(nil).Once()

		body, _ := json.Marshal(inputCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories/"+strconv.Itoa(categoryID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))

		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, map[string]interface{}{"message": "Category updated successfully"}, responseBody)

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Category Not Found", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 99
		inputCategory := models.Category{
			Name: "Non Existent",
		}

		mockCategoryService.On("UpdateCategory", mock.AnythingOfType("*models.Category")).Return(services.ErrNotFound).Once()

		body, _ := json.Marshal(inputCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories/"+strconv.Itoa(categoryID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))

		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "Category not found\n", rr.Body.String())

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 1
		inputCategory := models.Category{
			Name: "",
		}

		mockCategoryService.On("UpdateCategory", mock.AnythingOfType("*models.Category")).Return(services.ErrInvalidInput).Once()

		body, _ := json.Marshal(inputCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories/"+strconv.Itoa(categoryID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))

		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Invalid category data provided\n", rr.Body.String())

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 1
		inputCategory := models.Category{
			Name: "Error Category",
		}

		mockCategoryService.On("UpdateCategory", mock.AnythingOfType("*models.Category")).Return(errors.New("database error")).Once()

		body, _ := json.Marshal(inputCategory)
		req := httptest.NewRequest(http.MethodPut, "/categories/"+strconv.Itoa(categoryID), bytes.NewBuffer(body))
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))

		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid Category ID in Path", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		req := httptest.NewRequest(http.MethodPut, "/categories/abc", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", "abc")
		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid category ID")
		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid JSON Payload", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		req := httptest.NewRequest(http.MethodPut, "/categories/1", strings.NewReader("invalid json"))
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", "1")
		rr := httptest.NewRecorder()

		handler.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockCategoryService.AssertExpectations(t)
	})
}

func TestDeleteCategory(t *testing.T) {
	t.Run("Successful Deletion", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 1
		mockCategoryService.On("DeleteCategory", categoryID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/categories/"+strconv.Itoa(categoryID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))
		rr := httptest.NewRecorder()

		handler.DeleteCategory(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, map[string]interface{}{"message": "Category deleted successfully"}, responseBody)

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Category Not Found", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 99
		mockCategoryService.On("DeleteCategory", categoryID).Return(services.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/categories/"+strconv.Itoa(categoryID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))
		rr := httptest.NewRecorder()

		handler.DeleteCategory(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "Category not found\n", rr.Body.String())

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		categoryID := 1
		mockCategoryService.On("DeleteCategory", categoryID).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/categories/"+strconv.Itoa(categoryID), nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", strconv.Itoa(categoryID))
		rr := httptest.NewRecorder()

		handler.DeleteCategory(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Internal server error\n", rr.Body.String())

		mockCategoryService.AssertExpectations(t)
	})

	t.Run("Invalid Category ID in Path", func(t *testing.T) {
		mockCategoryService := new(MockCategoryService)
		handler := NewCategoryHandler(mockCategoryService)

		req := httptest.NewRequest(http.MethodDelete, "/categories/abc", nil)
		req = req.WithContext(req.Context())
		req.SetPathValue("category_id", "abc")
		rr := httptest.NewRecorder()

		handler.DeleteCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid category ID")
		mockCategoryService.AssertExpectations(t)
	})
}
