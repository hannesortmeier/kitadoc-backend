package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"kitadoc-backend/config"
	"kitadoc-backend/handlers/mocks"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupAudioRecordingTest(t *testing.T) (*mocks.AudioRecordingService, *AudioRecordingHandler, *config.Config, string) {
	mockService := new(mocks.AudioRecordingService)
	cfg := &config.Config{} // Initialize empty config
	cfg.FileStorage.UploadDir = filepath.Join(os.TempDir(), "audio_uploads")
	cfg.FileStorage.MaxSizeMB = 10 // Set to 10 MB for testing
	cfg.FileStorage.AllowedTypes = []string{"audio/mpeg", "audio/wav"}

	handler := NewAudioRecordingHandler(mockService, cfg)

	// Ensure upload directory exists
	err := os.MkdirAll(cfg.FileStorage.UploadDir, 0755)
	assert.NoError(t, err)

	return mockService, handler, cfg, cfg.FileStorage.UploadDir
}

func TestUploadAudio(t *testing.T) {
	_, handler, cfg, uploadDir := setupAudioRecordingTest(t)

	t.Run("success", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		// Create part with custom content type
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="audio"; filename="test_audio.mp3"`)
		header.Set("Content-Type", "audio/mpeg")
		part, err := writer.CreatePart(header)
		assert.NoError(t, err)
		_, err = io.Copy(part, bytes.NewBufferString("mock audio content"))
		assert.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-audio", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Add logger to context
		ctx := context.WithValue(req.Context(), "logger", logrus.NewEntry(logrus.New())) // Use a generic key for testing
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.UploadAudio(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		var response map[string]string
		json.NewDecoder(recorder.Body).Decode(&response)
		assert.Contains(t, response["message"], "Audio uploaded successfully")
		assert.Equal(t, "test_audio.mp3", response["filename"])

		// Assert that the temporary file was created and then deleted
		files, err := os.ReadDir(uploadDir)
		assert.NoError(t, err)
		assert.Len(t, files, 0, "Temporary file should have been deleted")
	})

	t.Run("file size exceeded limit", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("audio", "large_audio.mp3")
		assert.NoError(t, err)
		// Write content larger than MaxSizeMB
		_, err = io.Copy(part, bytes.NewBuffer(make([]byte, (cfg.FileStorage.MaxSizeMB+1)<<20))) // MaxSizeMB + 1 MB
		assert.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-audio", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Add logger to context
		ctx := context.WithValue(req.Context(), "logger", logrus.NewEntry(logrus.New())) // Use a generic key for testing
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.UploadAudio(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "file size exceeded limit")
	})

	t.Run("error retrieving audio file", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		// Do not create "audio" field
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-audio", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Add logger to context
		ctx := context.WithValue(req.Context(), "logger", logrus.NewEntry(logrus.New())) // Use a generic key for testing
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.UploadAudio(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Error retrieving audio file")
	})

	t.Run("disallowed file type", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("audio", "document.pdf")
		assert.NoError(t, err)
		_, err = io.Copy(part, bytes.NewBufferString("mock pdf content"))
		assert.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-audio", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		// Manually set Content-Type to a disallowed type
		req.Header.Set("Content-Type", "application/pdf")
		
		// Add logger to context
		ctx := context.WithValue(req.Context(), "logger", logrus.NewEntry(logrus.New())) // Use a generic key for testing
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.UploadAudio(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Content-Type isn't multipart/form-data")
	})

	t.Run("failed to create temporary file", func(t *testing.T) {
		// Simulate an error by making the upload directory read-only
		os.Chmod(uploadDir, 0444)
		defer os.Chmod(uploadDir, 0755) // Revert permissions after test

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		// Create part with custom content type
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="audio"; filename="test_audio.mp3"`)
		header.Set("Content-Type", "audio/mpeg")
		part, err := writer.CreatePart(header)
		assert.NoError(t, err)
		_, err = io.Copy(part, bytes.NewBufferString("mock audio content"))
		assert.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-audio", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Add logger to context
		ctx := context.WithValue(req.Context(), "logger", logrus.NewEntry(logrus.New())) // Use a generic key for testing
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.UploadAudio(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to create temporary file")
	})
}

func TestIsAllowedFileType(t *testing.T) {
	_, handler, _, _ := setupAudioRecordingTest(t)

	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"allowed mpeg", "audio/mpeg", true},
		{"allowed wav", "audio/wav", true},
		{"disallowed pdf", "application/pdf", false},
		{"disallowed text", "text/plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, handler.isAllowedFileType(tt.contentType))
		})
	}
}
