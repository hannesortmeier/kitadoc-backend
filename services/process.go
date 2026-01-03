package services

import (
	"errors"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

type ProcessService interface {
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

func (s *ProcessServiceImpl) GetByID(id int) (*models.Process, error) {
	process, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.GetGlobalLogger().Error("Process not found: %v", err)
			return nil, data.ErrNotFound
		}
		logger.GetGlobalLogger().Error("Failed to get process by id: %v", err)
		return nil, err
	}
	return process, nil
}
