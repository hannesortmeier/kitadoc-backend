package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/sirupsen/logrus"
)

// AudioAnalysisService defines the interface for audio analysis operations.
type AudioAnalysisService interface {
	ProcessAudio(ctx context.Context, logger *logrus.Entry, processId int, fileContent []byte) (models.AnalysisResult, error)
}

// AudioAnalysisServiceImpl implements AudioAnalysisService.
type AudioAnalysisServiceImpl struct {
	httpClient              *http.Client
	transcriptionServiceURL string
	llmAnalysisServiceURL   string
	childStore              data.ChildStore
	categoryStore           data.CategoryStore
	processStore            data.ProcessStore
}

// NewAudioAnalysisService creates a new AudioAnalysisServiceImpl.
func NewAudioAnalysisService(
	httpClient *http.Client,
	transcriptionServiceURL string,
	llmAnalysisServiceURL string,
	childStore data.ChildStore,
	categoryStore data.CategoryStore,
	processStore data.ProcessStore,
) *AudioAnalysisServiceImpl {
	return &AudioAnalysisServiceImpl{
		httpClient:              httpClient,
		transcriptionServiceURL: transcriptionServiceURL,
		llmAnalysisServiceURL:   llmAnalysisServiceURL,
		childStore:              childStore,
		categoryStore:           categoryStore,
		processStore:            processStore,
	}
}

