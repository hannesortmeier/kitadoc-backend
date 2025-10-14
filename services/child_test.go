package services_test

import (
	"errors"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/data/mocks"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateChild(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	log_level, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		child := &models.Child{
			FirstName:                "John",
			LastName:                 "Doe",
			Birthdate:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:                   "male",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Jane Doe",
			Parent2Name:              "John Doe Sr.",
		}
		mockChildStore.On("Create", mock.AnythingOfType("*models.Child")).Return(1, nil).Once()

		createdChild, err := service.CreateChild(child)

		assert.NoError(t, err)
		assert.NotNil(t, createdChild)
		assert.Equal(t, 1, createdChild.ID)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		child := &models.Child{
			FirstName: "", // Invalid empty first name
			LastName:  "Doe",
			Birthdate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "male",
		}

		createdChild, err := service.CreateChild(child)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdChild)
		mockChildStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Internal error during child creation
	t.Run("internal error on create child", func(t *testing.T) {
		child := &models.Child{
			FirstName:                "John",
			LastName:                 "Doe",
			Birthdate:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:                   "male",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Jane Doe",
			Parent2Name:              "John Doe Sr.",
		}
		mockChildStore.On("Create", mock.AnythingOfType("*models.Child")).Return(0, errors.New("db error")).Once()

		createdChild, err := service.CreateChild(child)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdChild)
		mockChildStore.AssertExpectations(t)
	})
}

func TestGetChildByID(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{ID: childID, FirstName: "Test Child"}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()

		child, err := service.GetChildByID(childID)

		assert.NoError(t, err)
		assert.NotNil(t, child)
		assert.Equal(t, expectedChild.ID, child.ID)
		assert.Equal(t, expectedChild.FirstName, child.FirstName)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 2: Child not found
	t.Run("not found", func(t *testing.T) {
		childID := 99
		mockChildStore.On("GetByID", childID).Return(nil, data.ErrNotFound).Once()

		child, err := service.GetChildByID(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, child)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		childID := 1
		mockChildStore.On("GetByID", childID).Return(nil, errors.New("db error")).Once()

		child, err := service.GetChildByID(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, child)
		mockChildStore.AssertExpectations(t)
	})
}

func TestUpdateChild(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		child := &models.Child{
			ID:                       1,
			FirstName:                "Updated John",
			LastName:                 "Doe",
			Birthdate:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:                   "male",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Jane Doe",
			Parent2Name:              "John Doe Sr.",
		}
		mockChildStore.On("Update", mock.AnythingOfType("*models.Child")).Return(nil).Once()

		err := service.UpdateChild(child)

		assert.NoError(t, err)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		child := &models.Child{
			ID:        1,
			FirstName: "", // Invalid: empty first name
			LastName:  "Doe",
			Birthdate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "male",
		}

		err := service.UpdateChild(child)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockChildStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Child not found during update
	t.Run("child not found on update", func(t *testing.T) {
		child := &models.Child{
			ID:                       99, // Non-existent ID
			FirstName:                "Updated John",
			LastName:                 "Doe",
			Birthdate:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:                   "male",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Jane Doe",
			Parent2Name:              "John Doe Sr.",
		}
		mockChildStore.On("Update", mock.AnythingOfType("*models.Child")).Return(data.ErrNotFound).Once()

		err := service.UpdateChild(child)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 4: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		child := &models.Child{
			ID:                       1,
			FirstName:                "Updated John",
			LastName:                 "Doe",
			Birthdate:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:                   "male",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Jane Doe",
			Parent2Name:              "John Doe Sr.",
		}
		mockChildStore.On("Update", mock.AnythingOfType("*models.Child")).Return(errors.New("db error")).Once()

		err := service.UpdateChild(child)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockChildStore.AssertExpectations(t)
	})
}

func TestDeleteChild(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		childID := 1
		mockChildStore.On("Delete", childID).Return(nil).Once()

		err := service.DeleteChild(childID)

		assert.NoError(t, err)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 2: Child not found
	t.Run("not found", func(t *testing.T) {
		childID := 99
		mockChildStore.On("Delete", childID).Return(data.ErrNotFound).Once()

		err := service.DeleteChild(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		childID := 1
		mockChildStore.On("Delete", childID).Return(errors.New("db error")).Once()

		err := service.DeleteChild(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockChildStore.AssertExpectations(t)
	})
}

func TestGetAllChildren(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		expectedChildren := []models.Child{
			{ID: 1, FirstName: "Child A"},
			{ID: 2, FirstName: "Child B"},
		}
		mockChildStore.On("GetAll").Return(expectedChildren, nil).Once()

		children, err := service.GetAllChildren()

		assert.NoError(t, err)
		assert.NotNil(t, children)
		assert.Equal(t, expectedChildren, children)
		mockChildStore.AssertExpectations(t)
	})

	// Test case 2: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockChildStore.On("GetAll").Return(nil, errors.New("db error")).Once()

		children, err := service.GetAllChildren()

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, children)
		mockChildStore.AssertExpectations(t)
	})
}

func TestBulkImportChildren(t *testing.T) {
	mockChildStore := new(mocks.MockChildStore)
	service := services.NewChildService(mockChildStore)

	// Test case 1: Placeholder for bulk import
	t.Run("placeholder", func(t *testing.T) {
		fileContent := []byte("dummy,csv,data")
		err := service.BulkImportChildren(fileContent)

		assert.Error(t, err)
		assert.Equal(t, services.ErrBulkImportFailed, err)
		mockChildStore.AssertNotCalled(t, "Create") // Should not call Create in this placeholder
	})
}
