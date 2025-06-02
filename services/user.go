package services

import (
	"errors"
	"log"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret is the secret key for JWT token generation and validation.
// In a real application, this should be loaded from environment variables or a secure configuration.
const JWTSecret = "supersecretjwtkeythatshouldbeverylongandcomplex"

// UserService defines the interface for user-related business logic operations.
type UserService interface {
	RegisterUser(username, password, role string) (*models.User, error)
	LoginUser(username, password string) (string, error) // Returns JWT token
	GetCurrentUser(tokenString string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
}

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	userStore data.UserStore
	validate  *validator.Validate
}

// NewUserService creates a new UserServiceImpl.
func NewUserService(userStore data.UserStore) *UserServiceImpl {
	return &UserServiceImpl{
		userStore: userStore,
		validate:  validator.New(),
	}
}

// RegisterUser registers a new user after hashing the password.
func (s *UserServiceImpl) RegisterUser(username, password, role string) (*models.User, error) {
	// Check if user already exists
	_, err := s.userStore.GetUserByUsername(username)
	if err == nil {
		return nil, ErrAlreadyExists
	}
	if !errors.Is(err, data.ErrNotFound) {
		log.Printf("Error checking if user exists: %v", err)
		return nil, ErrInternal
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
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
		return nil, ErrInvalidInput
	}

	id, err := s.userStore.Create(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, ErrInternal
	}
	user.ID = id
	return user, nil
}

// LoginUser authenticates a user and generates a JWT token.
func (s *UserServiceImpl) LoginUser(username, password string) (string, error) {
	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})

	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", ErrInternal
	}

	return tokenString, nil
}

// GetCurrentUser parses a JWT token and returns the corresponding user.
func (s *UserServiceImpl) GetCurrentUser(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
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

	return user, nil
}

// UpdateUser updates an existing user.
func (s *UserServiceImpl) UpdateUser(user *models.User) error {
	if err := models.ValidateUser(*user); err != nil {
		return ErrInvalidInput
	}

	// Fetch existing user to preserve password hash if not updated
	existingUser, err := s.userStore.GetByID(user.ID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	// If password is provided, hash it. Otherwise, keep the existing hash.
	if user.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
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
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}

// DeleteUser deletes a user by ID.
func (s *UserServiceImpl) DeleteUser(id int) error {
	err := s.userStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}