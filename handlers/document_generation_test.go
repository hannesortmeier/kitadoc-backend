package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/internal/testutils"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDocumentGenerationHandler(t *testing.T) {
	mockService := new(mocks.MockDocumentationEntryService)
	handler := NewDocumentGenerationHandler(mockService)
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.DocumentationEntryService)
}

func TestGenerateChildReport(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("Successful Report Generation", func(t *testing.T) {
		mockService := new(mocks.MockDocumentationEntryService)
		mockService.On("GenerateChildReport", mock.Anything, mock.Anything, 123).Return([]byte("test report content"), nil)

		handler := NewDocumentGenerationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/child-report/123", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("child_id", "123")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GenerateChildReport(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "test report content", recorder.Body.String())
		assert.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", recorder.Header().Get("Content-Type"))
		assert.Equal(t, "attachment; filename=\"child_report.docx\"", recorder.Header().Get("Content-Disposition"))

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Child ID", func(t *testing.T) {
		mockService := new(mocks.MockDocumentationEntryService)
		handler := NewDocumentGenerationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reports/abc", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("child_id", "abc")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GenerateChildReport(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "Invalid child ID\n", recorder.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns ErrChildReportGenerationFailed", func(t *testing.T) {
		mockService := new(mocks.MockDocumentationEntryService)
		mockService.On("GenerateChildReport", mock.Anything, mock.Anything, 123).Return(nil, services.ErrChildReportGenerationFailed)

		handler := NewDocumentGenerationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reports/123", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("child_id", "123")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GenerateChildReport(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Failed to generate child report\n", recorder.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Other Error", func(t *testing.T) {
		mockService := new(mocks.MockDocumentationEntryService)
		mockService.On("GenerateChildReport", mock.Anything, mock.Anything, 123).Return(nil, errors.New("some other service error"))

		handler := NewDocumentGenerationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reports/123", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("child_id", "123")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GenerateChildReport(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())
		mockService.AssertExpectations(t)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		mockService := new(mocks.MockDocumentationEntryService)
		mockService.On("GenerateChildReport", mock.Anything, mock.Anything, 123).Return(nil, context.Canceled)

		handler := NewDocumentGenerationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reports/123", nil)
		ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
		req.SetPathValue("child_id", "123")
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()
		handler.GenerateChildReport(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "Internal server error\n", recorder.Body.String())
		mockService.AssertExpectations(t)
	})
}
