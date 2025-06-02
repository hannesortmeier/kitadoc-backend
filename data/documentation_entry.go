package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/models"
)

// DocumentationEntryStore defines the interface for DocumentationEntry data operations.
type DocumentationEntryStore interface {
	Create(entry *models.DocumentationEntry) (int, error)
	GetByID(id int) (*models.DocumentationEntry, error)
	Update(entry *models.DocumentationEntry) error
	Delete(id int) error
	GetAllForChild(childID int) ([]models.DocumentationEntry, error)
	ApproveEntry(entryID int, approvedByUserID int) error
}

// SQLDocumentationEntryStore implements DocumentationEntryStore using database/sql.
type SQLDocumentationEntryStore struct {
	db *sql.DB
}

// NewSQLDocumentationEntryStore creates a new SQLDocumentationEntryStore.
func NewSQLDocumentationEntryStore(db *sql.DB) *SQLDocumentationEntryStore {
	return &SQLDocumentationEntryStore{db: db}
}

// Create inserts a new documentation entry into the database.
func (s *SQLDocumentationEntryStore) Create(entry *models.DocumentationEntry) (int, error) {
	query := `INSERT INTO documentation_entries (child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.CreatedAt, entry.UpdatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches a documentation entry by ID from the database.
func (s *SQLDocumentationEntryStore) GetByID(id int) (*models.DocumentationEntry, error) {
	query := `SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE entry_id = ?`
	row := s.db.QueryRow(query, id)
	entry := &models.DocumentationEntry{}
	err := row.Scan(&entry.ID, &entry.ChildID, &entry.TeacherID, &entry.CategoryID, &entry.ObservationDate, &entry.ObservationDescription, &entry.ApprovedByUserID, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return entry, nil
}

// Update updates an existing documentation entry in the database.
func (s *SQLDocumentationEntryStore) Update(entry *models.DocumentationEntry) error {
	query := `UPDATE documentation_entries SET child_id = ?, documenting_teacher_id = ?, category_id = ?, observation_date = ?, observation_description = ?, approved_by_user_id = ?, updated_at = ? WHERE entry_id = ?`
	result, err := s.db.Exec(query, entry.ChildID, entry.TeacherID, entry.CategoryID, entry.ObservationDate, entry.ObservationDescription, entry.ApprovedByUserID, entry.UpdatedAt, entry.ID)
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

// Delete deletes a documentation entry by ID from the database.
func (s *SQLDocumentationEntryStore) Delete(id int) error {
	query := `DELETE FROM documentation_entries WHERE entry_id = ?`
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

// GetAllForChild fetches all documentation entries for a specific child.
func (s *SQLDocumentationEntryStore) GetAllForChild(childID int) ([]models.DocumentationEntry, error) {
	query := `SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved_by_user_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`
	rows, err := s.db.Query(query, childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.DocumentationEntry
	for rows.Next() {
		entry := &models.DocumentationEntry{}
		err := rows.Scan(&entry.ID, &entry.ChildID, &entry.TeacherID, &entry.CategoryID, &entry.ObservationDate, &entry.ObservationDescription, &entry.ApprovedByUserID, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// ApproveEntry sets the approved_by_user_id for a documentation entry.
func (s *SQLDocumentationEntryStore) ApproveEntry(entryID int, approvedByUserID int) error {
	query := `UPDATE documentation_entries SET approved_by_user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE entry_id = ?`
	result, err := s.db.Exec(query, approvedByUserID, entryID)
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