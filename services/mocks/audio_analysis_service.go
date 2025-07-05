package mocks

import (
	"context"

	"kitadoc-backend/models"

	"github.com/stretchr/testify/mock"
)

// MockAudioAnalysisService is a mock of AudioAnalysisService.
type MockAudioAnalysisService struct {
	mock.Mock
}

// AnalyzeAudio is a mock of the AnalyzeAudio method.
func (m *MockAudioAnalysisService) AnalyzeAudio(ctx context.Context, fileContent []byte, filename string) (models.AnalysisResult, error) {
	args := m.Called(ctx, fileContent, filename)
	if args.Get(0) == nil {
		return models.AnalysisResult{}, args.Error(1)
	}
	return args.Get(0).(models.AnalysisResult), args.Error(1)
}
