package mocks

import (
	"context"
	"kitadoc-backend/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// MockDocumentationEntryService is a mock implementation of DocumentationEntryService.
type MockDocumentationEntryService struct {
	mock.Mock
}

// CreateDocumentationEntry mocks the CreateDocumentationEntry method.
func (m *MockDocumentationEntryService) CreateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) (*models.DocumentationEntry, error) {
	args := m.Called(logger, ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentationEntry), args.Error(1)
}

// GetDocumentationEntryByID mocks the GetDocumentationEntryByID method.
func (m *MockDocumentationEntryService) GetDocumentationEntryByID(logger *logrus.Entry, ctx context.Context, id int) (*models.DocumentationEntry, error) {
	args := m.Called(logger, ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentationEntry), args.Error(1)
}

// UpdateDocumentationEntry mocks the UpdateDocumentationEntry method.
func (m *MockDocumentationEntryService) UpdateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) error {
	args := m.Called(logger, ctx, entry)
	return args.Error(0)
}

// DeleteDocumentationEntry mocks the DeleteDocumentationEntry method.
func (m *MockDocumentationEntryService) DeleteDocumentationEntry(logger *logrus.Entry, ctx context.Context, id int) error {
	args := m.Called(logger, ctx, id)
	return args.Error(0)
}

// GetAllDocumentationForChild mocks the GetAllDocumentationForChild method.
func (m *MockDocumentationEntryService) GetAllDocumentationForChild(logger *logrus.Entry, ctx context.Context, childID int) ([]models.DocumentationEntry, error) {
	args := m.Called(logger, ctx, childID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DocumentationEntry), args.Error(1)
}

// ApproveDocumentationEntry mocks the ApproveDocumentationEntry method.
func (m *MockDocumentationEntryService) ApproveDocumentationEntry(logger *logrus.Entry, ctx context.Context, entryID int, approvedByUserID int) error {
	args := m.Called(logger, ctx, entryID, approvedByUserID)
	return args.Error(0)
}

// GenerateChildReport mocks the GenerateChildReport method.
func (m *MockDocumentationEntryService) GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int) ([]byte, error) {
	args := m.Called(logger, ctx, childID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
