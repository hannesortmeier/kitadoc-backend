package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/models"
)

// TeacherStore defines the interface for Teacher data operations.
type TeacherStore interface {
	Create(teacher *models.Teacher) (int, error)
	GetByID(id int) (*models.Teacher, error)
	Update(teacher *models.Teacher) error
	Delete(id int) error
	GetAll() ([]models.Teacher, error)
}

// SQLTeacherStore implements TeacherStore using database/sql.
type SQLTeacherStore struct {
	db *sql.DB
}

// NewSQLTeacherStore creates a new SQLTeacherStore.
func NewSQLTeacherStore(db *sql.DB) *SQLTeacherStore {
	return &SQLTeacherStore{db: db}
}

// Create inserts a new teacher into the database.
func (s *SQLTeacherStore) Create(teacher *models.Teacher) (int, error) {
	query := `INSERT INTO teachers (first_name, last_name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, teacher.FirstName, teacher.LastName, teacher.CreatedAt, teacher.UpdatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches a teacher by ID from the database.
func (s *SQLTeacherStore) GetByID(id int) (*models.Teacher, error) {
	query := `SELECT teacher_id, first_name, last_name, created_at, updated_at FROM teachers WHERE teacher_id = ?`
	row := s.db.QueryRow(query, id)
	teacher := &models.Teacher{}
	err := row.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.CreatedAt, &teacher.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return teacher, nil
}

// Update updates an existing teacher in the database.
func (s *SQLTeacherStore) Update(teacher *models.Teacher) error {
	query := `UPDATE teachers SET first_name = ?, last_name = ?, updated_at = ? WHERE teacher_id = ?`
	result, err := s.db.Exec(query, teacher.FirstName, teacher.LastName, teacher.UpdatedAt, teacher.ID)
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

// Delete deletes a teacher by ID from the database.
func (s *SQLTeacherStore) Delete(id int) error {
	query := `DELETE FROM teachers WHERE teacher_id = ?`
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

// GetAll fetches all teachers from the database.
func (s *SQLTeacherStore) GetAll() ([]models.Teacher, error) {
	query := `SELECT teacher_id, first_name, last_name, created_at, updated_at FROM teachers`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var teachers []models.Teacher
	for rows.Next() {
		teacher := &models.Teacher{}
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.CreatedAt, &teacher.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, *teacher)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}
