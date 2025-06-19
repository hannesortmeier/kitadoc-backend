package services_test

import (
	"errors"
	"testing"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/services/mocks"
	"kitadoc-backend/internal/logger"


	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sirupsen/logrus"
)

func TestUploadAudioRecording(t *testing.T) {
	log_level, _ := logrus.ParseLevel("debug")
	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	// Test case 1: Successful upload
	t.Run("success", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recording := &models.AudioRecording{
			DocumentationEntryID: 1,
			FilePath:             "./test.mp3",
			DurationSeconds: 	120,
		}
		fileContent := []byte("dummy audio data")
		expectedEntry := &models.DocumentationEntry{ID: 1}

		mockDocumentationEntryStore.On("GetByID", recording.DocumentationEntryID).Return(expectedEntry, nil).Once()
		mockAudioRecordingStore.On("Create", mock.AnythingOfType("*models.AudioRecording")).Return(1, nil).Once()

		createdRecording, err := service.UploadAudioRecording(recording, fileContent)

		assert.NoError(t, err)
		assert.NotNil(t, createdRecording)
		assert.Equal(t, 1, createdRecording.ID)
		assert.NotEmpty(t, createdRecording.FilePath)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockAudioRecordingStore.AssertExpectations(t)
	})

	// Test case 2: Invalid input (validation error)
	t.Run("invalid input", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recording := &models.AudioRecording{
			DocumentationEntryID: 1,
			FilePath:             "./test.mp3",
			DurationSeconds:             0, // Invalid duration
		}
		fileContent := []byte("dummy audio data")

		createdRecording, err := service.UploadAudioRecording(recording, fileContent)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Key: 'AudioRecording.DurationSeconds' Error:Field validation for 'DurationSeconds' failed on the 'min' ta")
		assert.Nil(t, createdRecording)
		mockDocumentationEntryStore.AssertNotCalled(t, "GetByID")
		mockAudioRecordingStore.AssertNotCalled(t, "Create")
	})

	// Test case 3: Documentation entry not found
	t.Run("documentation entry not found", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recording := &models.AudioRecording{
			DocumentationEntryID: 99, // Non-existent entry
			FilePath:             "test.mp3",
			DurationSeconds: 	120,
		}
		fileContent := []byte("dummy audio data")

		mockDocumentationEntryStore.On("GetByID", recording.DocumentationEntryID).Return(nil, data.ErrNotFound).Once()

		createdRecording, err := service.UploadAudioRecording(recording, fileContent)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "documentation entry not found")
		assert.Nil(t, createdRecording)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockAudioRecordingStore.AssertNotCalled(t, "Create")
	})

	// Test case 4: Internal error during documentation entry fetch
	t.Run("internal error on documentation entry fetch", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recording := &models.AudioRecording{
			DocumentationEntryID: 1,
			FilePath:             "test.mp3",
			DurationSeconds: 	120,
		}
		fileContent := []byte("dummy audio data")

		mockDocumentationEntryStore.On("GetByID", recording.DocumentationEntryID).Return(nil, errors.New("db error")).Once()

		createdRecording, err := service.UploadAudioRecording(recording, fileContent)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdRecording)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockAudioRecordingStore.AssertNotCalled(t, "Create")
	})

	// Test case 6: Internal error during audio recording creation
	t.Run("internal error on create", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recording := &models.AudioRecording{
			DocumentationEntryID: 1,
			FilePath:             "test.mp3",
			DurationSeconds: 	120,
		}
		fileContent := []byte("dummy audio data")
		expectedEntry := &models.DocumentationEntry{ID: 1}

		mockDocumentationEntryStore.On("GetByID", recording.DocumentationEntryID).Return(expectedEntry, nil).Once()
		mockAudioRecordingStore.On("Create", mock.AnythingOfType("*models.AudioRecording")).Return(0, errors.New("db error")).Once()

		createdRecording, err := service.UploadAudioRecording(recording, fileContent)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, createdRecording)
		mockDocumentationEntryStore.AssertExpectations(t)
		mockAudioRecordingStore.AssertExpectations(t)
	})
}

func TestGetAudioRecordingByID(t *testing.T) {
	// Test case 1: Successful retrieval
	t.Run("success", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 1
		expectedRecording := &models.AudioRecording{ID: recordingID, FilePath: "audio.mp3"}
		mockAudioRecordingStore.On("GetByID", recordingID).Return(expectedRecording, nil).Once()

		recording, err := service.GetAudioRecordingByID(recordingID)

		assert.NoError(t, err)
		assert.NotNil(t, recording)
		assert.Equal(t, expectedRecording.ID, recording.ID)
		mockAudioRecordingStore.AssertExpectations(t)
	})

	// Test case 2: Recording not found
	t.Run("not found", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 99
		mockAudioRecordingStore.On("GetByID", recordingID).Return(nil, data.ErrNotFound).Once()

		recording, err := service.GetAudioRecordingByID(recordingID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		assert.Nil(t, recording)
		mockAudioRecordingStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 1
		mockAudioRecordingStore.On("GetByID", recordingID).Return(nil, errors.New("db error")).Once()

		recording, err := service.GetAudioRecordingByID(recordingID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		assert.Nil(t, recording)
		mockAudioRecordingStore.AssertExpectations(t)
	})
}

func TestDeleteAudioRecording(t *testing.T) {
	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 1
		mockAudioRecordingStore.On("Delete", recordingID).Return(nil).Once()

		err := service.DeleteAudioRecording(recordingID)

		assert.NoError(t, err)
		mockAudioRecordingStore.AssertExpectations(t)
	})

	// Test case 2: Recording not found
	t.Run("not found", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 99
		mockAudioRecordingStore.On("Delete", recordingID).Return(data.ErrNotFound).Once()

		err := service.DeleteAudioRecording(recordingID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrNotFound, err)
		mockAudioRecordingStore.AssertExpectations(t)
	})

	// Test case 3: Internal error
	t.Run("internal error", func(t *testing.T) {
		mockAudioRecordingStore := new(mocks.MockAudioRecordingStore)
		mockDocumentationEntryStore := new(mocks.MockDocumentationEntryStore)
		service := services.NewAudioRecordingService(mockAudioRecordingStore, mockDocumentationEntryStore)

		recordingID := 1
		mockAudioRecordingStore.On("Delete", recordingID).Return(errors.New("db error")).Once()

		err := service.DeleteAudioRecording(recordingID)

		assert.Error(t, err)
		assert.Equal(t, services.ErrInternal, err)
		mockAudioRecordingStore.AssertExpectations(t)
	})
}