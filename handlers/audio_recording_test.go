package handlers_test

import (
	"bytes"
	"context"
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
		mockProcessService := &mocks.MockProcessService{}
		h := handlers.NewAudioRecordingHandler(mockAudioAnalysisService, mockDocEntryService, mockProcessService, &config.Config{
			FileStorage: struct {
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
					Category: models.AnalysisCategory{
						AnalysisCategoryID: 1,
					},
				},
			},
		}

		done := make(chan bool, 1)

		processID := 42
		mockProcessService.On("Create", "starting").Return(&models.Process{ProcessId: processID, Status: "starting"}, nil).Once()

		mockAudioAnalysisService.On("ProcessAudio", mock.Anything, mock.AnythingOfType("*logrus.Entry"), processID, []byte("dummy audio data")).Return(mockResponse, nil).Once()

		mockProcessService.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == processID && p.Status == "creating documentation entry"
		})).Return(nil).Once()

		mockDocEntryService.On("CreateDocumentationEntry", mock.Anything, mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
			done <- true
		}).Once()

		mockProcessService.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == processID && p.Status == "completed"
		})).Return(nil).Once()

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusAccepted, rr.Code)

		// Wait for the goroutine to complete
		select {
		case <-done:
			// All good
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for documentation entry creation")
		}

		// Allow a small window for the final status update
		time.Sleep(50 * time.Millisecond)

		mockAudioAnalysisService.AssertExpectations(t)
		mockDocEntryService.AssertExpectations(t)
		mockProcessService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockAudioAnalysisService := &services_mocks.MockAudioAnalysisService{}
		mockDocEntryService := &mocks.MockDocumentationEntryService{}
		mockProcessService := &mocks.MockProcessService{}
		h := handlers.NewAudioRecordingHandler(mockAudioAnalysisService, mockDocEntryService, mockProcessService, &config.Config{
			FileStorage: struct {
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

	t.Run("analysis service error", func(t *testing.T) {
		mockAudioAnalysisService := &services_mocks.MockAudioAnalysisService{}
		mockDocEntryService := &mocks.MockDocumentationEntryService{}
		mockProcessService := &mocks.MockProcessService{}
		h := handlers.NewAudioRecordingHandler(mockAudioAnalysisService, mockDocEntryService, mockProcessService, &config.Config{
			FileStorage: struct {
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
		assert.NoError(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/api/v1/audio/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		form := url.Values{}
		form.Add("teacher_id", "1")
		form.Add("timestamp", time.Now().Format(time.RFC3339))
		req.PostForm = form

		done := make(chan bool, 1)
		processID := 124
		mockProcessService.On("Create", "starting").Return(&models.Process{ProcessId: processID, Status: "starting"}, nil).Once()

		mockAudioAnalysisService.On("ProcessAudio", mock.Anything, mock.AnythingOfType("*logrus.Entry"), processID, []byte("dummy audio data")).Return(models.AnalysisResult{}, assert.AnError).Once()

		mockProcessService.On("Update", mock.MatchedBy(func(p *models.Process) bool {
			return p.ProcessId == processID && p.Status == "failed"
		})).Return(nil).Run(func(args mock.Arguments) {
			done <- true
		}).Once()

		rr := httptest.NewRecorder()
		h.UploadAudio(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusAccepted, rr.Code)

		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for process status update to failed")
		}

		mockAudioAnalysisService.AssertExpectations(t)
		mockProcessService.AssertExpectations(t)
	})
}
