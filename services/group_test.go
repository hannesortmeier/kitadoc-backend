package services_test

import (
	"errors"
	"testing"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sirupsen/logrus"
)

// MockGroupStore is a mock implementation of data.GroupStore
type MockGroupStore struct {
	mock.Mock
}

func (m *MockGroupStore) Create(group *models.Group) (int, error) {
	args := m.Called(group)
	return args.Int(0), args.Error(1)
}

func (m *MockGroupStore) GetByID(id int) (*models.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupStore) GetByName(name string) (*models.Group, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupStore) GetAll() ([]models.Group, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Group), args.Error(1)
}

func (m *MockGroupStore) Update(group *models.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGroupStore) AddChildToGroup(groupID, childID int) error {
	args := m.Called(groupID, childID)
	return args.Error(0)
}

func (m *MockGroupStore) RemoveChildFromGroup(groupID, childID int) error {
	args := m.Called(groupID, childID)
	return args.Error(0)
}

func (m *MockGroupStore) AddTeacherToGroup(groupID, teacherID int) error {
	args := m.Called(groupID, teacherID)
	return args.Error(0)
}

func (m *MockGroupStore) RemoveTeacherFromGroup(groupID, teacherID int) error {
	args := m.Called(groupID, teacherID)
	return args.Error(0)
}

func (m *MockGroupStore) GetChildrenInGroup(groupID int) ([]models.Child, error) {
	args := m.Called(groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Child), args.Error(1)
}

func (m *MockGroupStore) GetTeachersInGroup(groupID int) ([]models.Teacher, error) {
	args := m.Called(groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Teacher), args.Error(1)
}

