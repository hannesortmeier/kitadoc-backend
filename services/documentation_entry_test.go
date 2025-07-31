package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/services/mocks"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}
		expectedCategory := &models.Category{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(expectedCategory, nil).Once()
		mockDocumentationEntryStore.On("Create", mock.AnythingOfType("*models.DocumentationEntry")).Return(1, nil).Once()

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.NoError(t, err)
		assert.NotNil(t, createdEntry)
		assert.Equal(t, 1, createdEntry.ID)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID: 0, // Invalid ChildID
		}

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdEntry)
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Child not found
	t.Run("child not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                99, // Non-existent child
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}

		mockChildStore.On("GetByID", entry.ChildID).Return(nil, data.ErrNotFound).Once()

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child not found")
		assert.Nil(t, createdEntry)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Create")
	})

	// Test case 4: Teacher not found
	t.Run("teacher not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                1,
			TeacherID:              99, // Non-existent teacher
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}
		expectedChild := &models.Child{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(nil, data.ErrNotFound).Once()

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacher not found")
		assert.Nil(t, createdEntry)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Create")
	})

	// Test case 5: Category not found
	t.Run("category not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             99, // Non-existent category
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(nil, data.ErrNotFound).Once()

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
		assert.Nil(t, createdEntry)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "Create")
	})

	// Test case 6: Observation date in the future
	t.Run("future observation date", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(time.Hour), // Future date
			ObservationDescription: "Test observation",
		}

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input")
		assert.Nil(t, createdEntry)
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Create")
	})

	// Test case 7: Internal error during store creation
	t.Run("internal error on create", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}
		expectedCategory := &models.Category{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(expectedCategory, nil).Once()
		mockDocumentationEntryStore.On("Create", mock.AnythingOfType("*models.DocumentationEntry")).Return(0, errors.New("db error")).Once()

		createdEntry, err := service.CreateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdEntry)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}

