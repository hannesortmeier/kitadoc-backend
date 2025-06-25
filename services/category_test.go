package services_test

import (
	"errors"
	"testing"

	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/services/mocks"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategory(t *testing.T) {
	mockCategoryStore := new(mocks.MockCategoryStore)
	service := services.NewCategoryService(mockCategoryStore)

	log_level, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		category := &models.Category{Name: "New Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, data.ErrNotFound).Once()
		mockCategoryStore.On("Create", category).Return(1, nil).Once()

		createdCategory, err := service.CreateCategory(category)

		assert.NoError(t, err)
		assert.NotNil(t, createdCategory)
		assert.Equal(t, 1, createdCategory.ID)
		assert.Equal(t, "New Category", createdCategory.Name)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		category := &models.Category{Name: ""} // Invalid name
		createdCategory, err := service.CreateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdCategory)
		mockCategoryStore.AssertNotCalled(t, "GetByName")
		mockCategoryStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Category with same name already exists
	t.Run("already exists", func(t *testing.T) {
		category := &models.Category{Name: "Existing Category"}
		existingCategory := &models.Category{ID: 1, Name: "Existing Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(existingCategory, nil).Once()

		createdCategory, err := service.CreateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrAlreadyExists, err)
		assert.Nil(t, createdCategory)
		mockCategoryStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "Create")
	})

	// Test case 4: Internal error during GetByName check
	t.Run("internal error on GetByName", func(t *testing.T) {
		category := &models.Category{Name: "New Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, errors.New("db error")).Once()

		createdCategory, err := service.CreateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdCategory)
		mockCategoryStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "Create")
	})

	// Test case 5: Internal error during creation
	t.Run("internal error on create", func(t *testing.T) {
		category := &models.Category{Name: "New Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, data.ErrNotFound).Once()
		mockCategoryStore.On("Create", category).Return(0, errors.New("db error")).Once()

		createdCategory, err := service.CreateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdCategory)
		mockCategoryStore.AssertExpectations(t)
	})
}

func TestGetCategoryByID(t *testing.T) {
	mockCategoryStore := new(mocks.MockCategoryStore)
	service := services.NewCategoryService(mockCategoryStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		categoryID := 1
		expectedCategory := &models.Category{ID: categoryID, Name: "Test Category"}
		mockCategoryStore.On("GetByID", categoryID).Return(expectedCategory, nil).Once()

		category, err := service.GetCategoryByID(categoryID)

		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, expectedCategory.ID, category.ID)
		assert.Equal(t, expectedCategory.Name, category.Name)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 2: Category not found
	t.Run("not found", func(t *testing.T) {
		categoryID := 99
		mockCategoryStore.On("GetByID", categoryID).Return(nil, data.ErrNotFound).Once()

		category, err := service.GetCategoryByID(categoryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, category)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		categoryID := 1
		mockCategoryStore.On("GetByID", categoryID).Return(nil, errors.New("db error")).Once()

		category, err := service.GetCategoryByID(categoryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, category)
		mockCategoryStore.AssertExpectations(t)
	})
}

func TestUpdateCategory(t *testing.T) {
	mockCategoryStore := new(mocks.MockCategoryStore)
	service := services.NewCategoryService(mockCategoryStore)

	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		category := &models.Category{ID: 1, Name: "Updated Category"}
		// Simulate that the name is unique or the same as the existing one
		mockCategoryStore.On("GetByName", category.Name).Return(nil, data.ErrNotFound).Once()
		mockCategoryStore.On("Update", category).Return(nil).Once()

		err := service.UpdateCategory(category)

		assert.NoError(t, err)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		category := &models.Category{ID: 1, Name: ""} // Invalid name
		err := service.UpdateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockCategoryStore.AssertNotCalled(t, "GetByName")
		mockCategoryStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Category name conflict
	t.Run("name conflict", func(t *testing.T) {
		category := &models.Category{ID: 1, Name: "New Name"}
		conflictingCategory := &models.Category{ID: 2, Name: "New Name"}
		mockCategoryStore.On("GetByName", category.Name).Return(conflictingCategory, nil).Once()

		err := service.UpdateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrAlreadyExists, err)
		mockCategoryStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "Update")
	})

	// Test case 4: Internal error during GetByName check
	t.Run("internal error on GetByName", func(t *testing.T) {
		category := &models.Category{ID: 1, Name: "Updated Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, errors.New("db error")).Once()

		err := service.UpdateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockCategoryStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "Update")
	})

	// Test case 5: Category not found during update
	t.Run("not found on update", func(t *testing.T) {
		category := &models.Category{ID: 99, Name: "Updated Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, data.ErrNotFound).Once()
		mockCategoryStore.On("Update", category).Return(data.ErrNotFound).Once()

		err := service.UpdateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 6: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		category := &models.Category{ID: 1, Name: "Updated Category"}
		mockCategoryStore.On("GetByName", category.Name).Return(nil, data.ErrNotFound).Once()
		mockCategoryStore.On("Update", category).Return(errors.New("db error")).Once()

		err := service.UpdateCategory(category)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockCategoryStore.AssertExpectations(t)
	})
}

func TestDeleteCategory(t *testing.T) {
	mockCategoryStore := new(mocks.MockCategoryStore)
	service := services.NewCategoryService(mockCategoryStore)

	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		categoryID := 1
		mockCategoryStore.On("Delete", categoryID).Return(nil).Once()

		err := service.DeleteCategory(categoryID)

		assert.NoError(t, err)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 2: Category not found
	t.Run("not found", func(t *testing.T) {
		categoryID := 99
		mockCategoryStore.On("Delete", categoryID).Return(data.ErrNotFound).Once()

		err := service.DeleteCategory(categoryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		categoryID := 1
		mockCategoryStore.On("Delete", categoryID).Return(errors.New("db error")).Once()

		err := service.DeleteCategory(categoryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockCategoryStore.AssertExpectations(t)
	})
}

func TestGetAllCategories(t *testing.T) {
	mockCategoryStore := new(mocks.MockCategoryStore)
	service := services.NewCategoryService(mockCategoryStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		expectedCategories := []models.Category{
			{ID: 1, Name: "Category A"},
			{ID: 2, Name: "Category B"},
		}
		mockCategoryStore.On("GetAll").Return(expectedCategories, nil).Once()

		categories, err := service.GetAllCategories()

		assert.NoError(t, err)
		assert.NotNil(t, categories)
		assert.Equal(t, expectedCategories, categories)
		mockCategoryStore.AssertExpectations(t)
	})

	// Test case 2: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockCategoryStore.On("GetAll").Return(nil, errors.New("db error")).Once()

		categories, err := service.GetAllCategories()

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, categories)
		mockCategoryStore.AssertExpectations(t)
	})
}
