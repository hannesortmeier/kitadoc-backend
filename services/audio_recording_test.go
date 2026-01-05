package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitadoc-backend/data/mocks"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAudioAnalysisService_AnalyzeAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		transcriptionResult := "hello world foo bar"
		mockTranscriptionService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(transcriptionResult)
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockTranscriptionService.Close() })

		analysisResult := models.AnalysisResult{
			NumberOfEntries: 1,
			AnalysisResults: []models.ChildAnalysisObject{
				{
					ChildID:              1,
					FirstName:            "John",
					LastName:             "Doe",
					TranscriptionSummary: "hello world",
					Category: models.AnalysisCategory{
						AnalysisCategoryID:   1,
						AnalysisCategoryName: "General",
					},
				},
			},
		}
		mockLLMAnalysisService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(analysisResult)
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockLLMAnalysisService.Close() })

		mockChildStore := new(mocks.MockChildStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockProcessStore := new(mocks.MockProcessStore)

		mockChildStore.On("GetAll").Return([]models.Child{{ID: 1, FirstName: "John", LastName: "Doe"}}, nil)
		description := ""
		mockCategoryStore.On("GetAll").Return([]models.Category{{ID: 1, Name: "General", Description: &description}}, nil)

		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "transcribing"
		})).Return(nil)

		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "analysing"
		})).Return(nil)

		service := services.NewAudioAnalysisService(
			mockLLMAnalysisService.Client(),
			mockTranscriptionService.URL,
			mockLLMAnalysisService.URL,
			mockChildStore,
			mockCategoryStore,
			mockProcessStore,
		)

		fileContent := []byte("dummy audio data")
		processId := 42

		result, err := service.ProcessAudio(ctx, logrus.NewEntry(logrus.New()), processId, fileContent)

		assert.NoError(t, err)
		assert.Equal(t, analysisResult, result)
	})

	t.Run("http client do failed", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This handler will not be called
		}))
		t.Cleanup(func() { mockServer.Close() })

		mockChildStore := new(mocks.MockChildStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockProcessStore := new(mocks.MockProcessStore)

		mockChildStore.On("GetAll").Return([]models.Child{}, nil)
		mockCategoryStore.On("GetAll").Return([]models.Category{}, nil)

		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "transcribing"
		})).Return(nil)

		service := services.NewAudioAnalysisService(
			mockServer.Client(),
			"http://invalid-url",
			"http://invalid-url2",
			mockChildStore,
			mockCategoryStore,
			mockProcessStore,
		)

		fileContent := []byte("dummy audio data")
		processId := 42

		result, err := service.ProcessAudio(ctx, logrus.NewEntry(logrus.New()), processId, fileContent)

		assert.Error(t, err)
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})

	t.Run("llm analysis service returned non-ok status", func(t *testing.T) {
		// Mock Transcription Service (Success)
		transcriptionResult := "hello world"
		mockTranscriptionService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(transcriptionResult)
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockTranscriptionService.Close() })

		// Mock LLM Analysis Service (Failure)
		mockLLMAnalysisService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(func() { mockLLMAnalysisService.Close() })

		mockChildStore := new(mocks.MockChildStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockProcessStore := new(mocks.MockProcessStore)

		// Expectations for AnalyzeTranscription
		mockChildStore.On("GetAll").Return([]models.Child{}, nil)
		mockCategoryStore.On("GetAll").Return([]models.Category{}, nil)

		// Expect process updates
		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "transcribing"
		})).Return(nil)
		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "analysing"
		})).Return(nil)

		service := services.NewAudioAnalysisService(
			mockTranscriptionService.Client(),
			mockTranscriptionService.URL,
			mockLLMAnalysisService.URL,
			mockChildStore,
			mockCategoryStore,
			mockProcessStore,
		)

		fileContent := []byte("dummy audio data")
		processId := 42

		result, err := service.ProcessAudio(ctx, logrus.NewEntry(logrus.New()), processId, fileContent)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "llm analysis service returned status 500")
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})

	t.Run("failed to decode response from transcription service", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("invalid json"))
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockServer.Close() })

		mockChildStore := new(mocks.MockChildStore)
		mockCategoryStore := new(mocks.MockCategoryStore)
		mockProcessStore := new(mocks.MockProcessStore)

		// Process status update expected for transcription
		mockProcessStore.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == 42 && p.Status == "transcribing"
		})).Return(nil)

		service := services.NewAudioAnalysisService(
			mockServer.Client(),
			mockServer.URL,
			"http://unused-url",
			mockChildStore,
			mockCategoryStore,
			mockProcessStore,
		)

		fileContent := []byte("dummy audio data")
		processId := 42

		result, err := service.ProcessAudio(ctx, logrus.NewEntry(logrus.New()), processId, fileContent)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})
}
