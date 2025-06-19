package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"kitadoc-backend/config"
	"kitadoc-backend/middleware"
	"kitadoc-backend/services"
)

// AudioRecordingHandler handles audio recording-related HTTP requests.
type AudioRecordingHandler struct {
	AudioRecordingService services.AudioRecordingService
	Config                *config.Config
}

// NewAudioRecordingHandler creates a new AudioRecordingHandler.
func NewAudioRecordingHandler(audioRecordingService services.AudioRecordingService, cfg *config.Config) *AudioRecordingHandler {
	return &AudioRecordingHandler{
		AudioRecordingService: audioRecordingService,
		Config:                cfg,
	}
}

// UploadAudio handles uploading an audio recording.
func (audioRecordingHandler *AudioRecordingHandler) UploadAudio(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())

	// 1. Parse multipart form data with file size limit
	maxUploadSize := int64(audioRecordingHandler.Config.FileStorage.MaxSizeMB) << 20 // Convert MB to bytes
	request.Body = http.MaxBytesReader(writer, request.Body, maxUploadSize)
	err := request.ParseMultipartForm(maxUploadSize)
	if err != nil {
		logger.WithError(err).Error("Failed to parse multipart form or file size exceeded limit")
		http.Error(writer, fmt.Sprintf("Failed to parse multipart form or file size exceeded limit (%d MB): %v", audioRecordingHandler.Config.FileStorage.MaxSizeMB, err), http.StatusBadRequest)
		return
	}

	// 2. Get the file from the form
	file, fileHeader, err := request.FormFile("audio")
	if err != nil {
		logger.WithError(err).Error("Error retrieving audio file from form")
		http.Error(writer, "Error retrieving audio file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !audioRecordingHandler.isAllowedFileType(contentType) {
		logger.WithField("content_type", contentType).Warn("Disallowed file type uploaded")
		http.Error(writer, fmt.Sprintf("Disallowed file type: %s. Allowed types are: %s", contentType, strings.Join(audioRecordingHandler.Config.FileStorage.AllowedTypes, ", ")), http.StatusBadRequest)
		return
	}

	// 4. Create a temporary file to save the uploaded audio
	tempDir := audioRecordingHandler.Config.FileStorage.UploadDir
	// The directory is ensured to exist during config validation, so no need to check here.
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("audio_%d_%s", time.Now().UnixNano(), fileHeader.Filename))
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		logger.WithError(err).Error("Failed to create temporary file for audio upload")
		http.Error(writer, "Failed to create temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		// Delete the temporary file immediately after processing
		if err := os.Remove(tempFilePath); err != nil {
			logger.WithError(err).WithField("file_path", tempFilePath).Error("Failed to delete temporary audio file")
		} else {
			logger.WithField("file_path", tempFilePath).Info("Temporary audio file deleted successfully")
		}
	}()

	// 5. Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		logger.WithError(err).Error("Failed to save audio file to temporary location")
		http.Error(writer, "Failed to save audio file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Call the service layer to process the audio (e.g., save metadata to DB)
	// In a real application, you would pass the tempFilePath to the service.
	// For now, we'll just return a success message.
	// Example:
	// audioRecording, err := audioRecordingHandler.AudioRecordingService.ProcessAudio(r.Context(), tempFilePath, r.FormValue("child_id"), r.FormValue("teacher_id"))
	// if err != nil {
	//     logger.WithError(err).Error("Failed to process audio recording in service")
	//     http.Error(writer, "Failed to process audio recording", http.StatusInternalServerError)
	//     return
	// }
	logger.WithFields(logrus.Fields{
		"filename": fileHeader.Filename,
		"size":     fileHeader.Size,
		"temp_path": tempFilePath,
	}).Info("Audio uploaded and processed successfully (temporary file deleted)")

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{
		"message":  "Audio uploaded successfully and processed. Temporary file deleted.",
		"filename": fileHeader.Filename,
	})
}

// isAllowedFileType checks if the uploaded file's content type is allowed.
func (audioRecordingHandler *AudioRecordingHandler) isAllowedFileType(contentType string) bool {
	for _, allowedType := range audioRecordingHandler.Config.FileStorage.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}