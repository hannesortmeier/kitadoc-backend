package services

import (
	"errors"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// TeacherService defines the interface for teacher-related business logic operations.
type TeacherService interface {
	CreateTeacher(teacher *models.Teacher) (*models.Teacher, error)
	GetTeacherByID(id int) (*models.Teacher, error)
	UpdateTeacher(teacher *models.Teacher) error
	DeleteTeacher(id int) error
	GetAllTeachers() ([]models.Teacher, error)
}

// TeacherServiceImpl implements TeacherService.
type TeacherServiceImpl struct {
	teacherStore data.TeacherStore
	validate     *validator.Validate
}

// NewTeacherService creates a new TeacherServiceImpl.
func NewTeacherService(teacherStore data.TeacherStore) *TeacherServiceImpl {
	return &TeacherServiceImpl{
		teacherStore: teacherStore,
		validate:     validator.New(),
	}
}

// CreateTeacher creates a new teacher.
func (s *TeacherServiceImpl) CreateTeacher(teacher *models.Teacher) (*models.Teacher, error) {
	if err := models.ValidateTeacher(*teacher); err != nil {
		return nil, ErrInvalidInput
	}

	teacher.CreatedAt = time.Now()
	teacher.UpdatedAt = time.Now()

	id, err := s.teacherStore.Create(teacher)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error creating teacher: %v", err)
		return nil, ErrInternal
	}
	teacher.ID = id
	return teacher, nil
}

// GetTeacherByID fetches a teacher by ID.
func (s *TeacherServiceImpl) GetTeacherByID(id int) (*models.Teacher, error) {
	teacher, err := s.teacherStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return teacher, nil
}

// UpdateTeacher updates an existing teacher.
func (s *TeacherServiceImpl) UpdateTeacher(teacher *models.Teacher) error {
	if err := models.ValidateTeacher(*teacher); err != nil {
		return ErrInvalidInput
	}

	teacher.UpdatedAt = time.Now()
	err := s.teacherStore.Update(teacher)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteTeacher deletes a teacher by ID.
func (s *TeacherServiceImpl) DeleteTeacher(id int) error {
	err := s.teacherStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// GetAllTeachers fetches all teachers.
func (s *TeacherServiceImpl) GetAllTeachers() ([]models.Teacher, error) {
	// The data layer's GetAll method for teachers doesn't have pagination/filtering,
	// so we can directly call it.
	teachers, err := s.teacherStore.GetAll()
	if err != nil {
		return nil, ErrInternal
	}
	return teachers, nil
}