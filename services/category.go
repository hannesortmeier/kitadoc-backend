package services

import (
	"errors"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// CategoryService defines the interface for category-related business logic operations.
type CategoryService interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id int) error
	GetAllCategories() ([]models.Category, error)
}

// CategoryServiceImpl implements CategoryService.
type CategoryServiceImpl struct {
	categoryStore data.CategoryStore
	validate      *validator.Validate
}

// NewCategoryService creates a new CategoryServiceImpl.
func NewCategoryService(categoryStore data.CategoryStore) *CategoryServiceImpl {
	return &CategoryServiceImpl{
		categoryStore: categoryStore,
		validate:      validator.New(),
	}
}

// CreateCategory creates a new category.
func (s *CategoryServiceImpl) CreateCategory(category *models.Category) (*models.Category, error) {
	if err := models.ValidateCategory(*category); err != nil {
		return nil, ErrInvalidInput
	}

	// Check for unique category name
	existingCategory, err := s.categoryStore.GetByName(category.Name)
	if err == nil && existingCategory != nil {
		return nil, ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		logger.GetGlobalLogger().Errorf("Error checking category name uniqueness: %v", err)
		return nil, ErrInternal
	}

	id, err := s.categoryStore.Create(category)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error creating category: %v", err)
		return nil, ErrInternal
	}
	category.ID = id
	return category, nil
}

// GetCategoryByID fetches a category by ID.
func (s *CategoryServiceImpl) GetCategoryByID(id int) (*models.Category, error) {
	category, err := s.categoryStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return category, nil
}

// UpdateCategory updates an existing category.
func (s *CategoryServiceImpl) UpdateCategory(category *models.Category) error {
	if err := models.ValidateCategory(*category); err != nil {
		return ErrInvalidInput
	}

	// Check for unique category name if name is changed
	existingCategory, err := s.categoryStore.GetByName(category.Name)
	if err == nil && existingCategory != nil && existingCategory.ID != category.ID {
		return ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return ErrInternal
	}

	err = s.categoryStore.Update(category)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteCategory deletes a category by ID.
func (s *CategoryServiceImpl) DeleteCategory(id int) error {
	err := s.categoryStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// GetAllCategories fetches all categories.
func (s *CategoryServiceImpl) GetAllCategories() ([]models.Category, error) {
	categories, err := s.categoryStore.GetAll()
	if err != nil {
		return nil, ErrInternal
	}
	return categories, nil
}