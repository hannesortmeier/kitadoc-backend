package mocks

import (
	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockGroupService is a mock implementation of GroupService.
type MockGroupService struct {
	mock.Mock
}

// CreateGroup mocks the CreateGroup method.
func (m *MockGroupService) CreateGroup(group *models.Group) (*models.Group, error) {
	args := m.Called(group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

// GetGroupByID mocks the GetGroupByID method.
func (m *MockGroupService) GetGroupByID(id int) (*models.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

// UpdateGroup mocks the UpdateGroup method.
func (m *MockGroupService) UpdateGroup(group *models.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

// DeleteGroup mocks the DeleteGroup method.
func (m *MockGroupService) DeleteGroup(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetAllGroups mocks the GetAllGroups method.
func (m *MockGroupService) GetAllGroups() ([]models.Group, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Group), args.Error(1)
}