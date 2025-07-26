package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kitadoc-backend/config"
	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// AudioRecordingHandler handles audio recording-related HTTP requests.
type AudioRecordingHandler struct {
	AudioAnalysisService      services.AudioAnalysisService
	DocumentationEntryService services.DocumentationEntryService // Add DocumentationEntryService
	Config                    *config.Config
}

// NewAudioRecordingHandler creates a new AudioRecordingHandler.
func NewAudioRecordingHandler(audioAnalysisService services.AudioAnalysisService, documentationEntryService services.DocumentationEntryService, cfg *config.Config) *AudioRecordingHandler {
	return &AudioRecordingHandler{
		AudioAnalysisService:      audioAnalysisService,
		DocumentationEntryService: documentationEntryService,
		Config:                    cfg,
	}
}

// Helper methods for error handling
func (h *AudioRecordingHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
}

func (h *AudioRecordingHandler) writeBadRequestError(w http.ResponseWriter, message string) {
	h.writeErrorResponse(w, http.StatusBadRequest, message)
}

func (h *AudioRecordingHandler) writeInternalServerError(w http.ResponseWriter, message string) {
	h.writeErrorResponse(w, http.StatusInternalServerError, message)
}

// UploadAudio handles uploading an audio recording and forwarding it to the audio-proc service.
func (h *AudioRecordingHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLoggerWithReqID(r.Context())
	logger.Info("Starting audio upload process")

	// 1. Parse multipart form data with file size limit
	logger.Info("Parsing multipart form")
	maxUploadSize := int64(h.Config.FileStorage.MaxSizeMB) << 20 // Convert MB to bytes
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.WithError(err).Error("Failed to parse multipart form or file size exceeded limit")
		h.writeBadRequestError(w, fmt.Sprintf("Failed to parse multipart form or file size exceeded limit (%d MB): %v", h.Config.FileStorage.MaxSizeMB, err))
		return
	}
	logger.Info("Successfully parsed multipart form")

	// 2. Get the file, teacher_id, and timestamp from the form
	logger.Info("Retrieving file and form values")
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		logger.WithError(err).Error("Error retrieving audio file from form")
		h.writeBadRequestError(w, "Error retrieving audio file: "+err.Error())
		return
	}
	defer file.Close()

	teacherID := r.FormValue("teacher_id")
	if teacherID == "" {
		logger.Warn("teacher_id is missing from the request")
		h.writeBadRequestError(w, "teacher_id is required")
		return
	}

	timestampStr := r.FormValue("timestamp")
	if timestampStr == "" {
		logger.Warn("timestamp is missing from the request")
		h.writeBadRequestError(w, "timestamp is required")
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		logger.WithError(err).Error("Invalid timestamp format")
		h.writeBadRequestError(w, "Invalid timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
		return
	}
	logger.Infof("Received file: %s, teacher_id: %s, timestamp: %s", fileHeader.Filename, teacherID, timestampStr)

	// 3. Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	logger.Infof("Validating file type: %s", contentType)
	if !h.isAllowedFileType(contentType) {
		logger.WithField("content_type", contentType).Warn("Disallowed file type uploaded")
		h.writeBadRequestError(w, fmt.Sprintf("Disallowed file type: %s. Allowed types are: %s", contentType, strings.Join(h.Config.FileStorage.AllowedTypes, ", ")))
		return
	}

	// 4. Read the file content into a byte slice
	logger.Info("Reading file content")
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.WithError(err).Error("Failed to read audio file content")
		h.writeInternalServerError(w, "Failed to read audio file content: "+err.Error())
		return
	}
	logger.Infof("Successfully read %d bytes from file", len(fileContent))

	// Respond immediately to the client
	w.WriteHeader(http.StatusOK)
	logger.Info("Sent 200 OK response to client")

	// Perform analysis and persistence in a goroutine
	go func() {
		// Use a new context for the goroutine if the original request context might be cancelled
		// or if you need a longer timeout for the background task.
		// For simplicity, we're using a background context here.
		ctx := context.Background()

		// 5. Call the service layer to analyze the audio
		logger.Info("Calling audio analysis service (async)")
		analysisResult, err := h.AudioAnalysisService.AnalyzeAudio(ctx, fileContent, fileHeader.Filename)
		if err != nil {
			logger.WithError(err).Error("Failed to analyze audio (async)")
			return
		}
		logger.WithField("analysis_result", analysisResult).Debug("Audio analysis result (async)")

		// 6. Persist the analysis result as a documentation entry
		logger.Info("Persisting analysis result (async)")
		teacherIDInt, err := strconv.Atoi(teacherID)
		if err != nil {
			logger.WithError(err).Error("Invalid teacher ID (async)")
			return
		}

		if analysisResult.NumberOfEntries == 0 {
			logger.Warn("No analysis results found (async)")
			return
		}

		for _, childAnalysis := range analysisResult.AnalysisResults {
			docEntry := models.DocumentationEntry{
				TeacherID:              teacherIDInt,
				ObservationDate:        timestamp,
				ObservationDescription: childAnalysis.TranscriptionSummary,
				CategoryID:             childAnalysis.Category.AnalysisCategoryID,
				ChildID:                childAnalysis.ChildID,
				IsApproved:             false,
			}

			_, err := h.DocumentationEntryService.CreateDocumentationEntry(logger, ctx, &docEntry)
			if err != nil {
				logger.WithError(err).Error("Failed to create documentation entry from audio analysis (async)")
				return
			}
		}
		logger.Info("Finished asynchronous audio processing")
	}()

	logger.Info("Finished audio upload process (handler)")
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
