package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/internal/testutils"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDocumentationEntryHandler(t *testing.T) {
	mockService := new(mocks.MockDocumentationEntryService)
	handler := NewDocumentationEntryHandler(mockService)
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.DocumentationEntryService)
}

func TestCreateDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name               string
		inputPayload       interface{}
		mockServiceSetup   func(*mocks.MockDocumentationEntryService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Successful Creation",
			inputPayload: models.DocumentationEntry{
				ChildID:              1,
				TeacherID:            1,
				CategoryID:           1,
				ObservationDate:      time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
				ObservationDescription: "Test observation",
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("CreateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(&models.DocumentationEntry{
					ID:                   1,
					ChildID:              1,
					TeacherID:            1,
					CategoryID:           1,
					ObservationDate:      time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
					ObservationDescription: "Test observation",
					CreatedAt:            time.Now(),
					UpdatedAt:            time.Now(),
				}, nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
			expectedBody:       `{"id":1,"child_id":1,"teacher_id":1,"category_id":1,"observation_date":"2023-01-15T00:00:00Z","observation_description":"Test observation","is_approved":false,"approved_by_user_id":null,"created_at":"%s","updated_at":"%s"}`,
		},
		{
			name:               "Invalid JSON Payload",
			inputPayload:       `{"child_id": "invalid"}`,
			mockServiceSetup:   func(m *mocks.MockDocumentationEntryService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid request payload\n",
		},
		{
			name: "Service Returns ErrInvalidInput",
			inputPayload: models.DocumentationEntry{
				ChildID: 1,
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("CreateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(nil, services.ErrInvalidInput).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid documentation entry data provided\n",
		},
		{
			name: "Service Returns Other Error",
			inputPayload: models.DocumentationEntry{
				ChildID: 1,
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("CreateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(nil, errors.New("database error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockDocumentationEntryService)
			tt.mockServiceSetup(mockService)

			handler := NewDocumentationEntryHandler(mockService)

			var reqBody bytes.Buffer
			if tt.inputPayload != nil {
				json.NewEncoder(&reqBody).Encode(tt.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPost, "/entries", &reqBody)
			req = req.WithContext(context.WithValue(req.Context(), testutils.ContextKeyLogger, logger))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler.CreateDocumentationEntry(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			if tt.expectedStatusCode == http.StatusCreated {
				var actualEntry models.DocumentationEntry
				err := json.Unmarshal(recorder.Body.Bytes(), &actualEntry)
				assert.NoError(t, err)
				assert.Equal(t, 1, actualEntry.ID)
				assert.Equal(t, 1, actualEntry.ChildID)
				assert.Equal(t, 1, actualEntry.TeacherID)
				assert.Equal(t, 1, actualEntry.CategoryID)
				assert.Equal(t, time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC), actualEntry.ObservationDate)
				assert.Equal(t, "Test observation", actualEntry.ObservationDescription)
				assert.False(t, actualEntry.IsApproved)
				assert.Nil(t, actualEntry.ApprovedByUserID)
				assert.WithinDuration(t, time.Now(), actualEntry.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, time.Now(), actualEntry.UpdatedAt, 5*time.Second)
			} else {
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetDocumentationEntriesByChildID(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name               string
		childIDParam       string
		mockServiceSetup   func(*mocks.MockDocumentationEntryService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:         "Successful Fetch",
			childIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("GetAllDocumentationForChild", mock.Anything, mock.Anything, 1).Return([]models.DocumentationEntry{
					{ID: 1, ChildID: 1, ObservationDescription: "Entry 1"},
					{ID: 2, ChildID: 1, ObservationDescription: "Entry 2"},
				}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":1,"child_id":1,"teacher_id":0,"category_id":0,"observation_date":"0001-01-01T00:00:00Z","observation_description":"Entry 1","is_approved":false,"approved_by_user_id":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":2,"child_id":1,"teacher_id":0,"category_id":0,"observation_date":"0001-01-01T00:00:00Z","observation_description":"Entry 2","is_approved":false,"approved_by_user_id":null,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]` + "\n",
		},
		{
			name:         "Invalid Child ID",
			childIDParam: "abc",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				// No service call expected
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid child ID\n",
		},
		{
			name:         "Service Returns Error",
			childIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("GetAllDocumentationForChild", mock.Anything, mock.Anything, 1).Return(nil, errors.New("service error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockDocumentationEntryService)
			tt.mockServiceSetup(mockService)

			handler := NewDocumentationEntryHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/entries/child/"+tt.childIDParam, nil)
			ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
			req.SetPathValue("child_id", tt.childIDParam)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			handler.GetDocumentationEntriesByChildID(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name               string
		entryIDParam       string
		inputPayload       interface{}
		mockServiceSetup   func(*mocks.MockDocumentationEntryService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:         "Successful Update",
			entryIDParam: "1",
			inputPayload: models.DocumentationEntry{
				ID:                   1,
				ChildID:              1,
				TeacherID:            1,
				CategoryID:           1,
				ObservationDate:      time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
				ObservationDescription: "Updated observation",
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("UpdateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Documentation entry updated successfully"}` + "\n",
		},
		{
			name:               "Invalid Entry ID",
			entryIDParam:       "abc",
			inputPayload:       models.DocumentationEntry{},
			mockServiceSetup:   func(m *mocks.MockDocumentationEntryService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid entry ID\n",
		},
		{
			name:               "Invalid JSON Payload",
			entryIDParam:       "1",
			inputPayload:       `{"child_id": "invalid"}`,
			mockServiceSetup:   func(m *mocks.MockDocumentationEntryService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid request payload\n",
		},
		{
			name:         "Service Returns ErrNotFound",
			entryIDParam: "99",
			inputPayload: models.DocumentationEntry{
				ID: 99, ChildID: 1, TeacherID: 1, CategoryID: 1, ObservationDate: time.Now(), ObservationDescription: "Test",
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("UpdateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(services.ErrNotFound).Once()
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Documentation entry not found\n",
		},
		{
			name:         "Service Returns ErrInvalidInput",
			entryIDParam: "1",
			inputPayload: models.DocumentationEntry{
				ID: 1, ChildID: 1, TeacherID: 1, CategoryID: 1, ObservationDate: time.Now(), ObservationDescription: "Test",
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("UpdateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(services.ErrInvalidInput).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid documentation entry data provided\n",
		},
		{
			name:         "Service Returns Other Error",
			entryIDParam: "1",
			inputPayload: models.DocumentationEntry{
				ID: 1, ChildID: 1, TeacherID: 1, CategoryID: 1, ObservationDate: time.Now(), ObservationDescription: "Test",
			},
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("UpdateDocumentationEntry", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DocumentationEntry")).Return(errors.New("database error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockDocumentationEntryService)
			tt.mockServiceSetup(mockService)

			handler := NewDocumentationEntryHandler(mockService)

			var reqBody bytes.Buffer
			if tt.inputPayload != nil {
				json.NewEncoder(&reqBody).Encode(tt.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPut, "/entries/"+tt.entryIDParam, &reqBody)
			ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
			req.SetPathValue("entry_id", tt.entryIDParam)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler.UpdateDocumentationEntry(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name               string
		entryIDParam       string
		mockServiceSetup   func(*mocks.MockDocumentationEntryService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:         "Successful Deletion",
			entryIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("DeleteDocumentationEntry", mock.Anything, mock.Anything, 1).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Documentation entry deleted successfully"}` + "\n",
		},
		{
			name:         "Invalid Entry ID",
			entryIDParam: "abc",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				// No service call expected
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid entry ID\n",
		},
		{
			name:         "Service Returns ErrNotFound",
			entryIDParam: "99",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("DeleteDocumentationEntry", mock.Anything, mock.Anything, 99).Return(services.ErrNotFound).Once()
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Documentation entry not found\n",
		},
		{
			name:         "Service Returns Other Error",
			entryIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("DeleteDocumentationEntry", mock.Anything, mock.Anything, 1).Return(errors.New("database error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockDocumentationEntryService)
			tt.mockServiceSetup(mockService)

			handler := NewDocumentationEntryHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/entries/"+tt.entryIDParam, nil)
			ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
			req.SetPathValue("entry_id", tt.entryIDParam)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			handler.DeleteDocumentationEntry(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestApproveDocumentationEntry(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name               string
		entryIDParam       string
		mockServiceSetup   func(*mocks.MockDocumentationEntryService)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:         "Successful Approval",
			entryIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("ApproveDocumentationEntry", mock.Anything, mock.Anything, 1, 1).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Documentation entry approved successfully"}` + "\n",
		},
		{
			name:         "Invalid Entry ID",
			entryIDParam: "abc",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				// No service call expected
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid entry ID\n",
		},
		{
			name:         "Service Returns ErrNotFound",
			entryIDParam: "99",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("ApproveDocumentationEntry", mock.Anything, mock.Anything, 99, 1).Return(services.ErrNotFound).Once()
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Documentation entry not found\n",
		},
		{
			name:         "Service Returns Other Error",
			entryIDParam: "1",
			mockServiceSetup: func(m *mocks.MockDocumentationEntryService) {
				m.On("ApproveDocumentationEntry", mock.Anything, mock.Anything, 1, 1).Return(errors.New("service error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockDocumentationEntryService)
			tt.mockServiceSetup(mockService)

			handler := NewDocumentationEntryHandler(mockService)

			req := httptest.NewRequest(http.MethodPost, "/entries/"+tt.entryIDParam+"/approve", nil)
			ctx := context.WithValue(req.Context(), testutils.ContextKeyLogger, logger)
			req.SetPathValue("entry_id", tt.entryIDParam)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			handler.ApproveDocumentationEntry(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}