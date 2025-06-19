package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/internal/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSQLUserStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLUserStore(db)

	user := &models.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "teacher",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	log_level, _ := logrus.ParseLevel("debug")

	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (username, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)).
			WithArgs(user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(user)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (username, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)).
			WithArgs(user.Username, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(user)
		assert.Error(t, err)
		assert.Equal(t, -1, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLUserStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLUserStore(db)

	userID := 1
	expectedUser := &models.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         "teacher",
		CreatedAt:    time.Now().Truncate(time.Second),
		UpdatedAt:    time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.PasswordHash, expectedUser.Role, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := store.GetByID(userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Username, user.Username)
		assert.Equal(t, expectedUser.PasswordHash, user.PasswordHash)
		assert.Equal(t, expectedUser.Role, user.Role)
		assert.WithinDuration(t, expectedUser.CreatedAt, user.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedUser.UpdatedAt, user.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := store.GetByID(userID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		user, err := store.GetByID(userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLUserStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLUserStore(db)

	user := &models.User{
		ID:           1,
		Username:     "updateduser",
		PasswordHash: "newhashedpassword",
		Role:         "admin",
		UpdatedAt:    time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET username = ?, password_hash = ?, role = ?, updated_at = ? WHERE user_id = ?`)).
			WithArgs(user.Username, user.PasswordHash, user.Role, user.UpdatedAt, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET username = ?, password_hash = ?, role = ?, updated_at = ? WHERE user_id = ?`)).
			WithArgs(user.Username, user.PasswordHash, user.Role, user.UpdatedAt, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(user)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET username = ?, password_hash = ?, role = ?, updated_at = ? WHERE user_id = ?`)).
			WithArgs(user.Username, user.PasswordHash, user.Role, user.UpdatedAt, user.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLUserStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLUserStore(db)

	userID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(userID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE user_id = ?`)).
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLUserStore_GetUserByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLUserStore(db)

	username := "testuser"
	expectedUser := &models.User{
		ID:           1,
		Username:     username,
		PasswordHash: "hashedpassword",
		Role:         "teacher",
		CreatedAt:    time.Now().Truncate(time.Second),
		UpdatedAt:    time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.PasswordHash, expectedUser.Role, expectedUser.CreatedAt, expectedUser.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE username = ?`)).
			WithArgs(username).
			WillReturnRows(rows)

		user, err := store.GetUserByUsername(username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Username, user.Username)
		assert.Equal(t, expectedUser.PasswordHash, user.PasswordHash)
		assert.Equal(t, expectedUser.Role, user.Role)
		assert.WithinDuration(t, expectedUser.CreatedAt, user.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedUser.UpdatedAt, user.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE username = ?`)).
			WithArgs(username).
			WillReturnError(sql.ErrNoRows)

		user, err := store.GetUserByUsername(username)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, username, password_hash, role, created_at, updated_at FROM users WHERE username = ?`)).
			WithArgs(username).
			WillReturnError(errors.New("db error"))

		user, err := store.GetUserByUsername(username)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}