// ProcessAudio processes the audio file and returns the analysis result.
// The audio file is first transcribed by the transcription service, and
// then the transcription is analysed by the llm analysis service.
func (service *AudioAnalysisServiceImpl) ProcessAudio(
	ctx context.Context,
	logger *logrus.Entry,
	processId int,
	fileContent []byte,
) (models.AnalysisResult, error) {
	logger.Info("Starting audio transcription")

	service.UpdateProcessStatus(logger, processId, "transcribing")
	transcription, err := service.transcribeAudio(ctx, logger, fileContent)
	if err != nil {
		logger.WithError(err).Error("Failed to transcribe audio")
		return models.AnalysisResult{}, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	logger.Debugf("Transcription result: %s", transcription)
	logger.Info("Starting analysis of transcription")

	service.UpdateProcessStatus(logger, processId, "analysing")
	analysis, err := service.analyseTranscription(ctx, logger, transcription)
	if err != nil {
		logger.WithError(err).Error("Failed to analyse transcription")
		return models.AnalysisResult{}, fmt.Errorf("failed to analyse transcription: %w", err)
	}

	logger.Info("Completed audio processing")

	return analysis, nil
}

func (service *AudioAnalysisServiceImpl) transcribeAudio(ctx context.Context, logger *logrus.Entry, fileContent []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := fmt.Sprintf("audio_%s", time.Now().Format("20060102150405"))

	// Create a new form-data header with the provided filename.
	audioPart, err := writer.CreateFormFile("audio_file", filename)
	if err != nil {
		logger.WithError(err).Error("Failed to create form file")
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the file content to the form file.
	_, err = io.Copy(audioPart, bytes.NewReader(fileContent))
	if err != nil {
		logger.WithError(err).Error("Failed to copy file content to form file")
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the multipart writer.
	err = writer.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close multipart writer")
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	logger.Debugf("Request body size: %d", body.Len())

	// Create a new HTTP request to the transcription service.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, service.transcriptionServiceURL, body)
	if err != nil {
		logger.WithError(err).Error("Failed to create request to transcription service")
		return "", fmt.Errorf("failed to create request to transcription service: %w", err)
	}

	// Set the content type for the request.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request.
	resp, err := service.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("Failed to send request to transcription service")
		return "", fmt.Errorf("failed to send request to transcription service: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.WithError(err).Error("Failed to close response body")
		}
	}()

	// Check the response status code.
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		}).Error("Received non-OK response from transcription service")
		return "", fmt.Errorf("transcription service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response.
	var transcription string
	if err := json.NewDecoder(resp.Body).Decode(&transcription); err != nil {
		logger.WithError(err).Error("failed to decode response from transcription service")
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return transcription, nil
}

func (service *AudioAnalysisServiceImpl) analyseTranscription(ctx context.Context, logger *logrus.Entry, transcription string) (models.AnalysisResult, error) {
	// Create a new multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form field for the transcription
	transcriptionPart, err := writer.CreateFormField("transcription")
	if err != nil {
		logger.WithError(err).Error("Failed to create form field for transcription")
		return models.AnalysisResult{}, fmt.Errorf("failed to create form field for transcription string: %w", err)
	}

	_, err = transcriptionPart.Write([]byte(transcription))
	if err != nil {
		logger.WithError(err).Error("Failed to write transcription to form field")
		return models.AnalysisResult{}, fmt.Errorf("failed to write transcription to form field: %w", err)
	}

	// Fetch children data
	children, err := service.childStore.GetAll()
	if err != nil {
		logger.WithError(err).Error("Failed to get all children")
		return models.AnalysisResult{}, fmt.Errorf("failed to get all children: %w", err)
	}

	type ChildData struct {
		ID        int    `json:"child_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	childrenData := make([]ChildData, len(children))
	for i, c := range children {
		childrenData[i] = ChildData{
			ID:        c.ID,
			FirstName: c.FirstName,
			LastName:  c.LastName,
		}
	}

	childrenJSON, err := json.Marshal(childrenData)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal children data")
		return models.AnalysisResult{}, fmt.Errorf("failed to marshal children data: %w", err)
	}

	childrenPart, err := writer.CreateFormField("children_data")
	if err != nil {
		logger.WithError(err).Error("Failed to create form field for children data")
		return models.AnalysisResult{}, fmt.Errorf("failed to create form field for children: %w", err)
	}
	_, err = childrenPart.Write(childrenJSON)
	if err != nil {
		logger.WithError(err).Error("Failed to write children data to form field")
		return models.AnalysisResult{}, fmt.Errorf("failed to write children data to form field: %w", err)
	}

	// Fetch category data
	categories, err := service.categoryStore.GetAll()
	if err != nil {
		logger.WithError(err).Error("Failed to get all categories")
		return models.AnalysisResult{}, fmt.Errorf("failed to get all categories: %w", err)
	}

	type CategoryData struct {
		ID          int    `json:"category_id"`
		Name        string `json:"category_name"`
		Description string `json:"description"`
	}

	categoryData := make([]CategoryData, len(categories))
	for i, c := range categories {
		var description string
		if c.Description != nil {
			description = *c.Description
		}
		categoryData[i] = CategoryData{
			ID:          c.ID,
			Name:        c.Name,
			Description: description,
		}
	}

	categoryJSON, err := json.Marshal(categoryData)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal category data")
		return models.AnalysisResult{}, fmt.Errorf("failed to marshal category data: %w", err)
	}

	categoryPart, err := writer.CreateFormField("category_data")
	if err != nil {
		logger.WithError(err).Error("Failed to create form field for category data")
		return models.AnalysisResult{}, fmt.Errorf("failed to create form field for category: %w", err)
	}
	_, err = categoryPart.Write(categoryJSON)
	if err != nil {
		logger.WithError(err).Error("Failed to write category data to form field")
		return models.AnalysisResult{}, fmt.Errorf("failed to write category data to form field: %w", err)
	}

	// Close the multipart writer.
	err = writer.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close multipart writer")
		return models.AnalysisResult{}, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	logger.Debugf("Request body size: %d", body.Len())

	// Create a new HTTP request to the llm analysis service.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, service.llmAnalysisServiceURL, body)
	if err != nil {
		logger.WithError(err).Error("Failed to create request to llm analysis service")
		return models.AnalysisResult{}, fmt.Errorf("failed to create request to llm analysis: %w", err)
	}

	// Set the content type for the request.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request.
	resp, err := service.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("Failed to send request to llm analysis service")
		return models.AnalysisResult{}, fmt.Errorf("failed to send request to llm analysis service: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.WithError(err).Error("Failed to close response body")
		}
	}()

	// Check the response status code.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		}).Error("Received non-OK response from llm analysis service")
		return models.AnalysisResult{}, fmt.Errorf("llm analysis service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response.
	var analysisResult models.AnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&analysisResult); err != nil {
		logger.WithError(err).Error("Failed to decode response from llm analysis service")
		return models.AnalysisResult{}, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("Successfully received analysis from llm analysis service")
	return analysisResult, nil
}

// Checks if a process entry in the database was created and updates its status.
func (service *AudioAnalysisServiceImpl) UpdateProcessStatus(logger *logrus.Entry, processId int, status string) {
	if processId != -1 {
		if err := service.processStore.Update(&models.Process{ProcessId: processId, Status: status}); err != nil {
			logger.WithError(err).Error("Failed to update process status")
		}
	}
}
