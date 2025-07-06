package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/stretchr/testify/assert"
)

func TestAudioAnalysisService_AnalyzeAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(analysisResult)
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockServer.Close() })

		service := services.NewAudioAnalysisService(mockServer.Client(), mockServer.URL)

		fileContent := []byte("dummy audio data")
		filename := "test.wav"

		result, err := service.AnalyzeAudio(ctx, fileContent, filename)

		assert.NoError(t, err)
		assert.Equal(t, analysisResult, result)
	})

	t.Run("http request creation failed", func(t *testing.T) {
		// This is hard to test in isolation, more of an integration test concern.
	})

	t.Run("http client do failed", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This handler will not be called
		}))
		t.Cleanup(func() { mockServer.Close() })

		service := services.NewAudioAnalysisService(mockServer.Client(), "http://invalid-url")

		fileContent := []byte("dummy audio data")
		filename := "test.wav"

		result, err := service.AnalyzeAudio(ctx, fileContent, filename)

		assert.Error(t, err)
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})

	t.Run("audio-proc service returned non-ok status", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(func() { mockServer.Close() })

		service := services.NewAudioAnalysisService(mockServer.Client(), mockServer.URL)

		fileContent := []byte("dummy audio data")
		filename := "test.wav"

		result, err := service.AnalyzeAudio(ctx, fileContent, filename)

		assert.Error(t, err)
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})

	t.Run("failed to decode response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("invalid json"))
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockServer.Close() })

		service := services.NewAudioAnalysisService(mockServer.Client(), mockServer.URL)

		fileContent := []byte("dummy audio data")
		filename := "test.wav"

		result, err := service.AnalyzeAudio(ctx, fileContent, filename)

		assert.Error(t, err)
		assert.Equal(t, models.AnalysisResult{NumberOfEntries: 0, AnalysisResults: []models.ChildAnalysisObject(nil)}, result)
	})
}
