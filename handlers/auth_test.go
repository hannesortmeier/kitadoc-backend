package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"kitadoc-backend/handlers/mocks"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	// Initialize a discard logger for tests to prevent nil pointer dereferences
	logger.InitGlobalLogger(logrus.DebugLevel, &logrus.TextFormatter{DisableColors: true})
	os.Exit(m.Run())
}

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		reqBody := LoginRequest{Username: "testuser", Password: "password123"}
		mockService.On("LoginUser", mock.Anything, reqBody.Username, reqBody.Password).Return("mock_token", nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]string
		json.NewDecoder(rr.Body).Decode(&response) //nolint:errcheck
		assert.Equal(t, "mock_token", response["token"])
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockService.AssertNotCalled(t, "LoginUser", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		reqBody := LoginRequest{Username: "testuser", Password: "wrongpassword"}
		mockService.On("LoginUser", mock.Anything, reqBody.Username, reqBody.Password).Return("", services.ErrInvalidCredentials).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid username or password")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		reqBody := LoginRequest{Username: "testuser", Password: "password123"}
		mockService.On("LoginUser", mock.Anything, reqBody.Username, reqBody.Password).Return("", errors.New("db error")).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestLogout(t *testing.T) {
	mockService := new(mocks.UserService)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.NewDecoder(rr.Body).Decode(&response) //nolint:errcheck
	assert.Equal(t, "Logged out successfully", response["message"])
	mockService.AssertNotCalled(t, "LoginUser", mock.Anything, mock.Anything, mock.Anything) // Ensure no service calls
}

func TestGetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		user := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, user)
		req := httptest.NewRequest(http.MethodGet, "/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.GetMe(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var fetchedUser models.User
		json.NewDecoder(rr.Body).Decode(&fetchedUser) //nolint:errcheck
		assert.Equal(t, user.ID, fetchedUser.ID)
		assert.Equal(t, user.Username, fetchedUser.Username)
		mockService.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
	})

	t.Run("user not found in context", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/me", nil) // No user in context
		rr := httptest.NewRecorder()

		handler.GetMe(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found in context")
	})
}

func TestRegisterUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userRequest := RegisterUserRequest{Username: "newuser", Password: "password123", Role: "teacher"}
		expectedUser := models.User{
			Username:     userRequest.Username,
			PasswordHash: "hashedpassword", // This would be the hashed password in a real scenario
			Role:         userRequest.Role,
		}
		mockService.On("RegisterUser", mock.Anything, userRequest.Username, userRequest.Password, userRequest.Role).Return(&expectedUser, nil).Once()

		body, _ := json.Marshal(userRequest)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var createdUser models.User
		json.NewDecoder(rr.Body).Decode(&createdUser) //nolint:errcheck
		assert.Equal(t, expectedUser.Username, expectedUser.Username)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockService.AssertNotCalled(t, "RegisterUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userRequest := RegisterUserRequest{Username: "existinguser", Password: "password123", Role: "teacher"}
		mockService.On("RegisterUser", mock.Anything, userRequest.Username, userRequest.Password, userRequest.Role).Return(nil, services.ErrAlreadyExists).Once()

		body, _ := json.Marshal(userRequest)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "User with this username already exists")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user data provided", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userRequest := RegisterUserRequest{Username: "invalid", Password: "", Role: "teacher"}
		mockService.On("RegisterUser", mock.Anything, userRequest.Username, userRequest.Password, userRequest.Role).Return(nil, services.ErrInvalidInput).Once()

		body, _ := json.Marshal(userRequest)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid user data provided")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userRequest := RegisterUserRequest{Username: "invalid", Password: "", Role: "teacher"}
		mockService.On("RegisterUser", mock.Anything, userRequest.Username, userRequest.Password, userRequest.Role).Return(nil, errors.New("db error")).Once()

		body, _ := json.Marshal(userRequest)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		updatedUser := models.User{ID: 1, Username: "updateduser", Role: "teacher"}
		mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		body, _ := json.Marshal(updatedUser)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "User updated successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("user not found in context", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		updatedUser := models.User{ID: 1, Username: "updateduser", Role: "teacher"}
		body, _ := json.Marshal(updatedUser)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body)) // No user in context
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found in context")
		mockService.AssertNotCalled(t, "UpdateUser", mock.Anything, mock.Anything)
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer([]byte("invalid json"))).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
		mockService.AssertNotCalled(t, "UpdateUser", mock.Anything, mock.Anything)
	})

	t.Run("user not found in service", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		updatedUser := models.User{ID: 1, Username: "updateduser", Role: "teacher"}
		mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(services.ErrNotFound).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		body, _ := json.Marshal(updatedUser)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found")
		mockService.AssertExpectations(t)
	})

	t.Run("invalid user data provided", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		updatedUser := models.User{ID: 1, Username: "invalid", Role: "invalid_role"} // Invalid role
		mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(services.ErrInvalidInput).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		body, _ := json.Marshal(updatedUser)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid user data provided")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		updatedUser := models.User{ID: 1, Username: "updateduser", Role: "teacher"}
		mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("db error")).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		body, _ := json.Marshal(updatedUser)
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.UpdateUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		mockService.On("DeleteUser", mock.Anything, userInContext.ID).Return(nil).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.DeleteUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "User deleted successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("user not found in context", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil) // No user in context
		rr := httptest.NewRecorder()

		handler.DeleteUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found in context")
		mockService.AssertNotCalled(t, "DeleteUser", mock.Anything, mock.Anything)
	})

	t.Run("user not found in service", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		mockService.On("DeleteUser", mock.Anything, userInContext.ID).Return(services.ErrNotFound).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.DeleteUser(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found")
		mockService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService := new(mocks.UserService)
		handler := NewAuthHandler(mockService)

		userInContext := &models.User{ID: 1, Username: "testuser", Role: "teacher"}
		mockService.On("DeleteUser", mock.Anything, userInContext.ID).Return(errors.New("db error")).Once()

		ctx := context.WithValue(context.Background(), middleware.ContextKeyUser, userInContext)
		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.DeleteUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
		mockService.AssertExpectations(t)
	})
}