func TestGetDocumentationEntryByID(t *testing.T) {
	mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
	mockChildStore := new(mocks.MockChildStore)
	mockTeacherStore := new(mocks.MockTeacherStore)
	mockCategoryStore := new(mocks.MockCategoryStore)
	mockUserStore := new(mocks.MockUserStore)
	service := services.NewDocumentationEntryService(
		mockDocumentationEntryStore,
		mockChildStore,
		mockTeacherStore,
		mockCategoryStore,
		mockUserStore,
	)

	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		entryID := 1
		expectedEntry := &models.DocumentationEntry{ID: entryID, ObservationDescription: "Test Content"}
		mockDocumentationEntryStore.On("GetByID", entryID).Return(expectedEntry, nil).Once()

		entry, err := service.GetDocumentationEntryByID(logger, ctx, entryID)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry.ID, entry.ID)
		assert.Equal(t, expectedEntry.ObservationDescription, entry.ObservationDescription)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Entry not found
	t.Run("not found", func(t *testing.T) {
		entryID := 99
		mockDocumentationEntryStore.On("GetByID", entryID).Return(nil, data.ErrNotFound).Once()

		entry, err := service.GetDocumentationEntryByID(logger, ctx, entryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, entry)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		entryID := 1
		mockDocumentationEntryStore.On("GetByID", entryID).Return(nil, errors.New("db error")).Once()

		entry, err := service.GetDocumentationEntryByID(logger, ctx, entryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, entry)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}

func TestUpdateDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Updated observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}
		expectedCategory := &models.Category{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(expectedCategory, nil).Once()
		mockDocumentationEntryStore.On("Update", mock.AnythingOfType("*models.DocumentationEntry")).Return(nil).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.NoError(t, err)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:        1,
			ChildID:   0, // Invalid ChildID
			TeacherID: 1,
		}

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Child not found
	t.Run("child not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                99, // Non-existent child
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Updated observation",
		}

		mockChildStore.On("GetByID", entry.ChildID).Return(nil, data.ErrNotFound).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child not found")
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Update")
	})

	// Test case 4: Teacher not found
	t.Run("teacher not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                1,
			TeacherID:              99, // Non-existent teacher
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Updated observation",
		}
		expectedChild := &models.Child{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(nil, data.ErrNotFound).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacher not found")
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Update")
	})

	// Test case 5: Category not found
	t.Run("category not found", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             99, // Non-existent category
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Test observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(nil, data.ErrNotFound).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "Update")
	})

	// Test case 6: Observation date in the future
	t.Run("future observation date", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(time.Hour).UTC(), // Future date
			ObservationDescription: "Test observation",
		}

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input")
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
		mockCategoryStore.AssertNotCalled(t, "GetByID")
		mockDocumentationEntryStore.AssertNotCalled(t, "Update")
	})

	// Test case 7: Entry not found during update
	t.Run("not found on update", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     99, // Non-existent ID
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Updated observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}
		expectedCategory := &models.Category{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(expectedCategory, nil).Once()
		mockDocumentationEntryStore.On("Update", mock.AnythingOfType("*models.DocumentationEntry")).Return(data.ErrNotFound).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 8: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockUserStore := new(mocks.MockUserStore)
		service := services.NewDocumentationEntryService(
			mockDocumentationEntryStore,
			mockChildStore,
			mockTeacherStore,
			mockCategoryStore,
			mockUserStore,
		)

		entry := &models.DocumentationEntry{
			ID:                     1,
			ChildID:                1,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        time.Now().Add(-time.Hour),
			ObservationDescription: "Updated observation",
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}
		expectedCategory := &models.Category{ID: 1}

		mockChildStore.On("GetByID", entry.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", entry.TeacherID).Return(expectedTeacher, nil).Once()
		mockCategoryStore.On("GetByID", entry.CategoryID).Return(expectedCategory, nil).Once()
		mockDocumentationEntryStore.On("Update", mock.AnythingOfType("*models.DocumentationEntry")).Return(errors.New("db error")).Once()

		err := service.UpdateDocumentationEntry(logger, ctx, entry)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockCategoryStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}

func TestDeleteDocumentationEntry(t *testing.T) {
	mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
	mockChildStore := new(mocks.MockChildStore)
	mockTeacherStore := new(mocks.MockTeacherStore)
	mockCategoryStore := new(mocks.MockCategoryStore)
	mockUserStore := new(mocks.MockUserStore)
	service := services.NewDocumentationEntryService(
		mockDocumentationEntryStore,
		mockChildStore,
		mockTeacherStore,
		mockCategoryStore,
		mockUserStore,
	)

	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		entryID := 1
		mockDocumentationEntryStore.On("Delete", entryID).Return(nil).Once()

		err := service.DeleteDocumentationEntry(logger, ctx, entryID)

		assert.NoError(t, err)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Entry not found
	t.Run("not found", func(t *testing.T) {
		entryID := 99
		mockDocumentationEntryStore.On("Delete", entryID).Return(data.ErrNotFound).Once()

		err := service.DeleteDocumentationEntry(logger, ctx, entryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		entryID := 1
		mockDocumentationEntryStore.On("Delete", entryID).Return(errors.New("db error")).Once()

		err := service.DeleteDocumentationEntry(logger, ctx, entryID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}

func TestGetAllDocumentationForChild(t *testing.T) {
	mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
	mockChildStore := new(mocks.MockChildStore)
	mockTeacherStore := new(mocks.MockTeacherStore)
	mockCategoryStore := new(mocks.MockCategoryStore)
	mockUserStore := new(mocks.MockUserStore)
	service := services.NewDocumentationEntryService(
		mockDocumentationEntryStore,
		mockChildStore,
		mockTeacherStore,
		mockCategoryStore,
		mockUserStore,
	)

	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{ID: childID}
		expectedEntries := []models.DocumentationEntry{
			{ID: 1, ChildID: childID, CategoryID: 1, ObservationDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), ObservationDescription: "Entry 1"},
			{ID: 2, ChildID: childID, CategoryID: 2, ObservationDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), ObservationDescription: "Entry 2"},
		}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockDocumentationEntryStore.On("GetAllForChild", childID).Return(expectedEntries, nil).Once()

		entries, err := service.GetAllDocumentationForChild(logger, ctx, childID)

		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Equal(t, expectedEntries, entries)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Child not found
	t.Run("child not found", func(t *testing.T) {
		childID := 99
		mockChildStore.On("GetByID", childID).Return(nil, data.ErrNotFound).Once()

		entries, err := service.GetAllDocumentationForChild(logger, ctx, childID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child not found")
		assert.Nil(t, entries)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "GetAllForChild")
	})

	// Test case 3: Internal error during child fetch
	t.Run("internal error on child fetch", func(t *testing.T) {
		childID := 1
		mockChildStore.On("GetByID", childID).Return(nil, errors.New("db error")).Once()

		entries, err := service.GetAllDocumentationForChild(logger, ctx, childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, entries)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "GetAllForChild")
	})

	// Test case 4: Internal error during entry fetch
	t.Run("internal error on entry fetch", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{ID: childID}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockDocumentationEntryStore.On("GetAllForChild", childID).Return(nil, errors.New("db error")).Once()

		entries, err := service.GetAllDocumentationForChild(logger, ctx, childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, entries)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}

func TestApproveDocumentationEntry(t *testing.T) {
	mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
	mockChildStore := new(mocks.MockChildStore)
	mockTeacherStore := new(mocks.MockTeacherStore)
	mockCategoryStore := new(mocks.MockCategoryStore)
	mockUserStore := new(mocks.MockUserStore)
	service := services.NewDocumentationEntryService(
		mockDocumentationEntryStore,
		mockChildStore,
		mockTeacherStore,
		mockCategoryStore,
		mockUserStore,
	)

	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful approval
	t.Run("success", func(t *testing.T) {
		entryID := 1
		approvedByTeacherID := 1
		existingEntry := &models.DocumentationEntry{ID: entryID, IsApproved: false}
		approvingUser := &models.Teacher{ID: approvedByTeacherID}

		mockDocumentationEntryStore.On("GetByID", entryID).Return(existingEntry, nil).Once()
		mockTeacherStore.On("GetByID", approvedByTeacherID).Return(approvingUser, nil).Once()
		mockDocumentationEntryStore.On("ApproveEntry", entryID, approvedByTeacherID).Return(nil).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByTeacherID)

		assert.NoError(t, err)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Entry not found
	t.Run("entry not found", func(t *testing.T) {
		entryID := 99
		approvedByUserID := 1
		mockDocumentationEntryStore.On("GetByID", entryID).Return(nil, data.ErrNotFound).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByUserID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockUserStore.AssertNotCalled(t, "GetByID")
	})

	// Test case 3: Internal error during entry fetch
	t.Run("internal error on entry fetch", func(t *testing.T) {
		entryID := 1
		approvedByUserID := 1
		mockDocumentationEntryStore.On("GetByID", entryID).Return(nil, errors.New("db error")).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByUserID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockUserStore.AssertNotCalled(t, "GetByID")
	})

	// Test case 4: Approving user not found
	t.Run("approving user not found", func(t *testing.T) {
		entryID := 1
		approvedByTeacherID := 99
		existingEntry := &models.DocumentationEntry{ID: entryID, IsApproved: false}

		mockDocumentationEntryStore.On("GetByID", entryID).Return(existingEntry, nil).Once()
		mockTeacherStore.On("GetByID", approvedByTeacherID).Return(nil, data.ErrNotFound).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByTeacherID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "approving teacher not found")
		mockDocumentationEntryStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "ApproveEntry")
	})

	// Test case 5: Internal error during teacher fetch
	t.Run("internal error on teacher fetch", func(t *testing.T) {
		entryID := 1
		approvedByTeacherID := 1
		existingEntry := &models.DocumentationEntry{ID: entryID, IsApproved: false}

		mockDocumentationEntryStore.On("GetByID", entryID).Return(existingEntry, nil).Once()
		mockTeacherStore.On("GetByID", approvedByTeacherID).Return(nil, errors.New("db error")).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByTeacherID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "ApproveEntry")
	})

	// Test case 6: Entry already approved
	t.Run("already approved", func(t *testing.T) {
		entryID := 1
		approvedByTeacherID := 1
		existingEntry := &models.DocumentationEntry{ID: entryID, IsApproved: true} // Already approved
		approvingTeacher := &models.Teacher{ID: approvedByTeacherID}

		mockDocumentationEntryStore.On("GetByID", entryID).Return(existingEntry, nil).Once()
		mockTeacherStore.On("GetByID", approvedByTeacherID).Return(approvingTeacher, nil).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByTeacherID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "documentation entry is already approved")
		mockDocumentationEntryStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "ApproveEntry")
	})

	// Test case 7: Internal error during approval
	t.Run("internal error on approve", func(t *testing.T) {
		entryID := 1
		approvedByTeacherID := 1
		existingEntry := &models.DocumentationEntry{ID: entryID, IsApproved: false}
		approvingTeacher := &models.Teacher{ID: approvedByTeacherID}

		mockDocumentationEntryStore.On("GetByID", entryID).Return(existingEntry, nil).Once()
		mockTeacherStore.On("GetByID", approvedByTeacherID).Return(approvingTeacher, nil).Once()
		mockDocumentationEntryStore.On("ApproveEntry", entryID, approvedByTeacherID).Return(errors.New("db error")).Once()

		err := service.ApproveDocumentationEntry(logger, ctx, entryID, approvedByTeacherID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})
}

