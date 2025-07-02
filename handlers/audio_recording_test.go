package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"kitadoc-backend/config"
	"kitadoc-backend/handlers"
	"kitadoc-backend/services/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAudioRecordingHandler_UploadAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockService := &mocks.MockAudioAnalysisService{}
		h := handlers.NewAudioRecordingHandler(mockService, &config.Config{
			FileStorage: struct {
				UploadDir    string   `mapstructure:"upload_dir"`
				MaxSizeMB    int      `mapstructure:"max_size_mb"`
				AllowedTypes []string `mapstructure:"allowed_types"`
			}{
				MaxSizeMB:    10,
				AllowedTypes: []string{"audio/wav", "audio/mpeg"},
			},
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="audio"; filename="test.wav"`)
		header.Set("Content-Type", "audio/wav")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("dummy audio data"))

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/audio/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		mockResponse := map[string]interface{}{"transcription": "hello world"}
		mockService.On("AnalyzeAudio", ctx, []byte("dummy audio data"), "test.wav").Return(mockResponse, nil).Once()

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&response)

		assert.Equal(t, mockResponse, response)
		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := &mocks.MockAudioAnalysisService{}
		h := handlers.NewAudioRecordingHandler(mockService, &config.Config{
			FileStorage: struct {
				UploadDir    string   `mapstructure:"upload_dir"`
				MaxSizeMB    int      `mapstructure:"max_size_mb"`
				AllowedTypes []string `mapstructure:"allowed_types"`
			}{
				MaxSizeMB:    10,
				AllowedTypes: []string{"audio/wav", "audio/mpeg"},
			},
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="audio"; filename="test.wav"`)
		header.Set("Content-Type", "audio/wav")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("dummy audio data"))

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/audio/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		mockService.On("AnalyzeAudio", ctx, []byte("dummy audio data"), "test.wav").Return(nil, assert.AnError).Once()

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}
