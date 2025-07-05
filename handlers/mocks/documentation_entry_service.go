package mocks

import (
	"context"

	"kitadoc-backend/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// MockDocumentationEntryService is a mock type for the DocumentationEntryService type
type MockDocumentationEntryService struct {
	mock.Mock
}

// CreateDocumentationEntry provides a mock function with given fields: logger, ctx, entry
func (_m *MockDocumentationEntryService) CreateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) (*models.DocumentationEntry, error) {
	ret := _m.Called(logger, ctx, entry)

	var r0 *models.DocumentationEntry
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, *models.DocumentationEntry) *models.DocumentationEntry); ok {
		r0 = rf(logger, ctx, entry)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.DocumentationEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*logrus.Entry, context.Context, *models.DocumentationEntry) error); ok {
		r1 = rf(logger, ctx, entry)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDocumentationEntryByID provides a mock function with given fields: logger, ctx, id
func (_m *MockDocumentationEntryService) GetDocumentationEntryByID(logger *logrus.Entry, ctx context.Context, id int) (*models.DocumentationEntry, error) {
	ret := _m.Called(logger, ctx, id)

	var r0 *models.DocumentationEntry
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, int) *models.DocumentationEntry); ok {
		r0 = rf(logger, ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.DocumentationEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*logrus.Entry, context.Context, int) error); ok {
		r1 = rf(logger, ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateDocumentationEntry provides a mock function with given fields: logger, ctx, entry
func (_m *MockDocumentationEntryService) UpdateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) error {
	ret := _m.Called(logger, ctx, entry)

	var r0 error
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, *models.DocumentationEntry) error); ok {
		r0 = rf(logger, ctx, entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDocumentationEntry provides a mock function with given fields: logger, ctx, id
func (_m *MockDocumentationEntryService) DeleteDocumentationEntry(logger *logrus.Entry, ctx context.Context, id int) error {
	ret := _m.Called(logger, ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, int) error); ok {
		r0 = rf(logger, ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllDocumentationForChild provides a mock function with given fields: logger, ctx, childID
func (_m *MockDocumentationEntryService) GetAllDocumentationForChild(logger *logrus.Entry, ctx context.Context, childID int) ([]models.DocumentationEntry, error) {
	ret := _m.Called(logger, ctx, childID)

	var r0 []models.DocumentationEntry
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, int) []models.DocumentationEntry); ok {
		r0 = rf(logger, ctx, childID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.DocumentationEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*logrus.Entry, context.Context, int) error); ok {
		r1 = rf(logger, ctx, childID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ApproveDocumentationEntry provides a mock function with given fields: logger, ctx, entryID, approvedByUserID
func (_m *MockDocumentationEntryService) ApproveDocumentationEntry(logger *logrus.Entry, ctx context.Context, entryID int, approvedByUserID int) error {
	ret := _m.Called(logger, ctx, entryID, approvedByUserID)

	var r0 error
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, int, int) error); ok {
		r0 = rf(logger, ctx, entryID, approvedByUserID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GenerateChildReport provides a mock function with given fields: logger, ctx, childID
func (_m *MockDocumentationEntryService) GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int) ([]byte, error) {
	ret := _m.Called(logger, ctx, childID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(*logrus.Entry, context.Context, int) []byte); ok {
		r0 = rf(logger, ctx, childID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*logrus.Entry, context.Context, int) error); ok {
		r1 = rf(logger, ctx, childID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
