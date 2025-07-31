package services

import (
	"errors"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// AssignmentService defines the interface for assignment-related business logic operations.
type AssignmentService interface {
	CreateAssignment(assignment *models.Assignment) (*models.Assignment, error)
	GetAssignmentByID(id int) (*models.Assignment, error)
	UpdateAssignment(assignment *models.Assignment) error
	DeleteAssignment(id int) error
	GetAssignmentHistoryForChild(childID int) ([]models.Assignment, error)
}

// UpdateAssignment updates an existing assignment.
func (s *AssignmentServiceImpl) UpdateAssignment(assignment *models.Assignment) error {
	if err := models.ValidateAssignment(*assignment); err != nil {
		return ErrInvalidInput
	}

	// Fetch existing assignment to ensure it exists
	_, err := s.assignmentStore.GetByID(assignment.ID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Error fetching assignment by ID %d: %v", assignment.ID, err)
		return ErrInternal
	}

	assignment.UpdatedAt = time.Now()
	err = s.assignmentStore.Update(assignment)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Error updating assignment: %v", err)
		return ErrInternal
	}
	return nil
}

// AssignmentServiceImpl implements AssignmentService.
type AssignmentServiceImpl struct {
	assignmentStore data.AssignmentStore
	childStore      data.ChildStore
	teacherStore    data.TeacherStore
	validate        *validator.Validate
}

// NewAssignmentService creates a new AssignmentServiceImpl.
func NewAssignmentService(assignmentStore data.AssignmentStore, childStore data.ChildStore, teacherStore data.TeacherStore) *AssignmentServiceImpl {
	return &AssignmentServiceImpl{
		assignmentStore: assignmentStore,
		childStore:      childStore,
		teacherStore:    teacherStore,
		validate:        validator.New(),
	}
}

// CreateAssignment creates a new assignment.
func (s *AssignmentServiceImpl) CreateAssignment(assignment *models.Assignment) (*models.Assignment, error) {
	if err := models.ValidateAssignment(*assignment); err != nil {
		logger.GetGlobalLogger().Errorf("Error validating assignment: %v", err)
		return nil, ErrInvalidInput
	}

	// Validate ChildID
	_, err := s.childStore.GetByID(assignment.ChildID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("child not found")
		}
		logger.GetGlobalLogger().Errorf("Error fetching child by ID %d: %v", assignment.ChildID, err)
		return nil, ErrInternal
	}

	// Validate TeacherID
	_, err = s.teacherStore.GetByID(assignment.TeacherID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("teacher not found")
		}
		logger.GetGlobalLogger().Errorf("Error fetching teacher by ID %d: %v", assignment.TeacherID, err)
		return nil, ErrInternal
	}

	// Business rule: An assignment cannot start in the future.
	if assignment.StartDate.After(time.Now()) {
		return nil, errors.New("assignment start date cannot be in the future")
	}

	// Business rule: If EndDate is provided, it must be after StartDate.
	if assignment.EndDate != nil && assignment.EndDate.Before(assignment.StartDate) {
		return nil, errors.New("assignment end date cannot be before start date")
	}

	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	id, err := s.assignmentStore.Create(assignment)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error creating assignment: %v", err)
		return nil, ErrInternal
	}
	assignment.ID = id
	return assignment, nil
}

// GetAssignmentByID fetches an assignment by ID.
func (s *AssignmentServiceImpl) GetAssignmentByID(id int) (*models.Assignment, error) {
	assignment, err := s.assignmentStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return assignment, nil
}

// EndAssignment updates an existing assignment.
func (s *AssignmentServiceImpl) EndAssignment(assignmentID int) error {
	// Fetch the assignment to ensure it exists
	assignment, err := s.assignmentStore.GetByID(assignmentID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	// Business rule: An assignment cannot be ended if it has already ended.
	if assignment.EndDate != nil {
		return errors.New("assignment has already ended")
	}

	// Set the EndDate to now
	now := time.Now()
	assignment.EndDate = &now
	assignment.UpdatedAt = now

	err = s.assignmentStore.EndAssignment(assignmentID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteAssignment deletes an assignment by ID.
func (s *AssignmentServiceImpl) DeleteAssignment(id int) error {
	err := s.assignmentStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// GetAssignmentHistoryForChild fetches all assignments for a specific child.
func (s *AssignmentServiceImpl) GetAssignmentHistoryForChild(childID int) ([]models.Assignment, error) {
	// Validate ChildID
	_, err := s.childStore.GetByID(childID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.GetGlobalLogger().Errorf("Child with ID %d not found", childID)
			return nil, errors.New("child not found")
		}
		logger.GetGlobalLogger().Errorf("Error fetching child by ID %d: %v", childID, err)
		return nil, ErrInternal
	}

	assignments, err := s.assignmentStore.GetAssignmentHistoryForChild(childID)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error fetching assignment history for child ID %d: %v", childID, err)
		return nil, ErrInternal
	}
	return assignments, nil
}
