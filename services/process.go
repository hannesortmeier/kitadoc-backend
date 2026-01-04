package services

import (
	"errors"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

type ProcessService interface {
	Create(status string) (*models.Process, error)
	Update(process *models.Process) error
	GetByID(id int) (*models.Process, error)
}

type ProcessServiceImpl struct {
	store data.ProcessStore
}

func NewProcessService(store data.ProcessStore) *ProcessServiceImpl {
	return &ProcessServiceImpl{
		store: store,
	}
}

func (s *ProcessServiceImpl) Create(status string) (*models.Process, error) {
	process, err := s.store.Create(&models.Process{Status: status})
	if err != nil {
		logger.GetGlobalLogger().Errorf("Failed to create process: %v", err)
		return nil, err
	}
	return process, nil
}

func (s *ProcessServiceImpl) Update(process *models.Process) error {
	if err := s.store.Update(process); err != nil {
		logger.GetGlobalLogger().Errorf("Failed to update process: %v", err)
		return err
	}
	return nil
}

func (s *ProcessServiceImpl) GetByID(id int) (*models.Process, error) {
	process, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.GetGlobalLogger().Errorf("Process not found: %v", err)
			return nil, data.ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Failed to get process by id: %v", err)
		return nil, err
	}
	return process, nil
}
