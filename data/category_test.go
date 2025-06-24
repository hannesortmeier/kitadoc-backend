package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/internal/testutils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSQLCategoryStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	category := &models.Category{
		Name:        "Test Category",
		Description: testutils.StringPtr("A category for testing"),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO categories (category_name, description) VALUES (?, ?)`)).
			WithArgs(category.Name, category.Description).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(category)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO categories (category_name, description) VALUES (?, ?)`)).
			WithArgs(category.Name, category.Description).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(category)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLCategoryStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	categoryID := 1
	expectedCategory := &models.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Description: testutils.StringPtr("A category for testing"),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"category_id", "category_name", "description"}).
			AddRow(expectedCategory.ID, expectedCategory.Name, expectedCategory.Description)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnRows(rows)

		category, err := store.GetByID(categoryID)
		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, expectedCategory.ID, category.ID)
		assert.Equal(t, expectedCategory.Name, category.Name)
		assert.Equal(t, expectedCategory.Description, category.Description)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnError(sql.ErrNoRows)

		category, err := store.GetByID(categoryID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, category)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnError(errors.New("db error"))

		category, err := store.GetByID(categoryID)
		assert.Error(t, err)
		assert.Nil(t, category)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLCategoryStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	category := &models.Category{
		ID:          1,
		Name:        "Updated Category",
		Description: testutils.StringPtr("An updated description"),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE categories SET category_name = ?, description = ? WHERE category_id = ?`)).
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(category)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE categories SET category_name = ?, description = ? WHERE category_id = ?`)).
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(category)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE categories SET category_name = ?, description = ? WHERE category_id = ?`)).
			WithArgs(category.Name, category.Description, category.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(category)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLCategoryStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	categoryID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(categoryID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(categoryID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM categories WHERE category_id = ?`)).
			WithArgs(categoryID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(categoryID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLCategoryStore_GetByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	categoryName := "Existing Category"
	expectedCategory := &models.Category{
		ID:          1,
		Name:        categoryName,
		Description: testutils.StringPtr("Description of existing category"),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"category_id", "category_name", "description"}).
			AddRow(expectedCategory.ID, expectedCategory.Name, expectedCategory.Description)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_name = ?`)).
			WithArgs(categoryName).
			WillReturnRows(rows)

		category, err := store.GetByName(categoryName)
		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, expectedCategory.ID, category.ID)
		assert.Equal(t, expectedCategory.Name, category.Name)
		assert.Equal(t, expectedCategory.Description, category.Description)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_name = ?`)).
			WithArgs(categoryName).
			WillReturnError(sql.ErrNoRows)

		category, err := store.GetByName(categoryName)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, category)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories WHERE category_name = ?`)).
			WithArgs(categoryName).
			WillReturnError(errors.New("db error"))

		category, err := store.GetByName(categoryName)
		assert.Error(t, err)
		assert.Nil(t, category)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLCategoryStore_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := data.NewSQLCategoryStore(db)

	categories := []models.Category{
		{ID: 1, Name: "Category A", Description: testutils.StringPtr("Desc A")},
		{ID: 2, Name: "Category B", Description: testutils.StringPtr("Desc B")},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"category_id", "category_name", "description"}).
			AddRow(categories[0].ID, categories[0].Name, categories[0].Description).
			AddRow(categories[1].ID, categories[1].Name, categories[1].Description)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories`)).
			WillReturnRows(rows)

		fetchedCategories, err := store.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, fetchedCategories)
		assert.Len(t, fetchedCategories, 2)
		assert.Equal(t, categories[0].ID, fetchedCategories[0].ID)
		assert.Equal(t, categories[1].ID, fetchedCategories[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no categories found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories`)).
			WillReturnRows(sqlmock.NewRows([]string{"category_id", "category_name", "description"}))

		fetchedCategories, err := store.GetAll()
		assert.NoError(t, err)
		assert.Nil(t, fetchedCategories)
		assert.Len(t, fetchedCategories, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT category_id, category_name, description FROM categories`)).
			WillReturnError(errors.New("db error"))

		fetchedCategories, err := store.GetAll()
		assert.Error(t, err)
		assert.Nil(t, fetchedCategories)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}