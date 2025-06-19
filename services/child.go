package services

import (
	"errors"
	"time"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// ChildService defines the interface for child-related business logic operations.
type ChildService interface {
	CreateChild(child *models.Child) (*models.Child, error)
	GetChildByID(id int) (*models.Child, error)
	UpdateChild(child *models.Child) error
	DeleteChild(id int) error
	GetAllChildren() ([]models.Child, error)
	BulkImportChildren(fileContent []byte) error // Placeholder for file processing
}

// ChildServiceImpl implements ChildService.
type ChildServiceImpl struct {
	childStore data.ChildStore
	groupStore data.GroupStore // To check if group exists
	validate   *validator.Validate
}

// NewChildService creates a new ChildServiceImpl.
func NewChildService(childStore data.ChildStore, groupStore data.GroupStore) *ChildServiceImpl {
	validate := validator.New()
	validate.RegisterValidation("childbirthdate", models.ValidateChildBirthdate) // Register custom validation
	return &ChildServiceImpl{
		childStore: childStore,
		groupStore: groupStore,
		validate:   validate,
	}
}

// CreateChild creates a new child.
func (s *ChildServiceImpl) CreateChild(child *models.Child) (*models.Child, error) {
	if err := s.validate.Struct(child); err != nil {
		logger.GetGlobalLogger().Errorf("Validation error: %v", err)
		return nil, ErrInvalidInput
	}

	if child.GroupID != nil {
		_, err := s.groupStore.GetByID(*child.GroupID)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return nil, ErrNotFound // Group not found
			}
			logger.GetGlobalLogger().Errorf("Error checking group existence: %v", err)
			return nil, ErrInternal
		}
	}

	child.CreatedAt = time.Now()
	child.UpdatedAt = time.Now()

	id, err := s.childStore.Create(child)
	if err != nil {
		return nil, ErrInternal
	}
	child.ID = id
	return child, nil
}

// GetChildByID fetches a child by ID.
func (s *ChildServiceImpl) GetChildByID(id int) (*models.Child, error) {
	child, err := s.childStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return child, nil
}

// UpdateChild updates an existing child.
func (s *ChildServiceImpl) UpdateChild(child *models.Child) error {
	if err := s.validate.Struct(child); err != nil {
		return ErrInvalidInput
	}

	if child.GroupID != nil {
		_, err := s.groupStore.GetByID(*child.GroupID)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return ErrNotFound // Group not found
			}
			return ErrInternal
		}
	}

	child.UpdatedAt = time.Now()
	err := s.childStore.Update(child)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteChild deletes a child by ID.
func (s *ChildServiceImpl) DeleteChild(id int) error {
	err := s.childStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// GetAllChildren fetches all children.
func (s *ChildServiceImpl) GetAllChildren() ([]models.Child, error) {
	children, err := s.childStore.GetAll(nil) // No pagination or filtering
	if err != nil {
		return nil, ErrInternal
	}
	return children, nil
}

// BulkImportChildren handles the bulk import of children from a file.
// This is a placeholder for actual file processing logic.
func (s *ChildServiceImpl) BulkImportChildren(fileContent []byte) error {
	// In a real implementation, you would parse fileContent (e.g., CSV, Excel)
	// and then iterate through records to create children using s.childStore.Create.
	// You would also handle validation for each imported child.
	// For now, we just return a placeholder error.
	_ = fileContent // Suppress unused variable warning
	return ErrBulkImportFailed
}