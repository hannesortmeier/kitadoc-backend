package services

import (
	"errors"
	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

// KitaMasterdataService defines the interface for Kita master data-related business logic operations.
type KitaMasterdataService interface {
	GetKitaMasterdata() (*models.KitaMasterdata, error)
	UpdateKitaMasterdata(masterdata *models.KitaMasterdata) error
}

// KitaMasterdataServiceImpl implements KitaMasterdataService.
type KitaMasterdataServiceImpl struct {
	kitaMasterdataStore data.KitaMasterdataStore
}

// NewKitaMasterdataService creates a new KitaMasterdataServiceImpl.
func NewKitaMasterdataService(kitaMasterdataStore data.KitaMasterdataStore) *KitaMasterdataServiceImpl {
	return &KitaMasterdataServiceImpl{
		kitaMasterdataStore: kitaMasterdataStore,
	}
}

// GetKitaMasterdata fetches the Kita master data.
func (s *KitaMasterdataServiceImpl) GetKitaMasterdata() (*models.KitaMasterdata, error) {
	masterdata, err := s.kitaMasterdataStore.Get()
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.GetGlobalLogger().Info("Kita master data not found")
			return nil, ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Error fetching Kita master data: %v", err)
		return nil, ErrInternal
	}
	return masterdata, nil
}

// UpdateKitaMasterdata updates the Kita master data.
func (s *KitaMasterdataServiceImpl) UpdateKitaMasterdata(masterdata *models.KitaMasterdata) error {
	if err := models.ValidateKitaMasterdata(*masterdata); err != nil {
		logger.GetGlobalLogger().Errorf("Invalid Kita master data input: %v", err)
		return ErrInvalidInput
	}

	err := s.kitaMasterdataStore.Update(masterdata)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error updating Kita master data: %v", err)
		return ErrInternal
	}
	logger.GetGlobalLogger().Info("Kita master data updated successfully")
	return nil
}
