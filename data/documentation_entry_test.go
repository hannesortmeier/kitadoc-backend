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

func TestSQLDocumentationEntryStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	entry := &models.DocumentationEntry{
		ChildID:              1,
		TeacherID:            2,
		CategoryID:           3,
		ObservationDate:      time.Now(),
		ObservationDescription: "Test observation",
		ApprovedByUserID:     nil,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO documentation_entries (child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.CreatedAt, entry.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(entry)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO documentation_entries (child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.CreatedAt, entry.UpdatedAt).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(entry)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLDocumentationEntryStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	entryID := 1
	approvedByUserID := 10
	expectedEntry := &models.DocumentationEntry{
		ID:                     entryID,
		ChildID:                1,
		TeacherID:              2,
		CategoryID:             3,
		ObservationDate:        time.Now().Truncate(time.Second),
		ObservationDescription: "Test observation",
		ApprovedByUserID:       &approvedByUserID,
		CreatedAt:              time.Now().Truncate(time.Second),
		UpdatedAt:              time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"entry_id", "child_id", "documenting_teacher_id", "category_id", "observation_date", "observation_description", "approved_by_user_id", "created_at", "updated_at"}).
			AddRow(expectedEntry.ID, expectedEntry.ChildID, expectedEntry.TeacherID, expectedEntry.CategoryID, expectedEntry.ObservationDate, expectedEntry.ObservationDescription, expectedEntry.ApprovedByUserID, expectedEntry.CreatedAt, expectedEntry.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnRows(rows)

		entry, err := store.GetByID(entryID)
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry.ID, entry.ID)
		assert.Equal(t, expectedEntry.ChildID, entry.ChildID)
		assert.Equal(t, expectedEntry.TeacherID, entry.TeacherID)
		assert.Equal(t, expectedEntry.CategoryID, entry.CategoryID)
		assert.WithinDuration(t, expectedEntry.ObservationDate, entry.ObservationDate, time.Second)
		assert.Equal(t, expectedEntry.ObservationDescription, entry.ObservationDescription)
		assert.Equal(t, expectedEntry.ApprovedByUserID, entry.ApprovedByUserID)
		assert.WithinDuration(t, expectedEntry.CreatedAt, entry.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedEntry.UpdatedAt, entry.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnError(sql.ErrNoRows)

		entry, err := store.GetByID(entryID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, entry)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnError(errors.New("db error"))

		entry, err := store.GetByID(entryID)
		assert.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLDocumentationEntryStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	approvedByUserID := 10
	entry := &models.DocumentationEntry{
		ID:                     1,
		ChildID:                1,
		TeacherID:              2,
		CategoryID:             3,
		ObservationDate:        time.Now().Add(-time.Hour),
		ObservationDescription: "Updated observation",
		ApprovedByUserID:       &approvedByUserID,
		UpdatedAt:              time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET child_id = ?, documenting_teacher_id = ?, category_id = ?, observation_date = ?, observation_description = ?, approved_by_user_id = ?, updated_at = ? WHERE entry_id = ?`)).
			WithArgs(entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.UpdatedAt, entry.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(entry)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET child_id = ?, documenting_teacher_id = ?, category_id = ?, observation_date = ?, observation_description = ?, approved_by_user_id = ?, updated_at = ? WHERE entry_id = ?`)).
			WithArgs(entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.UpdatedAt, entry.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(entry)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET child_id = ?, documenting_teacher_id = ?, category_id = ?, observation_date = ?, observation_description = ?, approved_by_user_id = ?, updated_at = ? WHERE entry_id = ?`)).
			WithArgs(entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.UpdatedAt, entry.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(entry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLDocumentationEntryStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	entryID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(entryID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(entryID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM documentation_entries WHERE entry_id = ?`)).
			WithArgs(entryID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(entryID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLDocumentationEntryStore_GetAllForChild(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	childID := 1
	now := time.Now().Truncate(time.Second)
	approvedByUserID := 10
	entries := []models.DocumentationEntry{
		{
			ID:                     1,
			ChildID:                childID,
			TeacherID:              1,
			CategoryID:             1,
			ObservationDate:        now.Add(-time.Hour * 24),
			ObservationDescription: "Entry 1",
			ApprovedByUserID:       &approvedByUserID,
			CreatedAt:              now.Add(-time.Hour * 25),
			UpdatedAt:              now.Add(-time.Hour * 25),
		},
		{
			ID:                     2,
			ChildID:                childID,
			TeacherID:              2,
			CategoryID:             2,
			ObservationDate:        now.Add(-time.Hour * 48),
			ObservationDescription: "Entry 2",
			ApprovedByUserID:       nil,
			CreatedAt:              now.Add(-time.Hour * 49),
			UpdatedAt:              now.Add(-time.Hour * 49),
		},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"entry_id", "child_id", "documenting_teacher_id", "category_id", "observation_date", "observation_description", "approved_by_user_id", "created_at", "updated_at"}).
			AddRow(entries[0].ID, entries[0].ChildID, entries[0].TeacherID, entries[0].CategoryID, entries[0].ObservationDate, entries[0].ObservationDescription, entries[0].ApprovedByUserID, entries[0].CreatedAt, entries[0].UpdatedAt).
			AddRow(entries[1].ID, entries[1].ChildID, entries[1].TeacherID, entries[1].CategoryID, entries[1].ObservationDate, entries[1].ObservationDescription, entries[1].ApprovedByUserID, entries[1].CreatedAt, entries[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`)).
			WithArgs(childID).
			WillReturnRows(rows)

		fetchedEntries, err := store.GetAllForChild(childID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedEntries)
		assert.Len(t, fetchedEntries, 2)
		assert.Equal(t, entries[0].ID, fetchedEntries[0].ID)
		assert.Equal(t, entries[1].ID, fetchedEntries[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no entries found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`)).
			WithArgs(childID).
			WillReturnRows(sqlmock.NewRows([]string{"entry_id", "child_id", "documenting_teacher_id", "category_id", "observation_date", "observation_description", "approved_by_user_id", "created_at", "updated_at"}))

		fetchedEntries, err := store.GetAllForChild(childID)
		assert.NoError(t, err)
		assert.Nil(t, fetchedEntries)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`)).
			WithArgs(childID).
			WillReturnError(errors.New("db error"))

		fetchedEntries, err := store.GetAllForChild(childID)
		assert.Error(t, err)
		assert.Nil(t, fetchedEntries)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"entry_id", "child_id", "documenting_teacher_id", "category_id", "observation_date", "observation_description", "approved_by_user_id", "created_at", "updated_at"}).
			AddRow(entries[0].ID, entries[0].ChildID, "not-an-int", entries[0].CategoryID, entries[0].ObservationDate, entries[0].ObservationDescription, entries[0].ApprovedByUserID, entries[0].CreatedAt, entries[0].UpdatedAt) // Malformed row

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`)).
			WithArgs(childID).
			WillReturnRows(rows)

		fetchedEntries, err := store.GetAllForChild(childID)
		assert.Error(t, err)
		assert.Nil(t, fetchedEntries)
		assert.Contains(t, err.Error(), "converting driver.Value type string")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLDocumentationEntryStore_ApproveEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLDocumentationEntryStore(db)

	entryID := 1
	approvedByUserID := 10

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET approved_by_user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE entry_id = ?`)).
			WithArgs(approvedByUserID, entryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.ApproveEntry(entryID, approvedByUserID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET approved_by_user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE entry_id = ?`)).
			WithArgs(approvedByUserID, entryID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.ApproveEntry(entryID, approvedByUserID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE documentation_entries SET approved_by_user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE entry_id = ?`)).
			WithArgs(approvedByUserID, entryID).
			WillReturnError(errors.New("db error"))

		err := store.ApproveEntry(entryID, approvedByUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}