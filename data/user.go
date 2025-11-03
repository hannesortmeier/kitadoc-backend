package data

import (
	"database/sql"
	"errors"
	"fmt"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
	"reflect"
)

// UserStore defines the interface for User data operations.
type UserStore interface {
	Create(user *models.User) (int, error)
	GetByID(id int) (*models.User, error)
	Update(user *models.User) error
	Delete(id int) error
	GetUserByUsername(username string) (*models.User, error)
	GetAll() ([]*models.User, error)
	UpdatePassword(id int, passwordHash string) error
}

// SQLUserStore implements UserStore using database/sql.
type SQLUserStore struct {
	db            *sql.DB
	encryptionKey []byte
}

// NewSQLUserStore creates a new SQLUserStore.
func NewSQLUserStore(db *sql.DB, encryptionKey []byte) *SQLUserStore {
	return &SQLUserStore{db: db, encryptionKey: encryptionKey}
}

// toUserDB converts a models.User to a models.UserDB and encrypts PII fields.
func toUserDB(user *models.User, key []byte) (*models.UserDB, error) {
	dbUser := &models.UserDB{}

	userVal := reflect.ValueOf(user).Elem()
	dbUserVal := reflect.ValueOf(dbUser).Elem()

	for i := 0; i < userVal.NumField(); i++ {
		userField := userVal.Field(i)
		userTypeField := userVal.Type().Field(i)
		dbField := dbUserVal.FieldByName(userTypeField.Name)

		if !dbField.IsValid() || !dbField.CanSet() {
			continue
		}

		if tag := userTypeField.Tag.Get("pii"); tag == "true" {
			encrypted, err := Encrypt(userField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", userTypeField.Name, err)
			}
			dbField.SetString(encrypted)
		} else {
			if dbField.Type() == userField.Type() {
				dbField.Set(userField)
			}
		}
	}
	var err error
	// Generate HMAC for username. This is needed for a deterministic lookup.
	dbUser.UsernameHMAC, err = LookupHash(user.Username, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC for username: %w", err)
	}
	return dbUser, nil
}

// fromUserDB converts a models.UserDB to a models.User and decrypts PII fields.
func fromUserDB(dbUser *models.UserDB, key []byte) (*models.User, error) {
	user := &models.User{}

	dbUserVal := reflect.ValueOf(dbUser).Elem()
	userVal := reflect.ValueOf(user).Elem()
	userType := userVal.Type()

	for i := 0; i < dbUserVal.NumField(); i++ {
		dbField := dbUserVal.Field(i)
		dbTypeField := dbUserVal.Type().Field(i)
		userField := userVal.FieldByName(dbTypeField.Name)

		if !userField.IsValid() || !userField.CanSet() {
			continue
		}

		structField, found := userType.FieldByName(dbTypeField.Name)
		if !found {
			continue
		}

		if tag := structField.Tag.Get("pii"); tag == "true" {
			decrypted, err := Decrypt(dbField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", dbTypeField.Name, err)
			}
			userField.SetString(decrypted)
		} else {
			if userField.Type() == dbField.Type() {
				userField.Set(dbField)
			}
		}
	}
	return user, nil
}

// Create inserts a new user into the database.
func (s *SQLUserStore) Create(user *models.User) (int, error) {
	dbUser, err := toUserDB(user, s.encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO users (username, username_hmac, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, dbUser.Username, dbUser.UsernameHMAC, dbUser.PasswordHash, dbUser.Role, user.CreatedAt, user.UpdatedAt)
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
	dbUser := &models.UserDB{}
	err := row.Scan(&dbUser.ID, &dbUser.Username, &dbUser.PasswordHash, &dbUser.Role, &dbUser.CreatedAt, &dbUser.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.GetGlobalLogger().Infof("User with ID %d not found", id)
			return nil, ErrNotFound
		}
		return nil, err
	}

	return fromUserDB(dbUser, s.encryptionKey)
}

// Update updates an existing user in the database.
func (s *SQLUserStore) Update(user *models.User) error {
	dbUser, err := toUserDB(user, s.encryptionKey)
	if err != nil {
		return err
	}

	query := `UPDATE users SET username = ?, username_hmac = ?, password_hash = ?, role = ?, updated_at = ? WHERE user_id = ?`
	result, err := s.db.Exec(query, dbUser.Username, dbUser.UsernameHMAC, dbUser.PasswordHash, dbUser.Role, user.UpdatedAt, dbUser.ID)
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
	usernameHMAC, err := LookupHash(username, s.encryptionKey)
	if err != nil {
		return nil, err
	}

	query := `SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE username_hmac = ?`
	row := s.db.QueryRow(query, usernameHMAC)
	dbUser := &models.UserDB{}
	err = row.Scan(&dbUser.ID, &dbUser.Username, &dbUser.PasswordHash, &dbUser.Role, &dbUser.CreatedAt, &dbUser.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return fromUserDB(dbUser, s.encryptionKey)
}

// GetAll fetches all users from the database.
func (s *SQLUserStore) GetAll() ([]*models.User, error) {
	query := `SELECT user_id, username, password_hash, role, created_at, updated_at FROM users`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var users []*models.User
	for rows.Next() {
		dbUser := &models.UserDB{}
		err := rows.Scan(&dbUser.ID, &dbUser.Username, &dbUser.PasswordHash, &dbUser.Role, &dbUser.CreatedAt, &dbUser.UpdatedAt)
		if err != nil {
			return nil, err
		}

		user, err := fromUserDB(dbUser, s.encryptionKey)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdatePassword updates a user's password in the database.
func (s *SQLUserStore) UpdatePassword(id int, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE user_id = ?`
	logger.GetGlobalLogger().Infof("Updating password for user ID %d", id)
	result, err := s.db.Exec(query, passwordHash, id)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error updating password for user ID %d: %v", id, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting rows affected for user ID %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		logger.GetGlobalLogger().Errorf("No user found with ID %d to update password", id)
		return ErrNotFound
	}
	logger.GetGlobalLogger().Debugf("Password updated successfully for user ID %d", id)
	return nil
}
