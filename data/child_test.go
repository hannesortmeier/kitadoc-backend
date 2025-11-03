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

func TestSQLChildStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLChildStore(db, []byte("0123456789abcdef0123456789abcdef"))

	child := &models.Child{
		FirstName:                "John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:                   "Male",
		FamilyLanguage:           "English",
		MigrationBackground:      false,
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
		Address:                  "123 Main St",
		Parent1Name:              "Jane Doe",
		Parent2Name:              "John Doe Sr.",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO children (first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(child)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO children (first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	key := []byte("0123456789abcdef0123456789abcdef")
	store := data.NewSQLChildStore(db, key)

	childID := 1
	expectedChild := &models.Child{
		ID:                       childID,
		FirstName:                "John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		FamilyLanguage:           "English",
		MigrationBackground:      false,
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
		Address:                  "123 Main St",
		Parent1Name:              "Jane Doe",
		Parent2Name:              "John Doe Sr.",
		CreatedAt:                time.Now().Truncate(time.Second),
		UpdatedAt:                time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		encryptedFirstName, _ := data.Encrypt(expectedChild.FirstName, key)
		encryptedLastName, _ := data.Encrypt(expectedChild.LastName, key)
		encryptedBirthdate, _ := data.Encrypt(expectedChild.Birthdate.Format(time.RFC3339Nano), key)
		encryptedAddress, _ := data.Encrypt(expectedChild.Address, key)
		encryptedParent1Name, _ := data.Encrypt(expectedChild.Parent1Name, key)
		encryptedParent2Name, _ := data.Encrypt(expectedChild.Parent2Name, key)

		rows := sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "gender", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}).
			AddRow(expectedChild.ID, encryptedFirstName, encryptedLastName, encryptedBirthdate, expectedChild.Gender, expectedChild.FamilyLanguage, expectedChild.MigrationBackground, expectedChild.AdmissionDate, expectedChild.ExpectedSchoolEnrollment, encryptedAddress, encryptedParent1Name, encryptedParent2Name, expectedChild.CreatedAt, expectedChild.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnRows(rows)

		child, err := store.GetByID(childID)
		assert.NoError(t, err)
		assert.NotNil(t, child)
		assert.Equal(t, expectedChild.ID, child.ID)
		assert.Equal(t, expectedChild.FirstName, child.FirstName)
		assert.Equal(t, expectedChild.LastName, child.LastName)
		assert.WithinDuration(t, expectedChild.Birthdate, child.Birthdate, time.Second)
		assert.Equal(t, expectedChild.FamilyLanguage, child.FamilyLanguage)
		assert.Equal(t, expectedChild.MigrationBackground, child.MigrationBackground)
		assert.WithinDuration(t, expectedChild.AdmissionDate, child.AdmissionDate, time.Second)
		assert.WithinDuration(t, expectedChild.ExpectedSchoolEnrollment, child.ExpectedSchoolEnrollment, time.Second)
		assert.Equal(t, expectedChild.Address, child.Address)
		assert.Equal(t, expectedChild.Parent1Name, child.Parent1Name)
		assert.Equal(t, expectedChild.Parent2Name, child.Parent2Name)
		assert.WithinDuration(t, expectedChild.CreatedAt, child.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedChild.UpdatedAt, child.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
			WithArgs(childID).
			WillReturnError(sql.ErrNoRows)

		child, err := store.GetByID(childID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, child)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`)).
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

	store := data.NewSQLChildStore(db, []byte("0123456789abcdef0123456789abcdef"))

	child := &models.Child{
		ID:                       1,
		FirstName:                "Updated John",
		LastName:                 "Doe",
		Birthdate:                time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		FamilyLanguage:           "German",
		MigrationBackground:      true,
		AdmissionDate:            time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpectedSchoolEnrollment: time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
		Address:                  "456 Oak Ave",
		Parent1Name:              "Jane Doe",
		Parent2Name:              "John Doe Sr.",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, gender = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(child)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, gender = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(child)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, gender = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), child.ID).
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

	store := data.NewSQLChildStore(db, []byte("0123456789abcdef0123456789abcdef"))

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

	key := []byte("0123456789abcdef0123456789abcdef")
	store := data.NewSQLChildStore(db, key)

	now := time.Now().Truncate(time.Second)
	children := []models.Child{
		{
			ID:                       1,
			FirstName:                "Child A",
			LastName:                 "Last A",
			Birthdate:                now.AddDate(-5, 0, 0),
			FamilyLanguage:           "English",
			MigrationBackground:      false,
			AdmissionDate:            now.AddDate(-2, 0, 0),
			ExpectedSchoolEnrollment: now.AddDate(1, 0, 0),
			Address:                  "Address A",
			Parent1Name:              "Parent A1",
			Parent2Name:              "Parent A2",
			CreatedAt:                now.AddDate(-3, 0, 0),
			UpdatedAt:                now.AddDate(-3, 0, 0),
		},
		{
			ID:                       2,
			FirstName:                "Child B",
			LastName:                 "Last B",
			Birthdate:                now.AddDate(-6, 0, 0),
			FamilyLanguage:           "German",
			MigrationBackground:      true,
			AdmissionDate:            now.AddDate(-3, 0, 0),
			ExpectedSchoolEnrollment: now.AddDate(0, 0, 0),
			Address:                  "Address B",
			Parent1Name:              "Parent B1",
			Parent2Name:              "Parent B2",
			CreatedAt:                now.AddDate(-4, 0, 0),
			UpdatedAt:                now.AddDate(-4, 0, 0),
		},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "gender", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"})
		for _, child := range children {
			encryptedFirstName, _ := data.Encrypt(child.FirstName, key)
			encryptedLastName, _ := data.Encrypt(child.LastName, key)
			encryptedBirthdate, _ := data.Encrypt(child.Birthdate.Format(time.RFC3339Nano), key)
			encryptedAddress, _ := data.Encrypt(child.Address, key)
			encryptedParent1Name, _ := data.Encrypt(child.Parent1Name, key)
			encryptedParent2Name, _ := data.Encrypt(child.Parent2Name, key)
			rows.AddRow(child.ID, encryptedFirstName, encryptedLastName, encryptedBirthdate, child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, encryptedAddress, encryptedParent1Name, encryptedParent2Name, child.CreatedAt, child.UpdatedAt)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnRows(rows)

		fetchedChildren, err := store.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, fetchedChildren)
		assert.Len(t, fetchedChildren, 2)
		assert.Equal(t, children[0].ID, fetchedChildren[0].ID)
		assert.Equal(t, children[1].ID, fetchedChildren[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no children found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnRows(sqlmock.NewRows([]string{"child_id", "first_name", "last_name", "birthdate", "gender", "family_language", "migration_background", "admission_date", "expected_school_enrollment", "address", "parent1_name", "parent2_name", "created_at", "updated_at"}))

		fetchedChildren, err := store.GetAll()
		assert.NoError(t, err)
		assert.Len(t, fetchedChildren, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`)).
			WillReturnError(errors.New("db error"))

		fetchedChildren, err := store.GetAll()
		assert.Error(t, err)
		assert.Nil(t, fetchedChildren)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
