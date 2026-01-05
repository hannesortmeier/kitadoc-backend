package mocks

import (
	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockUserStore is a mock type for the UserStore type
type MockUserStore struct {
	mock.Mock
}

// Create provides a mock function with given fields: user
func (_m *MockUserStore) Create(user *models.User) (int, error) {
	ret := _m.Called(user)

	var r0 int
	if rf, ok := ret.Get(0).(func(*models.User) int); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.User) error); ok {
		r1 = rf(user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: id
func (_m *MockUserStore) GetByID(id int) (*models.User, error) {
	ret := _m.Called(id)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(int) *models.User); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: user
func (_m *MockUserStore) Update(user *models.User) error {
	ret := _m.Called(user)

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.User) error); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: id
func (_m *MockUserStore) Delete(id int) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetUserByUsername provides a mock function with given fields: username
func (_m *MockUserStore) GetUserByUsername(username string) (*models.User, error) {
	ret := _m.Called(username)

	var r0 *models.User
	if rf, ok := ret.Get(0).(func(string) *models.User); ok {
		r0 = rf(username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields:
func (_m *MockUserStore) GetAll() ([]*models.User, error) {
	ret := _m.Called()

	var r0 []*models.User
	if rf, ok := ret.Get(0).(func() []*models.User); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePassword provides a mock function with given fields: id, passwordHash
func (_m *MockUserStore) UpdatePassword(id int, passwordHash string) error {
	ret := _m.Called(id, passwordHash)

	var r0 error
	if rf, ok := ret.Get(0).(func(int, string) error); ok {
		r0 = rf(id, passwordHash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

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

func (m *MockAssignmentStore) GetAllAssignments() ([]models.Assignment, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Assignment), args.Error(1)
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

func (m *MockChildStore) GetAll() ([]models.Child, error) {
	args := m.Called()
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

// MockKitaMasterdataStore is a mock implementation of data.KitaMasterdataStore
type MockKitaMasterdataStore struct {
	mock.Mock
}

func (m *MockKitaMasterdataStore) Get() (*models.KitaMasterdata, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KitaMasterdata), args.Error(1)
}

func (m *MockKitaMasterdataStore) Update(data *models.KitaMasterdata) error {
	args := m.Called(data)
	return args.Error(0)
}

// MockProcessStore is a mock implementation of data.ProcessStore
type MockProcessStore struct {
	mock.Mock
}

func (m *MockProcessStore) Create(process *models.Process) (*models.Process, error) {
	args := m.Called(process)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Process), args.Error(1)
}

func (m *MockProcessStore) GetByID(id int) (*models.Process, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Process), args.Error(1)
}

func (m *MockProcessStore) Update(process *models.Process) error {
	args := m.Called(process)
	return args.Error(0)
}

func (m *MockProcessStore) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProcessStore) GetAll() ([]models.Process, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Process), args.Error(1)
}