func TestGenerateChildReport(t *testing.T) {
	mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
	mockChildStore := new(mocks.MockChildStore)
	mockTeacherStore := new(mocks.MockTeacherStore)
	mockCategoryStore := new(mocks.MockCategoryStore)
	mockUserStore := new(mocks.MockUserStore)
	service := services.NewDocumentationEntryService(
		mockDocumentationEntryStore,
		mockChildStore,
		mockTeacherStore,
		mockCategoryStore,
		mockUserStore,
	)

	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	// Test case 1: Successful report generation with entries
	t.Run("success with entries", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{
			ID:                       childID,
			FirstName:                "Report",
			LastName:                 "Child",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Parent One",
			Parent2Name:              "Parent Two",
		}
		expectedEntries := []models.DocumentationEntry{
			{ID: 1, ChildID: childID, CategoryID: 1, ObservationDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), ObservationDescription: "Entry 1"},
			{ID: 2, ChildID: childID, CategoryID: 2, ObservationDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), ObservationDescription: "Entry 2"},
		}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockDocumentationEntryStore.On("GetAllForChild", childID).Return(expectedEntries, nil).Once()

		reportBytes, err := service.GenerateChildReport(logger, ctx, childID, []models.Assignment{})

		assert.NoError(t, err)
		assert.NotNil(t, reportBytes)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 2: Successful report generation with no entries
	t.Run("success with no entries", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{
			ID:                       childID,
			FirstName:                "Report",
			LastName:                 "Child",
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            time.Now(),
			ExpectedSchoolEnrollment: time.Now().AddDate(1, 0, 0),
			Address:                  "123 Main St",
			Parent1Name:              "Parent One",
			Parent2Name:              "Parent Two",
		}
		expectedEntries := []models.DocumentationEntry{}

		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockDocumentationEntryStore.On("GetAllForChild", childID).Return(expectedEntries, nil).Once()

		reportBytes, err := service.GenerateChildReport(logger, ctx, childID, []models.Assignment{})

		assert.NoError(t, err)
		assert.NotNil(t, reportBytes)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})

	// Test case 3: Child not found
	t.Run("child not found", func(t *testing.T) {
		childID := 99
		mockChildStore.On("GetByID", childID).Return(nil, data.ErrNotFound).Once()

		reportBytes, err := service.GenerateChildReport(logger, ctx, childID, []models.Assignment{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, reportBytes)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "GetAllForChild")
	})

	// Test case 4: Internal error during child fetch
	t.Run("internal error on child fetch", func(t *testing.T) {
		childID := 1
		mockChildStore.On("GetByID", childID).Return(nil, errors.New("db error")).Once()

		reportBytes, err := service.GenerateChildReport(logger, ctx, childID, []models.Assignment{})

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, reportBytes)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertNotCalled(t, "GetAllForChild")
	})

	// Test case 5: Internal error during entries fetch
	t.Run("internal error on entries fetch", func(t *testing.T) {
		childID := 1
		expectedChild := &models.Child{ID: childID}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockDocumentationEntryStore.On("GetAllForChild", childID).Return(nil, errors.New("db error")).Once()

		reportBytes, err := service.GenerateChildReport(logger, ctx, childID, []models.Assignment{})

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, reportBytes)
		mockChildStore.AssertExpectations(t)
		mockDocumentationEntryStore.AssertExpectations(t)
	})
}
