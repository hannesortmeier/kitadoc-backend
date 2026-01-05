package mocks

import (
	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockProcessService is a mock implementation of services.ProcessService
type MockProcessService struct {
	mock.Mock
}

func (m *MockProcessService) Create(status string) (*models.Process, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Process), args.Error(1)
}

func (m *MockProcessService) Update(process *models.Process) error {
	args := m.Called(process)
	return args.Error(0)
}

func (m *MockProcessService) GetByID(id int) (*models.Process, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Process), args.Error(1)
}
