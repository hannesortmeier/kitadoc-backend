package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	UserService services.UserService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService services.UserService) *AuthHandler {
	return &AuthHandler{UserService: userService}
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // e.g., "teacher" or "admin"
}

// Login handles user login.
func (authHandler *AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	var req LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("Invalid request payload for Login")
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token, err := authHandler.UserService.LoginUser(logger, req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			logger.WithField("username", req.Username).Warn("Invalid credentials during login attempt")
			http.Error(writer, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		logger.WithError(err).Error("Internal server error during login")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(map[string]string{"token": token}); err != nil {
		logger.WithError(err).Error("Failed to encode login response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Logout handles user logout (token invalidation is typically client-side).
func (authHandler *AuthHandler) Logout(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	// For JWT, logout is typically handled client-side by discarding the token.
	// If server-side invalidation is needed, a token blacklist mechanism would be implemented.
	logger.Info("User logged out (client-side token discard)")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Logged out successfully"}); err != nil {
		logger.WithError(err).Error("Failed to encode logout response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetMe returns the currently authenticated user's information.
func (authHandler *AuthHandler) GetMe(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	user, ok := request.Context().Value(middleware.ContextKeyUser).(*models.User)
	if !ok {
		logger.Error("User not found in context for GetMe handler")
		http.Error(writer, "User not found in context", http.StatusInternalServerError)
		return
	}
	logger.WithField("user_id", user.ID).Info("Fetched current user information")

	if err := json.NewEncoder(writer).Encode(user); err != nil {
		logger.WithError(err).Error("Failed to encode user information response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RegisterUser handles user registration.
func (authHandler *AuthHandler) RegisterUser(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	var user RegisterUserRequest
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		logger.WithError(err).Warn("Invalid request payload for RegisterUser")
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdUser, err := authHandler.UserService.RegisterUser(logger, user.Username, user.Password, user.Role)
	if err != nil {
		if err == services.ErrAlreadyExists {
			logger.WithField("username", user.Username).Warn("Registration attempt for existing username")
			http.Error(writer, "User with this username already exists", http.StatusConflict)
			return
		}
		if err == services.ErrInvalidInput {
			logger.WithError(err).Warn("Invalid user data provided for registration")
			http.Error(writer, "Invalid user data provided", http.StatusBadRequest)
			return
		}
		logger.WithError(err).Error("Internal server error during user registration")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(writer).Encode(createdUser); err != nil {
		logger.WithError(err).Error("Failed to encode user registration response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateUser handles updating user information.
func (authHandler *AuthHandler) UpdateUser(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	// This handler would typically require an ID from the URL path,
	// but for simplicity, we'll assume the user can only update their own profile
	// or an admin can update any user.
	// For now, we'll just use the user from the context.
	userFromContext, ok := request.Context().Value(middleware.ContextKeyUser).(*models.User)
	if !ok {
		logger.Error("User not found in context for UpdateUser handler")
		http.Error(writer, "User not found in context", http.StatusInternalServerError)
		return
	}

	var updatedUser models.User
	if err := json.NewDecoder(request.Body).Decode(&updatedUser); err != nil {
		logger.WithError(err).Warn("Invalid request payload for UpdateUser")
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Ensure the ID from the request body matches the authenticated user's ID
	// or if an admin is performing the update.
	// For now, we'll assume the user is updating their own profile.
	updatedUser.ID = userFromContext.ID
	updatedUser.UpdatedAt = time.Now()

	err := authHandler.UserService.UpdateUser(logger, &updatedUser)
	if err != nil {
		if err == services.ErrNotFound {
			logger.WithField("user_id", updatedUser.ID).Warn("User not found for update")
			http.Error(writer, "User not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			logger.WithError(err).Warn("Invalid user data provided for update")
			http.Error(writer, "Invalid user data provided", http.StatusBadRequest)
			return
		}
		logger.WithError(err).WithField("user_id", updatedUser.ID).Error("Internal server error during user update")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.WithField("user_id", updatedUser.ID).Info("User updated successfully")

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "User updated successfully"}); err != nil {
		logger.WithError(err).Error("Failed to encode user update response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteUser handles deleting a user.
func (authHandler *AuthHandler) DeleteUser(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	// This handler would typically require an ID from the URL path.
	// For simplicity, we'll assume the user can only delete their own profile
	// or an admin can delete any user.
	// For now, we'll just use the user from the context.
	userFromContext, ok := request.Context().Value(middleware.ContextKeyUser).(*models.User)
	if !ok {
		logger.Error("User not found in context for DeleteUser handler")
		http.Error(writer, "User not found in context", http.StatusInternalServerError)
		return
	}

	err := authHandler.UserService.DeleteUser(logger, userFromContext.ID)
	if err != nil {
		if err == services.ErrNotFound {
			logger.WithField("user_id", userFromContext.ID).Warn("User not found for deletion")
			http.Error(writer, "User not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithField("user_id", userFromContext.ID).Error("Internal server error during user deletion")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}
	logger.WithField("user_id", userFromContext.ID).Info("User deleted successfully")

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "User deleted successfully"}); err != nil {
		logger.WithError(err).Error("Failed to encode user deletion response")
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
