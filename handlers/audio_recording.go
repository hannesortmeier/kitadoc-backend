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

	"github.com/sirupsen/logrus"
)

// AudioRecordingHandler handles audio recording-related HTTP requests.
type AudioRecordingHandler struct {
	AudioAnalysisService      services.AudioAnalysisService
	DocumentationEntryService services.DocumentationEntryService
	ProcessService            services.ProcessService
	Config                    *config.Config
}

// NewAudioRecordingHandler creates a new AudioRecordingHandler.
func NewAudioRecordingHandler(
	audioAnalysisService services.AudioAnalysisService,
	documentationEntryService services.DocumentationEntryService,
	processService services.ProcessService,
	cfg *config.Config,
) *AudioRecordingHandler {
	return &AudioRecordingHandler{
		AudioAnalysisService:      audioAnalysisService,
		DocumentationEntryService: documentationEntryService,
		ProcessService:            processService,
		Config:                    cfg,
	}
}

// Helper methods for error handling
func (handler *AudioRecordingHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
}

func (handler *AudioRecordingHandler) writeBadRequestError(w http.ResponseWriter, message string) {
	handler.writeErrorResponse(w, http.StatusBadRequest, message)
}

func (handler *AudioRecordingHandler) writeInternalServerError(w http.ResponseWriter, message string) {
	handler.writeErrorResponse(w, http.StatusInternalServerError, message)
}

// UploadAudio handles uploading an audio recording starting the transcription and analysis process as a goroutine.
func (handler *AudioRecordingHandler) UploadAudio(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	logger.Info("Starting audio processing")

	// 1. Parse multipart form data with file size limit
	logger.Info("Parsing multipart form")
	maxUploadSize := int64(handler.Config.FileStorage.MaxSizeMB) << 20 // Convert MB to bytes
	request.Body = http.MaxBytesReader(writer, request.Body, maxUploadSize)
	if err := request.ParseMultipartForm(maxUploadSize); err != nil {
		logger.WithError(err).Error("Failed to parse multipart form or file size exceeded limit")
		handler.writeBadRequestError(writer, fmt.Sprintf("Failed to parse multipart form or file size exceeded limit (%d MB): %v", handler.Config.FileStorage.MaxSizeMB, err))
		return
	}
	logger.Info("Successfully parsed multipart form")

	// 2. Get the file, teacher_id, and timestamp from the form
	logger.Info("Retrieving file and form values")
	file, fileHeader, err := request.FormFile("audio")
	if err != nil {
		logger.WithError(err).Error("Error retrieving audio file from form")
		handler.writeBadRequestError(writer, "Error retrieving audio file: "+err.Error())
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.WithError(err).Error("Failed to close uploaded audio file")
		}
	}()

	teacherID := request.FormValue("teacher_id")
	if teacherID == "" {
		logger.Warn("teacher_id is missing from the request")
		handler.writeBadRequestError(writer, "teacher_id is required")
		return
	}

	timestampStr := request.FormValue("timestamp")
	if timestampStr == "" {
		logger.Warn("timestamp is missing from the request")
		handler.writeBadRequestError(writer, "timestamp is required")
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		logger.WithError(err).Error("Invalid timestamp format")
		handler.writeBadRequestError(writer, "Invalid timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)")
		return
	}
	logger.Infof("Received file: %s, teacher_id: %s, timestamp: %s", fileHeader.Filename, teacherID, timestampStr)

	// 3. Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	logger.Infof("Validating file type: %s", contentType)
	if !handler.isAllowedFileType(contentType) {
		logger.WithField("content_type", contentType).Warn("Disallowed file type uploaded")
		handler.writeBadRequestError(
			writer,
			fmt.Sprintf(
				"Disallowed file type: %s. Allowed types are: %s",
				contentType,
				strings.Join(handler.Config.FileStorage.AllowedTypes, ", "),
			),
		)
		return
	}

	// 4. Read the file content into a byte slice
	logger.Info("Reading file content")
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.WithError(err).Error("Failed to read audio file content")
		handler.writeInternalServerError(writer, "Failed to read audio file content: "+err.Error())
		return
	}
	logger.Infof("Successfully read %d bytes from file", len(fileContent))

	// Create a new process entry in the database that the client can poll
	process, err := handler.ProcessService.Create("starting")
	var processId int
	if err != nil {
		logger.WithError(err).Error("Failed to create process entry in database for polling")
		processId = -1
	} else {
		processId = process.ProcessId
	}

	// Respond immediately to the client with the process ID
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(writer).Encode(map[string]int{"process_id": processId}); err != nil {
		logger.WithError(err).Error("Failed to encode response")
	}
	logger.Info("Sent 202 Accepted response to client")

	// Perform analysis and persistence in a goroutine
	go func(processId int) {
		// Use a new context for the goroutine
		ctx := context.Background()

		// 5. Call the service layer to analyze the audio
		logger.Info("Calling audio analysis service to process the audio")
		analysisResult, err := handler.AudioAnalysisService.ProcessAudio(ctx, logger, processId, fileContent)
		if err != nil {
			logger.WithError(err).Error("Failed to analyze audio")
			handler.UpdateProcessStatus(logger, processId, "failed")
			return
		}
		logger.WithField("analysis_result", analysisResult).Debug("Audio analysis result")

		// 6. Persist the analysis result as a documentation entry
		handler.UpdateProcessStatus(logger, processId, "creating documentation entry")
		logger.Info("Persisting analysis result")
		teacherIDInt, err := strconv.Atoi(teacherID)
		if err != nil {
			logger.WithError(err).Error("Invalid teacher ID")
			handler.UpdateProcessStatus(logger, processId, "failed")
			return
		}

		if analysisResult.NumberOfEntries == 0 {
			logger.Warn("No analysis results found")
			handler.UpdateProcessStatus(logger, processId, "failed")
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

			_, err := handler.DocumentationEntryService.CreateDocumentationEntry(logger, ctx, &docEntry)
			if err != nil {
				logger.WithError(err).Error("Failed to create documentation entry from audio analysis")
				handler.UpdateProcessStatus(logger, processId, "failed")
				return
			}
		}
		logger.Info("Finished asynchronous audio processing")
		handler.UpdateProcessStatus(logger, processId, "completed")
	}(processId)

	logger.Info("Finished audio upload process (handler)")
}

// Checks if a process entry in the database was created and updates its status.
func (handler *AudioRecordingHandler) UpdateProcessStatus(logger *logrus.Entry, processId int, status string) {
	if processId != -1 {
		if err := handler.ProcessService.Update(&models.Process{ProcessId: processId, Status: status}); err != nil {
			logger.WithError(err).Error("Failed to update process status")
		}
	}
}

// isAllowedFileType checks if the uploaded file's content type is allowed.
func (handler *AudioRecordingHandler) isAllowedFileType(contentType string) bool {
	for _, allowedType := range handler.Config.FileStorage.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
