package mocks

import (
	"github.com/stretchr/testify/mock"
	"kitadoc-backend/models"
)

// MockAssignmentStore is a mock implementation of data.AssignmentStore
type MockAssignmentStore struct {
	mock.Mock
}

func (m *MockAssignmentStore) Create(assignment *models.Assignment) (int, error) {
	args := m.Called(assignment)
	return args.Int(0), args.Error(1)
}

func (m *MockAssignmentStore) GetByID(id int) (*models.Assignment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assignment), args.Error(1)
}

func (m *MockAssignmentStore) Update(assignment *models.Assignment) error {
	args := m.Called(assignment)
	return args.Error(0)
}

func (m *MockAssignmentStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAssignmentStore) GetAssignmentHistoryForChild(childID int) ([]models.Assignment, error) {
	args := m.Called(childID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Assignment), args.Error(1)
}

func (m *MockAssignmentStore) EndAssignment(assignmentID int) error {
	args := m.Called(assignmentID)
	return args.Error(0)
}

// MockChildStore is a mock implementation of data.ChildStore
type MockChildStore struct {
	mock.Mock
}

func (m *MockChildStore) Create(child *models.Child) (int, error) {
	args := m.Called(child)
	return args.Int(0), args.Error(1)
}

func (m *MockChildStore) GetByID(id int) (*models.Child, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Child), args.Error(1)
}

func (m *MockChildStore) Update(child *models.Child) error {
	args := m.Called(child)
	return args.Error(0)
}

func (m *MockChildStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockChildStore) GetAll(groupID *int) ([]models.Child, error) {
	args := m.Called(groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Child), args.Error(1)
}

// MockTeacherStore is a mock implementation of data.TeacherStore
type MockTeacherStore struct {
	mock.Mock
}

func (m *MockTeacherStore) Create(teacher *models.Teacher) (int, error) {
	args := m.Called(teacher)
	return args.Int(0), args.Error(1)
}

func (m *MockTeacherStore) GetByID(id int) (*models.Teacher, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Teacher), args.Error(1)
}

func (m *MockTeacherStore) Update(teacher *models.Teacher) error {
	args := m.Called(teacher)
	return args.Error(0)
}

func (m *MockTeacherStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTeacherStore) GetAll() ([]models.Teacher, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Teacher), args.Error(1)
}

// MockDocumentationEntryStore is a mock implementation of data.DocumentationEntryStore
type MockDocumentationEntryStore struct {
	mock.Mock
}

func (m *MockDocumentationEntryStore) Create(entry *models.DocumentationEntry) (int, error) {
	args := m.Called(entry)
	return args.Int(0), args.Error(1)
}

func (m *MockDocumentationEntryStore) GetByID(id int) (*models.DocumentationEntry, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentationEntry), args.Error(1)
}

func (m *MockDocumentationEntryStore) Update(entry *models.DocumentationEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockDocumentationEntryStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDocumentationEntryStore) GetAll() ([]models.DocumentationEntry, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DocumentationEntry), args.Error(1)
}

func (m *MockDocumentationEntryStore) GetAllForChild(childID int) ([]models.DocumentationEntry, error) {
	args := m.Called(childID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DocumentationEntry), args.Error(1)
}

func (m *MockDocumentationEntryStore) ApproveEntry(entryID, approvedByUserID int) error {
	args := m.Called(entryID, approvedByUserID)
	return args.Error(0)
}

// MockCategoryStore is a mock implementation of data.CategoryStore
type MockCategoryStore struct {
	mock.Mock
}

func (m *MockCategoryStore) Create(category *models.Category) (int, error) {
	args := m.Called(category)
	return args.Int(0), args.Error(1)
}

func (m *MockCategoryStore) GetByID(id int) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryStore) GetByName(name string) (*models.Category, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryStore) GetAll() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryStore) Update(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserStore is a mock implementation of data.UserStore
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(user *models.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserStore) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStore) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStore) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserStore) GetAll() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

// MockAudioRecordingStore is a mock implementation of data.AudioRecordingStore
type MockAudioRecordingStore struct {
	mock.Mock
}

func (m *MockAudioRecordingStore) Create(recording *models.AudioRecording) (int, error) {
	args := m.Called(recording)
	return args.Int(0), args.Error(1)
}

func (m *MockAudioRecordingStore) GetByID(id int) (*models.AudioRecording, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AudioRecording), args.Error(1)
}

func (m *MockAudioRecordingStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
