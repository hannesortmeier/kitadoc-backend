package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
	"kitadoc-backend/services/mocks"
)

// TestUserService_RegisterUser tests the RegisterUser method of UserService.
func TestUserService_RegisterUser(t *testing.T) {
	mockStore := new(mocks.MockUserStore)
	testConfig := &config.Config{
		Server: struct {
			Port         int           "mapstructure:\"port\""
			ReadTimeout  time.Duration "mapstructure:\"read_timeout\""
			WriteTimeout time.Duration "mapstructure:\"write_timeout\""
			IdleTimeout  time.Duration "mapstructure:\"idle_timeout\""
			JWTSecret    string        "mapstructure:\"jwt_secret\""
		}{
			JWTSecret: "test_secret",
		},
	}
	userService := services.NewUserService(mockStore, testConfig)
	logger := logrus.NewEntry(logrus.New()) // Create a new logger entry for testing

	// Test case 1: Successful registration
	t.Run("Successful Registration", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "newuser").Return(&models.User{}, data.ErrNotFound).Once()
		mockStore.On("Create", mock.AnythingOfType("*models.User")).Return(1, nil).Once()

		user, err := userService.RegisterUser(logger, "newuser", "password123", "teacher")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "newuser", user.Username)
		mockStore.AssertExpectations(t)
	})

	// Test case 2: User already exists
	t.Run("User Already Exists", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "existinguser").Return(&models.User{ID: 1, Username: "existinguser"}, nil).Once()

		user, err := userService.RegisterUser(logger, "existinguser", "password123", "teacher")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, services.ErrAlreadyExists, err)
		mockStore.AssertExpectations(t)
	})

	// Test case 3: Invalid input (e.g., short password)
	t.Run("Invalid Input", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "invaliduser").Return(&models.User{}, data.ErrNotFound).Once()

		user, err := userService.RegisterUser(logger, "invaliduser", "short", "teacher") // Password too short
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, services.ErrInvalidInput, err)
		mockStore.AssertExpectations(t)
	})
}

// TestUserService_LoginUser tests the LoginUser method of UserService.
func TestUserService_LoginUser(t *testing.T) {
	mockStore := new(mocks.MockUserStore)
	testConfig := &config.Config{
		Server: struct {
			Port         int           "mapstructure:\"port\""
			ReadTimeout  time.Duration "mapstructure:\"read_timeout\""
			WriteTimeout time.Duration "mapstructure:\"write_timeout\""
			IdleTimeout  time.Duration "mapstructure:\"idle_timeout\""
			JWTSecret    string        "mapstructure:\"jwt_secret\""
		}{
			JWTSecret: "test_secret",
		},
	}
	userService := services.NewUserService(mockStore, testConfig)
	logger := logrus.NewEntry(logrus.New())

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
		Role:         "teacher",
	}

	// Test case 1: Successful login
	t.Run("Successful Login", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "testuser").Return(testUser, nil).Once()

		token, err := userService.LoginUser(logger, "testuser", "correctpassword")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockStore.AssertExpectations(t)
	})

	// Test case 2: Invalid password
	t.Run("Invalid Password", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "testuser").Return(testUser, nil).Once()

		token, err := userService.LoginUser(logger, "testuser", "wrongpassword")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, services.ErrInvalidCredentials, err)
		mockStore.AssertExpectations(t)
	})

	// Test case 3: User not found
	t.Run("User Not Found", func(t *testing.T) {
		mockStore.On("GetUserByUsername", "nonexistent").Return(&models.User{}, data.ErrNotFound).Once()

		token, err := userService.LoginUser(logger, "nonexistent", "password")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, services.ErrInvalidCredentials, err)
		mockStore.AssertExpectations(t)
	})
}

// TestUserService_GetUserByID tests the GetUserByID method.
func TestUserService_GetUserByID(t *testing.T) {
	mockStore := new(mocks.MockUserStore)
	testConfig := &config.Config{
		Server: struct {
			Port         int           "mapstructure:\"port\""
			ReadTimeout  time.Duration "mapstructure:\"read_timeout\""
			WriteTimeout time.Duration "mapstructure:\"write_timeout\""
			IdleTimeout  time.Duration "mapstructure:\"idle_timeout\""
			JWTSecret    string        "mapstructure:\"jwt_secret\""
		}{
			JWTSecret: "test_secret",
		},
	}
	userService := services.NewUserService(mockStore, testConfig)
	logger := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	expectedUser := &models.User{ID: 1, Username: "testuser", Role: "teacher"}

	t.Run("User Found", func(t *testing.T) {
		mockStore.On("GetByID", 1).Return(expectedUser, nil).Once()
		user, err := userService.GetUserByID(logger, ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockStore.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockStore.On("GetByID", 2).Return(&models.User{}, data.ErrNotFound).Once()
		user, err := userService.GetUserByID(logger, ctx, 2)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, services.ErrNotFound, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockStore.On("GetByID", 3).Return(&models.User{}, errors.New("db error")).Once()
		user, err := userService.GetUserByID(logger, ctx, 3)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, services.ErrInternal, err)
		mockStore.AssertExpectations(t)
	})
}
