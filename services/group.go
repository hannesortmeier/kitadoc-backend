package services

import (
	"errors"
	"log"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// GroupService defines the interface for group-related business logic operations.
type GroupService interface {
	CreateGroup(group *models.Group) (*models.Group, error)
	GetGroupByID(id int) (*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id int) error
	GetAllGroups() ([]models.Group, error)
}

// GroupServiceImpl implements GroupService.
type GroupServiceImpl struct {
	groupStore data.GroupStore
	validate   *validator.Validate
}

// NewGroupService creates a new GroupServiceImpl.
func NewGroupService(groupStore data.GroupStore) *GroupServiceImpl {
	return &GroupServiceImpl{
		groupStore: groupStore,
		validate:   validator.New(),
	}
}

// CreateGroup creates a new group.
func (s *GroupServiceImpl) CreateGroup(group *models.Group) (*models.Group, error) {
	if err := models.ValidateGroup(*group); err != nil {
		return nil, ErrInvalidInput
	}

	// Check for unique group name
	existingGroup, err := s.groupStore.GetByName(group.Name)
	if err == nil && existingGroup != nil {
		return nil, ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		log.Printf("Error checking group name uniqueness: %v", err)
		return nil, ErrInternal
	}

	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	id, err := s.groupStore.Create(group)
	if err != nil {
		return nil, ErrInternal
	}
	group.ID = id
	return group, nil
}

// GetGroupByID fetches a group by ID.
func (s *GroupServiceImpl) GetGroupByID(id int) (*models.Group, error) {
	group, err := s.groupStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return group, nil
}

// UpdateGroup updates an existing group.
func (s *GroupServiceImpl) UpdateGroup(group *models.Group) error {
	if err := models.ValidateGroup(*group); err != nil {
		return ErrInvalidInput
	}

	// Check for unique group name if name is changed
	existingGroup, err := s.groupStore.GetByName(group.Name)
	if err == nil && existingGroup != nil && existingGroup.ID != group.ID {
		return ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, data.ErrNotFound) {
		return ErrInternal
	}

	group.UpdatedAt = time.Now()
	err = s.groupStore.Update(group)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteGroup deletes a group by ID.
func (s *GroupServiceImpl) DeleteGroup(id int) error {
	err := s.groupStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// GetAllGroups fetches all groups.
func (s *GroupServiceImpl) GetAllGroups() ([]models.Group, error) {
	groups, err := s.groupStore.GetAll()
	if err != nil {
		return nil, ErrInternal
	}
	return groups, nil
}