package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"kitadoc-backend/config"
	"kitadoc-backend/middleware"
	"kitadoc-backend/services"
)

// AudioRecordingHandler handles audio recording-related HTTP requests.
type AudioRecordingHandler struct {
	AudioAnalysisService services.AudioAnalysisService
	Config               *config.Config
}

// NewAudioRecordingHandler creates a new AudioRecordingHandler.
func NewAudioRecordingHandler(audioAnalysisService services.AudioAnalysisService, cfg *config.Config) *AudioRecordingHandler {
	return &AudioRecordingHandler{
		AudioAnalysisService: audioAnalysisService,
		Config:               cfg,
	}
}

// UploadAudio handles uploading an audio recording and forwarding it to the audio-proc service.
func (h *AudioRecordingHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLoggerWithReqID(r.Context())

	// 1. Parse multipart form data with file size limit
	maxUploadSize := int64(h.Config.FileStorage.MaxSizeMB) << 20 // Convert MB to bytes
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.WithError(err).Error("Failed to parse multipart form or file size exceeded limit")
		http.Error(w, fmt.Sprintf("Failed to parse multipart form or file size exceeded limit (%d MB): %v", h.Config.FileStorage.MaxSizeMB, err), http.StatusBadRequest)
		return
	}

	// 2. Get the file from the form
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		logger.WithError(err).Error("Error retrieving audio file from form")
		http.Error(w, "Error retrieving audio file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !h.isAllowedFileType(contentType) {
		logger.WithField("content_type", contentType).Warn("Disallowed file type uploaded")
		http.Error(w, fmt.Sprintf("Disallowed file type: %s. Allowed types are: %s", contentType, strings.Join(h.Config.FileStorage.AllowedTypes, ", ")), http.StatusBadRequest)
		return
	}

	// 4. Read the file content into a byte slice
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.WithError(err).Error("Failed to read audio file content")
		http.Error(w, "Failed to read audio file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Call the service layer to analyze the audio
	analysisResult, err := h.AudioAnalysisService.AnalyzeAudio(r.Context(), fileContent, fileHeader.Filename)
	if err != nil {
		logger.WithError(err).Error("Failed to analyze audio")
		http.Error(w, fmt.Sprintf("Failed to analyze audio: %v", err), http.StatusInternalServerError)
		return
	}

	// 6. Return the analysis result to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(analysisResult); err != nil {
		logger.WithError(err).Error("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// isAllowedFileType checks if the uploaded file's content type is allowed.
func (h *AudioRecordingHandler) isAllowedFileType(contentType string) bool {
	for _, allowedType := range h.Config.FileStorage.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
