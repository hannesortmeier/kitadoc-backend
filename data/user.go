package data

import (
	"database/sql"
	"errors"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

// UserStore defines the interface for User data operations.
type UserStore interface {
	Create(user *models.User) (int, error)
	GetByID(id int) (*models.User, error)
	Update(user *models.User) error
	Delete(id int) error
	GetUserByUsername(username string) (*models.User, error)
}

// SQLUserStore implements UserStore using database/sql.
type SQLUserStore struct {
	db *sql.DB
}

// NewSQLUserStore creates a new SQLUserStore.
func NewSQLUserStore(db *sql.DB) *SQLUserStore {
	return &SQLUserStore{db: db}
}

// Create inserts a new user into the database.
func (s *SQLUserStore) Create(user *models.User) (int, error) {
	query := `INSERT INTO users (username, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error inserting user: %v", err)
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting last insert ID: %v", err)
		return -1, err
	}
	return int(id), nil
}

// GetByID fetches a user by ID from the database.
func (s *SQLUserStore) GetByID(id int) (*models.User, error) {
	query := `SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE user_id = ?`
	row := s.db.QueryRow(query, id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.GetGlobalLogger().Infof("User with ID %d not found", id)
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update updates an existing user in the database.
func (s *SQLUserStore) Update(user *models.User) error {
	query := `UPDATE users SET username = ?, password_hash = ?, role = ?, updated_at = ? WHERE user_id = ?`
	result, err := s.db.Exec(query, user.Username, user.PasswordHash, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a user by ID from the database.
func (s *SQLUserStore) Delete(id int) error {
	query := `DELETE FROM users WHERE user_id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetUserByUsername fetches a user by username from the database.
func (s *SQLUserStore) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE username = ?`
	row := s.db.QueryRow(query, username)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
