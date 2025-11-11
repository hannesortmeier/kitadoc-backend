package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/models"

	"github.com/mattn/go-sqlite3"
)

// CategoryStore defines the interface for Category data operations.
type CategoryStore interface {
	Create(category *models.Category) (int, error)
	GetByID(id int) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id int) error
	GetByName(name string) (*models.Category, error)
	GetAll() ([]models.Category, error)
}

// SQLCategoryStore implements CategoryStore using database/sql.
type SQLCategoryStore struct {
	db *sql.DB
}

// NewSQLCategoryStore creates a new SQLCategoryStore.
func NewSQLCategoryStore(db *sql.DB) *SQLCategoryStore {
	return &SQLCategoryStore{db: db}
}

// Create inserts a new category into the database.
func (s *SQLCategoryStore) Create(category *models.Category) (int, error) {
	query := `INSERT INTO categories (category_name, description) VALUES (?, ?)`
	result, err := s.db.Exec(query, category.Name, category.Description)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches a category by ID from the database.
func (s *SQLCategoryStore) GetByID(id int) (*models.Category, error) {
	query := `SELECT category_id, category_name, description FROM categories WHERE category_id = ?`
	row := s.db.QueryRow(query, id)
	category := &models.Category{}
	err := row.Scan(&category.ID, &category.Name, &category.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

// Update updates an existing category in the database.
func (s *SQLCategoryStore) Update(category *models.Category) error {
	query := `UPDATE categories SET category_name = ?, description = ? WHERE category_id = ?`
	result, err := s.db.Exec(query, category.Name, category.Description, category.ID)
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

// Delete deletes a category by ID from the database.
func (s *SQLCategoryStore) Delete(id int) error {
	query := `DELETE FROM categories WHERE category_id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		// Check for foreign key constraint violation
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && (sqliteErr.ExtendedCode == 1811 || sqliteErr.ExtendedCode == 787) {
			return ErrForeignKeyConstraint
		}
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

// GetByName fetches a category by name from the database.
func (s *SQLCategoryStore) GetByName(name string) (*models.Category, error) {
	query := `SELECT category_id, category_name, description FROM categories WHERE category_name = ?`
	row := s.db.QueryRow(query, name)
	category := &models.Category{}
	err := row.Scan(&category.ID, &category.Name, &category.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

// GetAll fetches all categories from the database.
func (s *SQLCategoryStore) GetAll() ([]models.Category, error) {
	query := `SELECT category_id, category_name, description FROM categories`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var categories []models.Category
	for rows.Next() {
		category := &models.Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
