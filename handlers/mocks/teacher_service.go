package mocks

import (
	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockTeacherService is a mock implementation of TeacherService.
type MockTeacherService struct {
	mock.Mock
}

// CreateTeacher mocks the CreateTeacher method.
func (m *MockTeacherService) CreateTeacher(teacher *models.Teacher) (*models.Teacher, error) {
	args := m.Called(teacher)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Teacher), args.Error(1)
}

// GetTeacherByID mocks the GetTeacherByID method.
func (m *MockTeacherService) GetTeacherByID(id int) (*models.Teacher, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Teacher), args.Error(1)
}

// UpdateTeacher mocks the UpdateTeacher method.
func (m *MockTeacherService) UpdateTeacher(teacher *models.Teacher) error {
	args := m.Called(teacher)
	return args.Error(0)
}

// DeleteTeacher mocks the DeleteTeacher method.
func (m *MockTeacherService) DeleteTeacher(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetAllTeachers mocks the GetAllTeachers method.
func (m *MockTeacherService) GetAllTeachers() ([]models.Teacher, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Teacher), args.Error(1)
}
