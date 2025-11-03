package services

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user-related business logic operations.
type UserService interface {
	RegisterUser(logger *logrus.Entry, username, password, role string) (*models.User, error)
	LoginUser(logger *logrus.Entry, username, password string) (string, error) // Returns JWT token
	GetCurrentUser(logger *logrus.Entry, tokenString string) (*models.User, error)
	GetUserByID(logger *logrus.Entry, ctx context.Context, id int) (*models.User, error)
	UpdateUser(logger *logrus.Entry, user *models.User) error
	DeleteUser(logger *logrus.Entry, id int) error
	GetAllUsers(logger *logrus.Entry) ([]*models.User, error)
	ChangePassword(logger *logrus.Entry, actor *models.User, userID int, oldPassword, newPassword string) error
}

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	userStore data.UserStore
	validate  *validator.Validate
	config    *config.Config // Add config to service
}

// NewUserService creates a new UserServiceImpl.
func NewUserService(userStore data.UserStore, cfg *config.Config) *UserServiceImpl {
	return &UserServiceImpl{
		userStore: userStore,
		validate:  validator.New(),
		config:    cfg,
	}
}

// RegisterUser registers a new user after hashing the password.
func (s *UserServiceImpl) RegisterUser(logger *logrus.Entry, username, password, role string) (*models.User, error) {
	// Check if user already exists
	_, err := s.userStore.GetUserByUsername(username)
	if err == nil {
		logger.WithField("username", username).Warn("User already exists during registration attempt")
		return nil, ErrAlreadyExists
	}
	if !errors.Is(err, data.ErrNotFound) {
		logger.WithError(err).WithField("username", username).Error("Error checking if user exists during registration")
		return nil, ErrInternal
	}

	// Add password length validation here
	if len(password) < 8 {
		logger.WithField("username", username).Warn("Password too short during registration attempt")
		return nil, ErrInvalidInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithError(err).Error("Error hashing password during registration")
		return nil, ErrInternal
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := models.ValidateUser(*user); err != nil {
		logger.WithError(err).Warn("Invalid user data provided during registration")
		return nil, ErrInvalidInput
	}

	id, err := s.userStore.Create(user)
	if err != nil {
		logger.WithError(err).Error("Error creating user in store")
		return nil, ErrInternal
	}
	user.ID = id
	logger.WithField("user_id", user.ID).Info("User registered successfully")
	return user, nil
}

// LoginUser authenticates a user and generates a JWT token.
func (s *UserServiceImpl) LoginUser(logger *logrus.Entry, username, password string) (string, error) {
	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("username", username).Warn("Login attempt with invalid credentials: user not found")
			return "", ErrInvalidCredentials
		}
		logger.WithError(err).WithField("username", username).Error("Error fetching user by username during login")
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		logger.WithField("username", username).Warn("Login attempt with invalid credentials: password mismatch")
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.Server.JWTSecret)) // Use JWTSecret from config
	if err != nil {
		logger.WithError(err).Error("Error signing JWT token")
		return "", ErrInternal
	}
	logger.WithField("user_id", user.ID).Info("User logged in successfully, JWT generated")
	return tokenString, nil
}

// GetCurrentUser parses a JWT token and returns the corresponding user.
func (s *UserServiceImpl) GetCurrentUser(logger *logrus.Entry, tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.WithField("signing_method", token.Method).Warn("Unexpected signing method for JWT")
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Server.JWTSecret), nil // Use JWTSecret from config
	})

	if err != nil {
		logger.WithError(err).Warn("Error parsing JWT token")
		return nil, ErrAuthenticationFailed
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Warn("Invalid or expired JWT token claims")
		return nil, ErrAuthenticationFailed
	}

	userID := int(claims["user_id"].(float64))
	username := claims["username"].(string)
	role := claims["role"].(string)

	user := &models.User{
		ID:       userID,
		Username: username,
		Role:     role,
	}
	logger.WithField("user_id", user.ID).Debug("Current user fetched from JWT")
	return user, nil
}

