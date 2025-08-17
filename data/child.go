package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/models"
)

// ChildStore defines the interface for Child data operations.
type ChildStore interface {
	Create(child *models.Child) (int, error)
	GetByID(id int) (*models.Child, error)
	Update(child *models.Child) error
	Delete(id int) error
	GetAll() ([]models.Child, error)
}

// SQLChildStore implements ChildStore using database/sql.
type SQLChildStore struct {
	db *sql.DB
}

// NewSQLChildStore creates a new SQLChildStore.
func NewSQLChildStore(db *sql.DB) *SQLChildStore {
	return &SQLChildStore{db: db}
}

// Create inserts a new child into the database.
func (s *SQLChildStore) Create(child *models.Child) (int, error) {
	query := `INSERT INTO children (first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, child.FirstName, child.LastName, child.Birthdate, child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches a child by ID from the database.
func (s *SQLChildStore) GetByID(id int) (*models.Child, error) {
	query := `SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children WHERE child_id = ?`
	row := s.db.QueryRow(query, id)
	child := &models.Child{}
	err := row.Scan(&child.ID, &child.FirstName, &child.LastName, &child.Birthdate, &child.Gender, &child.FamilyLanguage, &child.MigrationBackground, &child.AdmissionDate, &child.ExpectedSchoolEnrollment, &child.Address, &child.Parent1Name, &child.Parent2Name, &child.CreatedAt, &child.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return child, nil
}

// Update updates an existing child in the database.
func (s *SQLChildStore) Update(child *models.Child) error {
	query := `UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, gender = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`
	result, err := s.db.Exec(query, child.FirstName, child.LastName, child.Birthdate, child.Gender, child.FamilyLanguage, child.MigrationBackground, child.AdmissionDate, child.ExpectedSchoolEnrollment, child.Address, child.Parent1Name, child.Parent2Name, child.ID)
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

// Delete deletes a child by ID from the database.
func (s *SQLChildStore) Delete(id int) error {
	query := `DELETE FROM children WHERE child_id = ?`
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

// GetAll fetches all children with pagination and filtering options.
func (s *SQLChildStore) GetAll() ([]models.Child, error) {
	query := `SELECT child_id, first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name, created_at, updated_at FROM children`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var children []models.Child
	for rows.Next() {
		child := &models.Child{}
		err := rows.Scan(&child.ID, &child.FirstName, &child.LastName, &child.Birthdate, &child.Gender, &child.FamilyLanguage, &child.MigrationBackground, &child.AdmissionDate, &child.ExpectedSchoolEnrollment, &child.Address, &child.Parent1Name, &child.Parent2Name, &child.CreatedAt, &child.UpdatedAt)
		if err != nil {
			return nil, err
		}
		children = append(children, *child)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return children, nil
}
