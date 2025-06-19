package services_test

import (
	"errors"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/services/mocks"
	"kitadoc-backend/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sirupsen/logrus"
)

func TestCreateAssignment(t *testing.T) {
	log_level, _ := logrus.ParseLevel("debug")

	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour), // Past date
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", assignment.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", assignment.TeacherID).Return(expectedTeacher, nil).Once()
		mockAssignmentStore.On("Create", mock.AnythingOfType("*models.Assignment")).Return(1, nil).Once()

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.NoError(t, err)
		assert.NotNil(t, createdAssignment)
		assert.Equal(t, 1, createdAssignment.ID)
		mockAssignmentStore.AssertExpectations(t)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID: 0, // Invalid ChildID
		}

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertNotCalled(t, "Create")
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
	})

	// Test case 3: Child not found
	t.Run("child not found", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID:   99, // Non-existent child
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}

		mockChildStore.On("GetByID", assignment.ChildID).Return(nil, data.ErrNotFound).Once()

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child not found")
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertNotCalled(t, "Create")
		mockChildStore.AssertExpectations(t)
	})

	// Test case 4: Teacher not found
	t.Run("teacher not found", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID:   1,
			TeacherID: 99, // Non-existent teacher
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		expectedChild := &models.Child{ID: 1}

		mockChildStore.On("GetByID", assignment.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", assignment.TeacherID).Return(nil, data.ErrNotFound).Once()

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "teacher not found")
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertNotCalled(t, "Create")
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 5: Assignment start date in the future
	t.Run("future start date", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(24 * time.Hour), // Future date
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", assignment.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", assignment.TeacherID).Return(expectedTeacher, nil).Once()

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assignment start date cannot be in the future")
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertNotCalled(t, "Create")
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})

	// Test case 6: Assignment end date before start date
	t.Run("end date before start date", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(-48 * time.Hour) // Before start date
		assignment := &models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: startDate,
			EndDate:   &endDate,
		}
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", assignment.ChildID).Return(expectedChild, nil)
		mockTeacherStore.On("GetByID", assignment.TeacherID).Return(expectedTeacher, nil)

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input")
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertNotCalled(t, "Create")
		mockChildStore.AssertNotCalled(t, "GetByID")
		mockTeacherStore.AssertNotCalled(t, "GetByID")
	})

	// Test case 7: Internal error during assignment creation
	t.Run("internal error on create", func(t *testing.T) {
		// Create fresh mocks for this test case
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		
		expectedChild := &models.Child{ID: 1}
		expectedTeacher := &models.Teacher{ID: 1}

		mockChildStore.On("GetByID", assignment.ChildID).Return(expectedChild, nil).Once()
		mockTeacherStore.On("GetByID", assignment.TeacherID).Return(expectedTeacher, nil).Once()
		mockAssignmentStore.On("Create", mock.AnythingOfType("*models.Assignment")).Return(0, errors.New("db error")).Once()

		createdAssignment, err := service.CreateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdAssignment)
		mockAssignmentStore.AssertExpectations(t)
		mockChildStore.AssertExpectations(t)
		mockTeacherStore.AssertExpectations(t)
	})
}

func TestGetAssignmentByID(t *testing.T) {
	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		expectedAssignment := &models.Assignment{ID: assignmentID}
		mockAssignmentStore.On("GetByID", assignmentID).Return(expectedAssignment, nil).Once()

		assignment, err := service.GetAssignmentByID(assignmentID)

		assert.NoError(t, err)
		assert.NotNil(t, assignment)
		assert.Equal(t, expectedAssignment.ID, assignment.ID)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 2: Assignment not found
	t.Run("not found", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 99
		mockAssignmentStore.On("GetByID", assignmentID).Return(nil, data.ErrNotFound).Once()

		assignment, err := service.GetAssignmentByID(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, assignment)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		mockAssignmentStore.On("GetByID", assignmentID).Return(nil, errors.New("db error")).Once()

		assignment, err := service.GetAssignmentByID(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, assignment)
		mockAssignmentStore.AssertExpectations(t)
	})
}

func TestUpdateAssignment(t *testing.T) {
	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ID:        1,
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		existingAssignment := &models.Assignment{ID: 1}

		mockAssignmentStore.On("GetByID", assignment.ID).Return(existingAssignment, nil).Once()
		mockAssignmentStore.On("Update", mock.AnythingOfType("*models.Assignment")).Return(nil).Once()

		err := service.UpdateAssignment(assignment)

		assert.NoError(t, err)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ID:      1,
			ChildID: 0, // Invalid ChildID
		}

		err := service.UpdateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockAssignmentStore.AssertNotCalled(t, "GetByID")
		mockAssignmentStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Assignment not found during fetch
	t.Run("not found on fetch", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ID:        99, // Non-existent ID
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		mockAssignmentStore.On("GetByID", assignment.ID).Return(nil, data.ErrNotFound).Once()

		err := service.UpdateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockAssignmentStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "Update")
	})

	// Test case 4: Internal error during fetch
	t.Run("internal error on fetch", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ID:        1,
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		mockAssignmentStore.On("GetByID", assignment.ID).Return(nil, errors.New("db error")).Once()

		err := service.UpdateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAssignmentStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "Update")
	})

	// Test case 5: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignment := &models.Assignment{
			ID:        1,
			ChildID:   1,
			TeacherID: 1,
			StartDate: time.Now().Add(-24 * time.Hour),
		}
		existingAssignment := &models.Assignment{ID: 1}

		mockAssignmentStore.On("GetByID", assignment.ID).Return(existingAssignment, nil).Once()
		mockAssignmentStore.On("Update", mock.AnythingOfType("*models.Assignment")).Return(errors.New("db error")).Once()

		err := service.UpdateAssignment(assignment)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAssignmentStore.AssertExpectations(t)
	})
}

