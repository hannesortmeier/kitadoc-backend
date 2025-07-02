package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockAudioAnalysisService is a mock of AudioAnalysisService.
type MockAudioAnalysisService struct {
	mock.Mock
}

// AnalyzeAudio is a mock of the AnalyzeAudio method.
func (m *MockAudioAnalysisService) AnalyzeAudio(ctx context.Context, fileContent []byte, filename string) (map[string]interface{}, error) {
	args := m.Called(ctx, fileContent, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
