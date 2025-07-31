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
	"github.com/stretchr/testify/mock"
)

func TestCreateTeacher(t *testing.T) {
	mockTeacherStore := new(mocks.MockTeacherStore)
	service := services.NewTeacherService(mockTeacherStore)

	log_level, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		teacher := &models.Teacher{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
		}
		mockTeacherStore.On("Create", mock.AnythingOfType("*models.Teacher")).Return(1, nil).Once()

		createdTeacher, err := service.CreateTeacher(teacher)

		assert.NoError(t, err)
		assert.NotNil(t, createdTeacher)
		assert.Equal(t, 1, createdTeacher.ID)
		assert.Equal(t, "John", createdTeacher.FirstName)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		teacher := &models.Teacher{
			FirstName: "", // Invalid: empty first name
			LastName:  "Doe",
			Username:  "johndoe",
		}

		createdTeacher, err := service.CreateTeacher(teacher)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdTeacher)
		mockTeacherStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Internal error during creation
	t.Run("internal error on create", func(t *testing.T) {
		teacher := &models.Teacher{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
		}
		mockTeacherStore.On("Create", mock.AnythingOfType("*models.Teacher")).Return(0, errors.New("db error")).Once()

		createdTeacher, err := service.CreateTeacher(teacher)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdTeacher)
		mockTeacherStore.AssertExpectations(t)
	})
}

func TestGetTeacherByID(t *testing.T) {
	mockTeacherStore := new(mocks.MockTeacherStore)
	service := services.NewTeacherService(mockTeacherStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		teacherID := 1
		expectedTeacher := &models.Teacher{ID: teacherID, FirstName: "Test Teacher", Username: "testteacher"}
		mockTeacherStore.On("GetByID", teacherID).Return(expectedTeacher, nil).Once()

		teacher, err := service.GetTeacherByID(teacherID)

		assert.NoError(t, err)
		assert.NotNil(t, teacher)
		assert.Equal(t, expectedTeacher.ID, teacher.ID)
		assert.Equal(t, expectedTeacher.FirstName, teacher.FirstName)
		assert.Equal(t, expectedTeacher.Username, teacher.Username)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Teacher not found
	t.Run("not found", func(t *testing.T) {
		teacherID := 99
		mockTeacherStore.On("GetByID", teacherID).Return(nil, data.ErrNotFound).Once()

		teacher, err := service.GetTeacherByID(teacherID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, teacher)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		teacherID := 1
		mockTeacherStore.On("GetByID", teacherID).Return(nil, errors.New("db error")).Once()

		teacher, err := service.GetTeacherByID(teacherID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, teacher)
		mockTeacherStore.AssertExpectations(t)
	})
}

func TestUpdateTeacher(t *testing.T) {
	mockTeacherStore := new(mocks.MockTeacherStore)
	service := services.NewTeacherService(mockTeacherStore)

	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		teacher := &models.Teacher{
			ID:        1,
			FirstName: "Updated John",
			LastName:  "Doe",
			Username:  "johndoe",
		}
		mockTeacherStore.On("Update", mock.AnythingOfType("*models.Teacher")).Return(nil).Once()

		err := service.UpdateTeacher(teacher)

		assert.NoError(t, err)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		teacher := &models.Teacher{
			ID:        1,
			FirstName: "", // Invalid: empty first name
			LastName:  "Doe",
			Username:  "johndoe",
		}

		err := service.UpdateTeacher(teacher)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockTeacherStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Teacher not found during update
	t.Run("not found on update", func(t *testing.T) {
		teacher := &models.Teacher{
			ID:        99, // Non-existent ID
			FirstName: "Updated John",
			LastName:  "Doe",
			Username:  "johndoe",
		}
		mockTeacherStore.On("Update", mock.AnythingOfType("*models.Teacher")).Return(data.ErrNotFound).Once()

		err := service.UpdateTeacher(teacher)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 4: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		teacher := &models.Teacher{
			ID:        1,
			FirstName: "Updated John",
			LastName:  "Doe",
			Username:  "johndoe",
		}
		mockTeacherStore.On("Update", mock.AnythingOfType("*models.Teacher")).Return(errors.New("db error")).Once()

		err := service.UpdateTeacher(teacher)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockTeacherStore.AssertExpectations(t)
	})
}

func TestGetAllTeachers(t *testing.T) {
	mockTeacherStore := new(mocks.MockTeacherStore)
	service := services.NewTeacherService(mockTeacherStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		expectedTeachers := []models.Teacher{
			{ID: 1, FirstName: "Teacher A", Username: "teachera"},
			{ID: 2, FirstName: "Teacher B", Username: "teacherb"},
		}
		mockTeacherStore.On("GetAll").Return(expectedTeachers, nil).Once()

		teachers, err := service.GetAllTeachers()

		assert.NoError(t, err)
		assert.NotNil(t, teachers)
		assert.Equal(t, expectedTeachers, teachers)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockTeacherStore.On("GetAll").Return(nil, errors.New("db error")).Once()

		teachers, err := service.GetAllTeachers()

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, teachers)
		mockTeacherStore.AssertExpectations(t)
	})
}
