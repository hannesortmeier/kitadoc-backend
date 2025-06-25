package services

import (
	"errors"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// AudioRecordingService defines the interface for audio recording-related business logic operations.
type AudioRecordingService interface {
	UploadAudioRecording(recording *models.AudioRecording, fileContent []byte) (*models.AudioRecording, error)
	GetAudioRecordingByID(id int) (*models.AudioRecording, error)
	DeleteAudioRecording(id int) error
}

// AudioRecordingServiceImpl implements AudioRecordingService.
type AudioRecordingServiceImpl struct {
	audioRecordingStore     data.AudioRecordingStore
	documentationEntryStore data.DocumentationEntryStore
	validate                *validator.Validate
}

// NewAudioRecordingService creates a new AudioRecordingServiceImpl.
func NewAudioRecordingService(audioRecordingStore data.AudioRecordingStore, documentationEntryStore data.DocumentationEntryStore) *AudioRecordingServiceImpl {
	return &AudioRecordingServiceImpl{
		audioRecordingStore:     audioRecordingStore,
		documentationEntryStore: documentationEntryStore,
		validate:                validator.New(),
	}
}

// UploadAudioRecording handles the upload and creation of an audio recording.
func (service *AudioRecordingServiceImpl) UploadAudioRecording(recording *models.AudioRecording, fileContent []byte) (*models.AudioRecording, error) {
	if err := service.validate.Struct(recording); err != nil {
		return nil, err
	}

	// Validate DocumentationEntryID
	_, err := service.documentationEntryStore.GetByID(recording.DocumentationEntryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("documentation entry not found")
		}
		return nil, ErrInternal
	}

	// In a real application, you would save the fileContent to a storage system (e.g., local disk, S3).
	// The FilePath in the model would then store the path/URL to this saved file.
	// For this example, we'll just simulate success and set a dummy file path.
	if len(fileContent) == 0 {
		return nil, errors.New("audio file content is empty")
	}
	recording.FilePath = "/path/to/uploaded/audio/" + time.Now().Format("20060102150405") + ".mp3" // Dummy path

	recording.CreatedAt = time.Now()

	id, err := service.audioRecordingStore.Create(recording)
	if err != nil {
		return nil, ErrInternal
	}
	recording.ID = id
	return recording, nil
}

// GetAudioRecordingByID fetches an audio recording by ID.
func (s *AudioRecordingServiceImpl) GetAudioRecordingByID(id int) (*models.AudioRecording, error) {
	recording, err := s.audioRecordingStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return recording, nil
}

// DeleteAudioRecording deletes an audio recording by ID.
func (service *AudioRecordingServiceImpl) DeleteAudioRecording(id int) error {
	// In a real application, you would also delete the actual file from storage here.
	err := service.audioRecordingStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}
