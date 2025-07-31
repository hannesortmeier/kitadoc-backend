package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSQLTeacherStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLTeacherStore(db)

	teacher := &models.Teacher{
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO teachers (first_name, last_name, username, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)).
			WithArgs(teacher.FirstName, teacher.LastName, teacher.Username, teacher.CreatedAt, teacher.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(teacher)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO teachers (first_name, last_name, username, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`)).
			WithArgs(teacher.FirstName, teacher.LastName, teacher.Username, teacher.CreatedAt, teacher.UpdatedAt).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(teacher)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLTeacherStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLTeacherStore(db)

	teacherID := 1
	expectedTeacher := &models.Teacher{
		ID:        teacherID,
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"teacher_id", "first_name", "last_name", "username", "created_at", "updated_at"}).
			AddRow(expectedTeacher.ID, expectedTeacher.FirstName, expectedTeacher.LastName, expectedTeacher.Username, expectedTeacher.CreatedAt, expectedTeacher.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers WHERE teacher_id = ?`)).
			WithArgs(teacherID).
			WillReturnRows(rows)

		teacher, err := store.GetByID(teacherID)
		assert.NoError(t, err)
		assert.NotNil(t, teacher)
		assert.Equal(t, expectedTeacher.ID, teacher.ID)
		assert.Equal(t, expectedTeacher.FirstName, teacher.FirstName)
		assert.Equal(t, expectedTeacher.LastName, teacher.LastName)
		assert.Equal(t, expectedTeacher.Username, teacher.Username)
		assert.WithinDuration(t, expectedTeacher.CreatedAt, teacher.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedTeacher.UpdatedAt, teacher.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers WHERE teacher_id = ?`)).
			WithArgs(teacherID).
			WillReturnError(sql.ErrNoRows)

		teacher, err := store.GetByID(teacherID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, teacher)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers WHERE teacher_id = ?`)).
			WithArgs(teacherID).
			WillReturnError(errors.New("db error"))

		teacher, err := store.GetByID(teacherID)
		assert.Error(t, err)
		assert.Nil(t, teacher)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLTeacherStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLTeacherStore(db)

	teacher := &models.Teacher{
		ID:        1,
		FirstName: "Updated Jane",
		LastName:  "Smith",
		Username:  "updatedjane",
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE teachers SET first_name = ?, last_name = ?, username = ?, updated_at = ? WHERE teacher_id = ?`)).
			WithArgs(teacher.FirstName, teacher.LastName, teacher.Username, teacher.UpdatedAt, teacher.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(teacher)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE teachers SET first_name = ?, last_name = ?, username = ?, updated_at = ? WHERE teacher_id = ?`)).
			WithArgs(teacher.FirstName, teacher.LastName, teacher.Username, teacher.UpdatedAt, teacher.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(teacher)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE teachers SET first_name = ?, last_name = ?, username = ?, updated_at = ? WHERE teacher_id = ?`)).
			WithArgs(teacher.FirstName, teacher.LastName, teacher.Username, teacher.UpdatedAt, teacher.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(teacher)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLTeacherStore_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLTeacherStore(db)

	now := time.Now().Truncate(time.Second)
	teachers := []models.Teacher{
		{ID: 1, FirstName: "Teacher A", LastName: "Last A", Username: "teachera", CreatedAt: now, UpdatedAt: now},
		{ID: 2, FirstName: "Teacher B", LastName: "Last B", Username: "teacherb", CreatedAt: now, UpdatedAt: now},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"teacher_id", "first_name", "last_name", "username", "created_at", "updated_at"}).
			AddRow(teachers[0].ID, teachers[0].FirstName, teachers[0].LastName, teachers[0].Username, teachers[0].CreatedAt, teachers[0].UpdatedAt).
			AddRow(teachers[1].ID, teachers[1].FirstName, teachers[1].LastName, teachers[1].Username, teachers[1].CreatedAt, teachers[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers`)).
			WillReturnRows(rows)

		fetchedTeachers, err := store.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, fetchedTeachers)
		assert.Len(t, fetchedTeachers, 2)
		assert.Equal(t, teachers[0].ID, fetchedTeachers[0].ID)
		assert.Equal(t, teachers[1].ID, fetchedTeachers[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no teachers found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers`)).
			WillReturnRows(sqlmock.NewRows([]string{"teacher_id", "first_name", "last_name", "username", "created_at", "updated_at"}))

		fetchedTeachers, err := store.GetAll()
		assert.NoError(t, err)
		assert.Nil(t, fetchedTeachers)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers`)).
			WillReturnError(errors.New("db error"))

		fetchedTeachers, err := store.GetAll()
		assert.Error(t, err)
		assert.Nil(t, fetchedTeachers)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