func TestDeleteAssignment(t *testing.T) {
	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		mockAssignmentStore.On("Delete", assignmentID).Return(nil).Once()

		err := service.DeleteAssignment(assignmentID)

		assert.NoError(t, err)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 2: Assignment not found
	t.Run("not found", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 99
		mockAssignmentStore.On("Delete", assignmentID).Return(data.ErrNotFound).Once()

		err := service.DeleteAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		mockAssignmentStore.On("Delete", assignmentID).Return(errors.New("db error")).Once()

		err := service.DeleteAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAssignmentStore.AssertExpectations(t)
	})
}

func TestEndAssignment(t *testing.T) {
	// Test case 1: Successful end assignment
	t.Run("success", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		assignment := &models.Assignment{
			ID:        assignmentID,
			StartDate: time.Now().Add(-48 * time.Hour),
			EndDate:   nil, // Not yet ended
		}
		mockAssignmentStore.On("GetByID", assignmentID).Return(assignment, nil).Once()
		mockAssignmentStore.On("EndAssignment", assignmentID).Return(nil).Once()

		err := service.EndAssignment(assignmentID)

		assert.NoError(t, err)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 2: Assignment not found
	t.Run("not found", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 99
		mockAssignmentStore.On("GetByID", assignmentID).Return(nil, data.ErrNotFound).Once()

		err := service.EndAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockAssignmentStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "EndAssignment")
	})

	// Test case 3: Assignment already ended
	t.Run("already ended", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		now := time.Now()
		assignment := &models.Assignment{
			ID:        assignmentID,
			StartDate: time.Now().Add(-48 * time.Hour),
			EndDate:   &now, // Already ended
		}
		mockAssignmentStore.On("GetByID", assignmentID).Return(assignment, nil).Once()

		err := service.EndAssignment(assignmentID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assignment has already ended")
		mockAssignmentStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "EndAssignment")
	})

	// Test case 4: Internal error during fetch
	t.Run("internal error on fetch", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		mockAssignmentStore.On("GetByID", assignmentID).Return(nil, errors.New("db error")).Once()

		err := service.EndAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAssignmentStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "EndAssignment")
	})

	// Test case 5: Internal error during end assignment
	t.Run("internal error on end assignment", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		assignmentID := 1
		assignment := &models.Assignment{
			ID:        assignmentID,
			StartDate: time.Now().Add(-48 * time.Hour),
			EndDate:   nil,
		}
		mockAssignmentStore.On("GetByID", assignmentID).Return(assignment, nil).Once()
		mockAssignmentStore.On("EndAssignment", assignmentID).Return(errors.New("db error")).Once()

		err := service.EndAssignment(assignmentID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAssignmentStore.AssertExpectations(t)
	})
}

func TestGetAssignmentHistoryForChild(t *testing.T) {
	log_level, _ := logrus.ParseLevel("debug")

	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		childID := 1
		expectedChild := &models.Child{ID: childID}
		expectedAssignments := []models.Assignment{
			{ID: 1, ChildID: childID},
			{ID: 2, ChildID: childID},
		}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockAssignmentStore.On("GetAssignmentHistoryForChild", childID).Return(expectedAssignments, nil).Once()

		assignments, err := service.GetAssignmentHistoryForChild(childID)

		assert.NoError(t, err)
		assert.NotNil(t, assignments)
		assert.Equal(t, expectedAssignments, assignments)
		mockChildStore.AssertExpectations(t)
		mockAssignmentStore.AssertExpectations(t)
	})

	// Test case 2: Child not found
	t.Run("child not found", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		childID := 99
		mockChildStore.On("GetByID", childID).Return(nil, data.ErrNotFound).Once()

		assignments, err := service.GetAssignmentHistoryForChild(childID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child not found")
		assert.Nil(t, assignments)
		mockChildStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "GetAssignmentHistoryForChild")
	})

	// Test case 3: Internal error during child fetch
	t.Run("internal error on child fetch", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		childID := 42
		mockChildStore.On("GetByID", childID).Return(nil, errors.New("db error")).Once()

		assignments, err := service.GetAssignmentHistoryForChild(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, assignments)
		mockChildStore.AssertExpectations(t)
		mockAssignmentStore.AssertNotCalled(t, "GetAssignmentHistoryForChild")
	})

	// Test case 4: Internal error during assignment fetch
	t.Run("internal error on assignment fetch", func(t *testing.T) {
		mockAssignmentStore := new(mocks.MockAssignmentStore)
		mockChildStore := new(mocks.MockChildStore)
		mockTeacherStore := new(mocks.MockTeacherStore)
		service := services.NewAssignmentService(mockAssignmentStore, mockChildStore, mockTeacherStore)

		childID := 1
		expectedChild := &models.Child{ID: childID}
		mockChildStore.On("GetByID", childID).Return(expectedChild, nil).Once()
		mockAssignmentStore.On("GetAssignmentHistoryForChild", childID).Return(nil, errors.New("db error")).Once()

		assignments, err := service.GetAssignmentHistoryForChild(childID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, assignments)
		mockChildStore.AssertExpectations(t)
		mockAssignmentStore.AssertExpectations(t)
	})
}