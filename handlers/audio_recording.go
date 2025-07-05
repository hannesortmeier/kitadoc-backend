package handlers

import (
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

	// 1. Parse multipart form data with file size limit
	maxUploadSize := int64(h.Config.FileStorage.MaxSizeMB) << 20 // Convert MB to bytes
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.WithError(err).Error("Failed to parse multipart form or file size exceeded limit")
		h.writeBadRequestError(w, fmt.Sprintf("Failed to parse multipart form or file size exceeded limit (%d MB): %v", h.Config.FileStorage.MaxSizeMB, err))
		return
	}

	// 2. Get the file, teacher_id, and timestamp from the form
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		logger.WithError(err).Error("Error retrieving audio file from form")
		h.writeBadRequestError(w, "Error retrieving audio file: "+err.Error())
		return
	}
	if err := file.Close(); err != nil {
		logger.WithError(err).Error("Failed to close file")
	}

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

	// 3. Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !h.isAllowedFileType(contentType) {
		logger.WithField("content_type", contentType).Warn("Disallowed file type uploaded")
		h.writeBadRequestError(w, fmt.Sprintf("Disallowed file type: %s. Allowed types are: %s", contentType, strings.Join(h.Config.FileStorage.AllowedTypes, ", ")))
		return
	}

	// 4. Read the file content into a byte slic
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.WithError(err).Error("Failed to read audio file content")
		h.writeInternalServerError(w, "Failed to read audio file content: "+err.Error())
		return
	}

	// 5. Call the service layer to analyze the audio
	analysisResult, err := h.AudioAnalysisService.AnalyzeAudio(r.Context(), fileContent, fileHeader.Filename)
	if err != nil {
		logger.WithError(err).Error("Failed to analyze audio")
		h.writeInternalServerError(w, fmt.Sprintf("Failed to analyze audio: %v", err))
		return
	}

	// 6. Persist the analysis result as a documentation entry
	// Convert teacherID to int
	teacherIDInt, err := strconv.Atoi(teacherID)
	if err != nil {
		logger.WithError(err).Error("Invalid teacher ID")
		h.writeBadRequestError(w, "Invalid teacher ID")
		return
	}

	// Loop through analysis results and create documentation entries
	if analysisResult.NumberOfEntries == 0 {
		logger.Warn("No analysis results found")
		h.writeErrorResponse(w, 442, "No analysis results found") // Custom status code for no results
		return
	}

	var documentationEntryIds []int

	for _, childAnalysis := range analysisResult.AnalysisResults {

		docEntry := models.DocumentationEntry{
			TeacherID:              teacherIDInt,
			ObservationDate:        timestamp,
			ObservationDescription: childAnalysis.TranscriptionSummary,
			CategoryID:             childAnalysis.Category.AnalysisCategoryID,
			ChildID:                childAnalysis.ChildID,
			IsApproved:             false,
		}

		createdEntry, err := h.DocumentationEntryService.CreateDocumentationEntry(logger, r.Context(), &docEntry)
		if err != nil {
			logger.WithError(err).Error("Failed to create documentation entry from audio analysis")
			h.writeInternalServerError(w, fmt.Sprintf("Failed to create documentation entry: %v", err))
			return
		}
		documentationEntryIds = append(documentationEntryIds, createdEntry.ID)
	}

	// 7. Return the ids of the created documentation entry to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"documentationEntryIds": documentationEntryIds,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
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
