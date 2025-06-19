package mocks

import (
	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockChildService is a mock implementation of services.ChildService
type MockChildService struct {
	mock.Mock
}

func (m *MockChildService) CreateChild(child *models.Child) (*models.Child, error) {
	args := m.Called(child)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Child), args.Error(1)
}

func (m *MockChildService) GetChildByID(id int) (*models.Child, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Child), args.Error(1)
}

func (m *MockChildService) UpdateChild(child *models.Child) error {
	args := m.Called(child)
	return args.Error(0)
}

func (m *MockChildService) DeleteChild(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockChildService) GetAllChildren() ([]models.Child, error) {
	args := m.Called()
	return args.Get(0).([]models.Child), args.Error(1)
}

func (m *MockChildService) BulkImportChildren(fileContent []byte) error {
	args := m.Called(fileContent)
	return args.Error(0)
}