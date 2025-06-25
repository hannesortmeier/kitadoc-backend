package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/internal/testutils"
	"kitadoc-backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSQLChildStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db)

	groupID := 1
	child := &models.Child{
		FirstName:                "John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		GroupID:                  &groupID,
		FamilyLanguage:           testutils.StringPtr("English"),
		MigrationBackground:      testutils.BoolPtr(false),
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: testutils.TimePtr(time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)),
		Address:                  testutils.StringPtr("123 Main St"),
		Parent1Name:              testutils.StringPtr("Jane Doe"),
		Parent2Name:              nil,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO children (first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(child.FirstName, child.LastName, child.Birthdate, child.GroupID, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(child)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO children (first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(child.FirstName, child.LastName, child.Birthdate, child.GroupID, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(child)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLChildStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db)

	childID := 1
	groupID := 1
	expectedChild := &models.Child{
		ID:                       childID,
		FirstName:                "John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		GroupID:                  &groupID,
		FamilyLanguage:           testutils.StringPtr("English"),
		MigrationBackground:      testutils.BoolPtr(false),
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: testutils.TimePtr(time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)),
		Address:                  testutils.StringPtr("123 Main St"),
		Parent1Name:              testutils.StringPtr("Jane Doe"),
		Parent2Name:              nil,
		CreatedAt:                time.Now().Truncate(time.Second),
		UpdatedAt:                time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "group_id", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}).
			AddRow(expectedChild.ID, expectedChild.FirstName, expectedChild.LastName, expectedChild.Birthdate, expectedChild.GroupID, expectedChild.FamilyLanguage, expectedChild.MigrationBackground, expectedChild.AdmissionDate, expectedChild.ExpectedSchoolEnrollment, expectedChild.Address, expectedChild.Parent1Name, expectedChild.Parent2Name, expectedChild.CreatedAt, expectedChild.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnRows(rows)

		child, err := store.GetByID(childID)
		assert.NoError(t, err)
		assert.NotNil(t, child)
		assert.Equal(t, expectedChild.ID, child.ID)
		assert.Equal(t, expectedChild.FirstName, child.FirstName)
		assert.Equal(t, expectedChild.LastName, child.LastName)
		assert.WithinDuration(t, expectedChild.Birthdate, child.Birthdate, time.Second)
		assert.Equal(t, expectedChild.GroupID, child.GroupID)
		assert.Equal(t, expectedChild.FamilyLanguage, child.FamilyLanguage)
		assert.Equal(t, expectedChild.MigrationBackground, child.MigrationBackground)
		assert.WithinDuration(t, expectedChild.AdmissionDate, child.AdmissionDate, time.Second)
		assert.WithinDuration(t, *expectedChild.ExpectedSchoolEnrollment, *child.ExpectedSchoolEnrollment, time.Second)
		assert.Equal(t, expectedChild.Address, child.Address)
		assert.Equal(t, expectedChild.Parent1Name, child.Parent1Name)
		assert.Equal(t, expectedChild.Parent2Name, child.Parent2Name)
		assert.WithinDuration(t, expectedChild.CreatedAt, child.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedChild.UpdatedAt, child.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnError(sql.ErrNoRows)

		child, err := store.GetByID(childID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, child)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnError(errors.New("db error"))

		child, err := store.GetByID(childID)
		assert.Error(t, err)
		assert.Nil(t, child)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLChildStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db)

	groupID := 1
	child := &models.Child{
		ID:                       1,
		FirstName:                "Updated John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		GroupID:                  &groupID,
		FamilyLanguage:           testutils.StringPtr("German"),
		MigrationBackground:      testutils.BoolPtr(true),
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: testutils.TimePtr(time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)),
		Address:                  testutils.StringPtr("456 Oak Ave"),
		Parent1Name:              testutils.StringPtr("Jane Doe"),
		Parent2Name:              testutils.StringPtr("John Doe Sr."),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, group_id = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(child.FirstName, child.LastName, child.Birthdate, child.GroupID, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name, child.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(child)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, group_id = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(child.FirstName, child.LastName, child.Birthdate, child.GroupID, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name, child.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(child)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, group_id = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(child.FirstName, child.LastName, child.Birthdate, child.GroupID, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name, child.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(child)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLChildStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db)

	childID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(childID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(childID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(childID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLChildStore_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db)

	now := time.Now().Truncate(time.Second)
	groupID := 1
	children := []models.Child{
		{
			ID:                       1,
			FirstName:                "Child A",
			LastName:                 "Last A",
			Birthdate:                now.AddDate(-5, 0, 0),
			GroupID:                  &groupID,
			FamilyLanguage:           testutils.StringPtr("English"),
			MigrationBackground:      testutils.BoolPtr(false),
			AdmissionDate:            now.AddDate(-2, 0, 0),
			ExpectedSchoolEnrollment: testutils.TimePtr(now.AddDate(1, 0, 0)),
			Address:                  testutils.StringPtr("Address A"),
			Parent1Name:              testutils.StringPtr("Parent A1"),
			Parent2Name:              nil,
			CreatedAt:                now.AddDate(-3, 0, 0),
			UpdatedAt:                now.AddDate(-3, 0, 0),
		},
		{
			ID:                       2,
			FirstName:                "Child B",
			LastName:                 "Last B",
			Birthdate:                now.AddDate(-6, 0, 0),
			GroupID:                  &groupID,
			FamilyLanguage:           testutils.StringPtr("German"),
			MigrationBackground:      testutils.BoolPtr(true),
			AdmissionDate:            now.AddDate(-3, 0, 0),
			ExpectedSchoolEnrollment: testutils.TimePtr(now.AddDate(0, 0, 0)),
			Address:                  testutils.StringPtr("Address B"),
			Parent1Name:              testutils.StringPtr("Parent B1"),
			Parent2Name:              testutils.StringPtr("Parent B2"),
			CreatedAt:                now.AddDate(-4, 0, 0),
			UpdatedAt:                now.AddDate(-4, 0, 0),
		},
	}

	t.Run("success - no group filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "group_id", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}).
			AddRow(children[0].ID, children[0].FirstName, children[0].LastName, children[0].Birthdate, children[0].GroupID, children[0].FamilyLanguage, children[0].MigrationBackground, children[0].AdmissionDate, children[0].ExpectedSchoolEnrollment, children[0].Address, children[0].Parent1Name, children[0].Parent2Name, children[0].CreatedAt, children[0].UpdatedAt).
			AddRow(children[1].ID, children[1].FirstName, children[1].LastName, children[1].Birthdate, children[1].GroupID, children[1].FamilyLanguage, children[1].MigrationBackground, children[1].AdmissionDate, children[1].ExpectedSchoolEnrollment, children[1].Address, children[1].Parent1Name, children[1].Parent2Name, children[1].CreatedAt, children[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnRows(rows)

		fetchedChildren, err := store.GetAll(nil)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedChildren)
		assert.Len(t, fetchedChildren, 2)
		assert.Equal(t, children[0].ID, fetchedChildren[0].ID)
		assert.Equal(t, children[1].ID, fetchedChildren[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - with group filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "group_id", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}).
			AddRow(children[0].ID, children[0].FirstName, children[0].LastName, children[0].Birthdate, children[0].GroupID, children[0].FamilyLanguage, children[0].MigrationBackground, children[0].AdmissionDate, children[0].ExpectedSchoolEnrollment, children[0].Address, children[0].Parent1Name, children[0].Parent2Name, children[0].CreatedAt, children[0].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE group_id = ?`)).
			WithArgs(groupID).
			WillReturnRows(rows)

		fetchedChildren, err := store.GetAll(&groupID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedChildren)
		assert.Len(t, fetchedChildren, 1)
		assert.Equal(t, children[0].ID, fetchedChildren[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no children found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnRows(sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "group_id", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}))

		fetchedChildren, err := store.GetAll(nil)
		assert.NoError(t, err)
		assert.Nil(t, fetchedChildren)
		assert.Len(t, fetchedChildren, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnError(errors.New("db error"))

		fetchedChildren, err := store.GetAll(nil)
		assert.Error(t, err)
		assert.Nil(t, fetchedChildren)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
