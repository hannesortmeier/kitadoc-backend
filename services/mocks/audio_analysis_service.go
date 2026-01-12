package mocks

import (
	"context"

	"kitadoc-backend/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// MockAudioAnalysisService is a mock of AudioAnalysisService.
type MockAudioAnalysisService struct {
	mock.Mock
}

// ProcessAudio is a mock of the ProcessAudio method.
func (m *MockAudioAnalysisService) ProcessAudio(
	ctx context.Context,
	logger *logrus.Entry,
	processId int,
	fileContent []byte,
) ([]models.ChildAnalysisObject, error) {
	args := m.Called(ctx, logger, processId, fileContent)
	if args.Get(0) == nil {
		return []models.ChildAnalysisObject{}, args.Error(1)
	}
	return args.Get(0).([]models.ChildAnalysisObject), args.Error(1)
}
