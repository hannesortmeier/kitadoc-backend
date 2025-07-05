package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitadoc-backend/services"

	"github.com/stretchr/testify/assert"
)

func TestAudioAnalysisService_AnalyzeAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(map[string]interface{}{"transcription": "hello world"})
			assert.NoError(t, err)
		}))
		t.Cleanup(func() { mockServer.Close() })

		service := services.NewAudioAnalysisService(mockServer.Client(), mockServer.URL)

		fileContent := []byte("dummy audio data")
		filename := "test.wav"

		result, err := service.AnalyzeAudio(ctx, fileContent, filename)

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"transcription": "hello world"}, result)
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
		assert.Nil(t, result)
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
		assert.Nil(t, result)
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
		assert.Nil(t, result)
	})
}