// UpdateUser updates an existing user.
func (s *UserServiceImpl) UpdateUser(logger *logrus.Entry, user *models.User) error {
	if err := models.ValidateUser(*user); err != nil {
		logger.WithError(err).Warn("Invalid input for UpdateUser")
		return ErrInvalidInput
	}

	// Fetch existing user to preserve password hash if not updated
	existingUser, err := s.userStore.GetByID(user.ID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", user.ID).Warn("User not found for update")
			return ErrNotFound
		}
		logger.WithError(err).WithField("user_id", user.ID).Error("Error fetching existing user for update")
		return ErrInternal
	}

	// If password is provided, hash it. Otherwise, keep the existing hash.
	if user.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			logger.WithError(err).Error("Error hashing new password during user update")
			return ErrInternal
		}
		user.PasswordHash = string(hashedPassword)
	} else {
		user.PasswordHash = existingUser.PasswordHash
	}

	user.UpdatedAt = time.Now()
	err = s.userStore.Update(user)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", user.ID).Warn("User not found during update in store")
			return ErrNotFound
		}
		logger.WithError(err).WithField("user_id", user.ID).Error("Error updating user in store")
		return ErrInternal
	}
	logger.WithField("user_id", user.ID).Info("User updated successfully")
	return nil
}

// DeleteUser deletes a user by ID.
func (s *UserServiceImpl) DeleteUser(logger *logrus.Entry, id int) error {
	err := s.userStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", id).Warn("User not found for deletion")
			return ErrNotFound
		}
		logger.WithError(err).WithField("user_id", id).Error("Error deleting user from store")
		return ErrInternal
	}
	logger.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// GetUserByID fetches a user by ID.
func (s *UserServiceImpl) GetUserByID(logger *logrus.Entry, ctx context.Context, id int) (*models.User, error) {
	user, err := s.userStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", id).Warn("User not found by ID")
			return nil, ErrNotFound
		}
		logger.WithError(err).WithField("user_id", id).Error("Error fetching user by ID from store")
		return nil, ErrInternal
	}
	logger.WithField("user_id", id).Debug("User fetched by ID successfully")
	return user, nil
}

// GetAllUsers fetches all users.
func (s *UserServiceImpl) GetAllUsers(logger *logrus.Entry) ([]*models.User, error) {
	users, err := s.userStore.GetAll()
	if err != nil {
		logger.WithError(err).Error("Error fetching all users from store")
		return nil, ErrInternal
	}
	logger.Info("All users fetched successfully")
	return users, nil
}

// ChangePassword changes a user's password.
func (s *UserServiceImpl) ChangePassword(logger *logrus.Entry, actor *models.User, userID int, oldPassword, newPassword string) error {
	user, err := s.userStore.GetByID(userID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", userID).Warn("User not found for password change")
			return ErrNotFound
		}
		logger.WithError(err).WithField("user_id", userID).Error("Error fetching user for password change")
		return ErrInternal
	}

	// Admin can change any user's password without the old password
	if actor.Role == string(data.RoleAdmin) {
		logger.WithField("admin_id", actor.ID).WithField("user_id", userID).Info("Admin changing user's password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.WithError(err).Error("Error hashing new password")
			return ErrInternal
		}

		err = s.userStore.UpdatePassword(userID, string(hashedPassword))
		if err != nil {
			logger.WithError(err).WithField("user_id", userID).Error("Error updating password in store")
			return ErrInternal
		}
		logger.WithField("user_id", userID).Info("Password changed successfully by admin")
		return nil
	}

	// Regular user can only change their own password
	if actor.ID != userID {
		logger.WithFields(logrus.Fields{
			"actor_id": actor.ID,
			"user_id":  userID,
		}).Warn("Permission denied to change another user's password")
		return ErrPermissionDenied
	}
	logger.WithField("user_id", userID).Info("User changing own password")

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		logger.WithField("user_id", userID).Warn("Invalid old password provided")
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.WithError(err).Error("Error hashing new password")
		return ErrInternal
	}

	err = s.userStore.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Error updating password in store")
		return ErrInternal
	}

	logger.WithField("user_id", userID).Info("Password changed successfully")
	return nil
}
