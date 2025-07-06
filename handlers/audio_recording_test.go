package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"testing"
	"time"

	"kitadoc-backend/config"
	"kitadoc-backend/handlers"
	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/models"
	services_mocks "kitadoc-backend/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAudioRecordingHandler_UploadAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockAudioAnalysisService := &services_mocks.MockAudioAnalysisService{}
		mockDocEntryService := &mocks.MockDocumentationEntryService{}
		h := handlers.NewAudioRecordingHandler(mockAudioAnalysisService, mockDocEntryService, &config.Config{
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
		_, err := part.Write([]byte("dummy audio data"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/audio/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add teacher_id and timestamp to the request form
		form := url.Values{}
		form.Add("teacher_id", "1")
		form.Add("timestamp", time.Now().Format(time.RFC3339))
		req.PostForm = form

		mockResponse := models.AnalysisResult{
			NumberOfEntries: 1,
			AnalysisResults: []models.ChildAnalysisObject{
				{
					ChildID:              1,
					FirstName:            "John",
					LastName:             "Doe",
					TranscriptionSummary: "Test transcription summary",
					Category:             models.AnalysisCategory{},
				},
			},
		}
		mockAudioAnalysisService.On("AnalyzeAudio", ctx, []byte("dummy audio data"), "test.wav").Return(mockResponse, nil).Once()
		mockDocEntryService.On("CreateDocumentationEntry", mock.Anything, ctx, mock.Anything).Return(&models.DocumentationEntry{ID: 1}, nil).Once()

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string][]int
		err = json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, map[string][]int{"documentationEntryIds": []int{1}}, response)
		mockAudioAnalysisService.AssertExpectations(t)
		mockDocEntryService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockAudioAnalysisService := &services_mocks.MockAudioAnalysisService{}
		mockDocEntryService := &mocks.MockDocumentationEntryService{}
		h := handlers.NewAudioRecordingHandler(mockAudioAnalysisService, mockDocEntryService, &config.Config{
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
		_, err := part.Write([]byte("dummy audio data"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/audio/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add teacher_id and timestamp to the request form
		form := url.Values{}
		form.Add("teacher_id", "1")
		req.PostForm = form

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
