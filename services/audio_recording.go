package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

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
	httpClient   *http.Client
	audioProcURL string
}

// NewAudioAnalysisService creates a new AudioAnalysisServiceImpl.
func NewAudioAnalysisService(httpClient *http.Client, audioProcURL string) *AudioAnalysisServiceImpl {
	return &AudioAnalysisServiceImpl{
		httpClient:   httpClient,
		audioProcURL: audioProcURL,
	}
}

// AnalyzeAudio forwards the audio file to the audio-proc service for analysis.
func (s *AudioAnalysisServiceImpl) AnalyzeAudio(ctx context.Context, fileContent []byte, filename string) (models.AnalysisResult, error) {
	logger := middleware.GetLoggerWithReqID(ctx)

	// Create a new multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form-data header with the provided filename.
	part, err := writer.CreateFormFile("audio_file", filename)
	if err != nil {
		logger.WithError(err).Error("Failed to create form file")
		return models.AnalysisResult{}, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the file content to the form file.
	_, err = io.Copy(part, bytes.NewReader(fileContent))
	if err != nil {
		logger.WithError(err).Error("Failed to copy file content to form file")
		return models.AnalysisResult{}, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the multipart writer.
	err = writer.Close()
	if err != nil {
		logger.WithError(err).Error("Failed to close multipart writer")
		return models.AnalysisResult{}, fmt.Errorf("failed to close multipart writer: %w", err)
	}

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
	logger.Debug("Analysis result: ", analysisResult)
	return analysisResult, nil
}