func TestCreateGroup(t *testing.T) {
	mockGroupStore := new(MockGroupStore)
	service := services.NewGroupService(mockGroupStore)

	log_level, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful creation
	t.Run("success", func(t *testing.T) {
		group := &models.Group{Name: "New Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, data.ErrNotFound).Once()
		mockGroupStore.On("Create", mock.AnythingOfType("*models.Group")).Return(1, nil).Once()

		createdGroup, err := service.CreateGroup(group)

		assert.NoError(t, err)
		assert.NotNil(t, createdGroup)
		assert.Equal(t, 1, createdGroup.ID)
		assert.Equal(t, "New Group", createdGroup.Name)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		group := &models.Group{Name: ""} // Invalid name
		createdGroup, err := service.CreateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		assert.Nil(t, createdGroup)
		mockGroupStore.AssertNotCalled(t, "GetByName")
		mockGroupStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Group with same name already exists
	t.Run("already exists", func(t *testing.T) {
		group := &models.Group{Name: "Existing Group"}
		existingGroup := &models.Group{ID: 1, Name: "Existing Group"}
		mockGroupStore.On("GetByName", group.Name).Return(existingGroup, nil).Once()

		createdGroup, err := service.CreateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrAlreadyExists, err)
		assert.Nil(t, createdGroup)
		mockGroupStore.AssertExpectations(t)
		mockGroupStore.AssertNotCalled(t, "Create")
	})

	// Test case 4: Internal error during GetByName check
	t.Run("internal error on GetByName", func(t *testing.T) {
		group := &models.Group{Name: "New Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, errors.New("db error")).Once()

		createdGroup, err := service.CreateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdGroup)
		mockGroupStore.AssertExpectations(t)
		mockGroupStore.AssertNotCalled(t, "Create")
	})

	// Test case 5: Internal error during creation
	t.Run("internal error on create", func(t *testing.T) {
		group := &models.Group{Name: "New Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, data.ErrNotFound).Once()
		mockGroupStore.On("Create", mock.AnythingOfType("*models.Group")).Return(0, errors.New("db error")).Once()

		createdGroup, err := service.CreateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdGroup)
		mockGroupStore.AssertExpectations(t)
	})
}

func TestGetGroupByID(t *testing.T) {
	mockGroupStore := new(MockGroupStore)
	service := services.NewGroupService(mockGroupStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		groupID := 1
		expectedGroup := &models.Group{ID: groupID, Name: "Test Group"}
		mockGroupStore.On("GetByID", groupID).Return(expectedGroup, nil).Once()

		group, err := service.GetGroupByID(groupID)

		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, expectedGroup.ID, group.ID)
		assert.Equal(t, expectedGroup.Name, group.Name)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 2: Group not found
	t.Run("not found", func(t *testing.T) {
		groupID := 99
		mockGroupStore.On("GetByID", groupID).Return(nil, data.ErrNotFound).Once()

		group, err := service.GetGroupByID(groupID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, group)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		groupID := 1
		mockGroupStore.On("GetByID", groupID).Return(nil, errors.New("db error")).Once()

		group, err := service.GetGroupByID(groupID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, group)
		mockGroupStore.AssertExpectations(t)
	})
}

func TestUpdateGroup(t *testing.T) {
	mockGroupStore := new(MockGroupStore)
	service := services.NewGroupService(mockGroupStore)

	// Test case 1: Successful update
	t.Run("success", func(t *testing.T) {
		group := &models.Group{ID: 1, Name: "Updated Group"}
		// Simulate that the name is unique or the same as the existing one
		mockGroupStore.On("GetByName", group.Name).Return(nil, data.ErrNotFound).Once()
		mockGroupStore.On("Update", mock.AnythingOfType("*models.Group")).Return(nil).Once()

		err := service.UpdateGroup(group)

		assert.NoError(t, err)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		group := &models.Group{ID: 1, Name: ""} // Invalid name
		err := service.UpdateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockGroupStore.AssertNotCalled(t, "GetByName")
		mockGroupStore.AssertNotCalled(t, "Update")
	})

	// Test case 3: Group name conflict
	t.Run("name conflict", func(t *testing.T) {
		group := &models.Group{ID: 1, Name: "New Name"}
		conflictingGroup := &models.Group{ID: 2, Name: "New Name"}
		mockGroupStore.On("GetByName", group.Name).Return(conflictingGroup, nil).Once()

		err := service.UpdateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrAlreadyExists, err)
		mockGroupStore.AssertExpectations(t)
		mockGroupStore.AssertNotCalled(t, "Update")
	})

	// Test case 4: Internal error during GetByName check
	t.Run("internal error on GetByName", func(t *testing.T) {
		group := &models.Group{ID: 1, Name: "Updated Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, errors.New("db error")).Once()

		err := service.UpdateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockGroupStore.AssertExpectations(t)
		mockGroupStore.AssertNotCalled(t, "Update")
	})

	// Test case 5: Group not found during update
	t.Run("not found on update", func(t *testing.T) {
		group := &models.Group{ID: 99, Name: "Updated Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, data.ErrNotFound).Once() // Simulate no name conflict
		mockGroupStore.On("Update", mock.AnythingOfType("*models.Group")).Return(data.ErrNotFound).Once()

		err := service.UpdateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 6: Internal error during update
	t.Run("internal error on update", func(t *testing.T) {
		group := &models.Group{ID: 1, Name: "Updated Group"}
		mockGroupStore.On("GetByName", group.Name).Return(nil, data.ErrNotFound).Once() // Simulate no name conflict
		mockGroupStore.On("Update", mock.AnythingOfType("*models.Group")).Return(errors.New("db error")).Once()

		err := service.UpdateGroup(group)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockGroupStore.AssertExpectations(t)
	})
}

func TestDeleteGroup(t *testing.T) {
	mockGroupStore := new(MockGroupStore)
	service := services.NewGroupService(mockGroupStore)

	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		groupID := 1
		mockGroupStore.On("Delete", groupID).Return(nil).Once()

		err := service.DeleteGroup(groupID)

		assert.NoError(t, err)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 2: Group not found
	t.Run("not found", func(t *testing.T) {
		groupID := 99
		mockGroupStore.On("Delete", groupID).Return(data.ErrNotFound).Once()

		err := service.DeleteGroup(groupID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		groupID := 1
		mockGroupStore.On("Delete", groupID).Return(errors.New("db error")).Once()

		err := service.DeleteGroup(groupID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockGroupStore.AssertExpectations(t)
	})
}

func TestGetAllGroups(t *testing.T) {
	mockGroupStore := new(MockGroupStore)
	service := services.NewGroupService(mockGroupStore)

	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		expectedGroups := []models.Group{
			{ID: 1, Name: "Group A"},
			{ID: 2, Name: "Group B"},
		}
		mockGroupStore.On("GetAll").Return(expectedGroups, nil).Once()

		groups, err := service.GetAllGroups()

		assert.NoError(t, err)
		assert.NotNil(t, groups)
		assert.Equal(t, expectedGroups, groups)
		mockGroupStore.AssertExpectations(t)
	})

	// Test case 2: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockGroupStore.On("GetAll").Return(nil, errors.New("db error")).Once()

		groups, err := service.GetAllGroups()

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, groups)
		mockGroupStore.AssertExpectations(t)
	})
}