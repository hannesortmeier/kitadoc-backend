package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSQLGroupStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	group := &models.Group{
		Name: "Test Group",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO groups (group_name) VALUES (?)`)).
			WithArgs(group.Name).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(group)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO groups (group_name) VALUES (?)`)).
			WithArgs(group.Name).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(group)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique constraint failed", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO groups (group_name) VALUES (?)`)).
			WithArgs(group.Name).
			WillReturnError(errors.New("UNIQUE constraint failed: groups.group_name"))

		id, err := store.Create(group)
		assert.Error(t, err)
		assert.Equal(t, data.ErrConflict, err)
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLGroupStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	groupID := 1
	expectedGroup := &models.Group{
		ID:   groupID,
		Name: "Test Group",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"group_id", "group_name"}).
			AddRow(expectedGroup.ID, expectedGroup.Name)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnRows(rows)

		group, err := store.GetByID(groupID)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, expectedGroup.ID, group.ID)
		assert.Equal(t, expectedGroup.Name, group.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnError(sql.ErrNoRows)

		group, err := store.GetByID(groupID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, group)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnError(errors.New("db error"))

		group, err := store.GetByID(groupID)
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLGroupStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	group := &models.Group{
		ID:   1,
		Name: "Updated Group",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE groups SET group_name = ? WHERE group_id = ?`)).
			WithArgs(group.Name, group.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(group)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE groups SET group_name = ? WHERE group_id = ?`)).
			WithArgs(group.Name, group.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(group)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE groups SET group_name = ? WHERE group_id = ?`)).
			WithArgs(group.Name, group.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(group)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLGroupStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	groupID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(groupID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(groupID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM groups WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(groupID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLGroupStore_GetByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	groupName := "Existing Group"
	expectedGroup := &models.Group{
		ID:   1,
		Name: groupName,
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"group_id", "group_name"}).
			AddRow(expectedGroup.ID, expectedGroup.Name)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_name = ?`)).
			WithArgs(groupName).
			WillReturnRows(rows)

		group, err := store.GetByName(groupName)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, expectedGroup.ID, group.ID)
		assert.Equal(t, expectedGroup.Name, group.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_name = ?`)).
			WithArgs(groupName).
			WillReturnError(sql.ErrNoRows)

		group, err := store.GetByName(groupName)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, group)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups WHERE group_name = ?`)).
			WithArgs(groupName).
			WillReturnError(errors.New("db error"))

		group, err := store.GetByName(groupName)
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLGroupStore_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLGroupStore(db)

	groups := []models.Group{
		{ID: 1, Name: "Group A"},
		{ID: 2, Name: "Group B"},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"group_id", "group_name"}).
			AddRow(groups[0].ID, groups[0].Name).
			AddRow(groups[1].ID, groups[1].Name)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups`)).
			WillReturnRows(rows)

		fetchedGroups, err := store.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, fetchedGroups)
		assert.Len(t, fetchedGroups, 2)
		assert.Equal(t, groups[0].ID, fetchedGroups[0].ID)
		assert.Equal(t, groups[1].ID, fetchedGroups[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no groups found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups`)).
			WillReturnRows(sqlmock.NewRows([]string{"group_id", "group_name"}))

		fetchedGroups, err := store.GetAll()
		assert.NoError(t, err)
		assert.Nil(t, fetchedGroups)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT group_id, group_name FROM groups`)).
			WillReturnError(errors.New("db error"))

		fetchedGroups, err := store.GetAll()
		assert.Error(t, err)
		assert.Nil(t, fetchedGroups)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}