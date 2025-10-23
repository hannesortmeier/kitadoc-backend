package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"kitadoc-backend/data"
	"kitadoc-backend/middleware"
	"kitadoc-backend/models"

	"github.com/sirupsen/logrus"
)

// AudioAnalysisService defines the interface for audio analysis operations.
type AudioAnalysisService interface {
	AnalyzeAudio(ctx context.Context, fileContent []byte, filename string) (models.AnalysisResult, error)
}

// AudioAnalysisServiceImpl implements AudioAnalysisService.
type AudioAnalysisServiceImpl struct {
	httpClient    *http.Client
	audioProcURL  string
	childStore    data.ChildStore
	categoryStore data.CategoryStore
}

// NewAudioAnalysisService creates a new AudioAnalysisServiceImpl.
func NewAudioAnalysisService(httpClient *http.Client, audioProcURL string, childStore data.ChildStore, categoryStore data.CategoryStore) *AudioAnalysisServiceImpl {
	return &AudioAnalysisServiceImpl{
		httpClient:    httpClient,
		audioProcURL:  audioProcURL,
		childStore:    childStore,
		categoryStore: categoryStore,
	}
}

// AnalyzeAudio forwards the audio file to the audio-proc service for analysis.
func (s *AudioAnalysisServiceImpl) AnalyzeAudio(ctx context.Context, fileContent []byte, filename string) (models.AnalysisResult, error) {
	logger := middleware.GetLoggerWithReqID(ctx)

	// Create a new multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form-data header with the provided filename.
	audioPart, err := writer.CreateFormFile("audio_file", filename)
	if err != nil {
		logger.WithError(err).Error("Failed to create form file")
		return models.AnalysisResult{}, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the file content to the form file.
	_, err = io.Copy(audioPart, bytes.NewReader(fileContent))
	if err != nil {
		logger.WithError(err).Error("Failed to copy file content to form file")
		return models.AnalysisResult{}, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Fetch children data
	children, err := s.childStore.GetAll()
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
		return models.AnalysisResult{}, fmt.Errorf("failed to create form field: %w", err)
	}
	_, err = childrenPart.Write(childrenJSON)
	if err != nil {
		logger.WithError(err).Error("Failed to write children data to form field")
		return models.AnalysisResult{}, fmt.Errorf("failed to write children data: %w", err)
	}

	// Fetch category data
	categories, err := s.categoryStore.GetAll()
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
		return models.AnalysisResult{}, fmt.Errorf("failed to create form field: %w", err)
	}
	_, err = categoryPart.Write(categoryJSON)
	if err != nil {
		logger.WithError(err).Error("Failed to write category data to form field")
		return models.AnalysisResult{}, fmt.Errorf("failed to write category data: %w", err)
	}

	// Close the multipart writer.
	err = writer.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close multipart writer")
		return models.AnalysisResult{}, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	logger.Debugf("Request body size: %d", body.Len())

	// Create a new HTTP request to the audio-proc service.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.audioProcURL, body)
	if err != nil {
		logger.WithError(err).Error("Failed to create request to audio-proc service")
		return models.AnalysisResult{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the content type for the request.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request.
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("Failed to send request to audio-proc service")
		return models.AnalysisResult{}, fmt.Errorf("failed to send request to audio-proc service: %w", err)
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
		}).Error("Received non-OK response from audio-proc service")
		return models.AnalysisResult{}, fmt.Errorf("audio-proc service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response.
	var analysisResult models.AnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&analysisResult); err != nil {
		logger.WithError(err).Error("Failed to decode response from audio-proc service")
		return models.AnalysisResult{}, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("Successfully received analysis from audio-proc service")
	return analysisResult, nil
}
