package data

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"kitadoc-backend/models"
)

// DocumentationEntryStore defines the interface for DocumentationEntry data operations.
type DocumentationEntryStore interface {
	Create(entry *models.DocumentationEntry) (int, error)
	GetByID(id int) (*models.DocumentationEntry, error)
	Update(entry *models.DocumentationEntry) error
	Delete(id int) error
	GetAllForChild(childID int) ([]models.DocumentationEntry, error)
	ApproveEntry(entryID int, approvedByTeacherID int) error
}

// SQLDocumentationEntryStore implements DocumentationEntryStore using database/sql.
type SQLDocumentationEntryStore struct {
	db            *sql.DB
	encryptionKey []byte
}

// NewSQLDocumentationEntryStore creates a new SQLDocumentationEntryStore.
func NewSQLDocumentationEntryStore(db *sql.DB, encryptionKey []byte) *SQLDocumentationEntryStore {
	return &SQLDocumentationEntryStore{db: db, encryptionKey: encryptionKey}
}

// toDocumentationEntryDB converts a models.DocumentationEntry to a models.DocumentationEntryDB and encrypts PII fields.
func toDocumentationEntryDB(entry *models.DocumentationEntry, key []byte) (*models.DocumentationEntryDB, error) {
	dbEntry := &models.DocumentationEntryDB{}

	entryVal := reflect.ValueOf(entry).Elem()
	dbEntryVal := reflect.ValueOf(dbEntry).Elem()

	for i := 0; i < entryVal.NumField(); i++ {
		entryField := entryVal.Field(i)
		entryTypeField := entryVal.Type().Field(i)
		dbField := dbEntryVal.FieldByName(entryTypeField.Name)

		if !dbField.IsValid() || !dbField.CanSet() {
			continue
		}

		if tag := entryTypeField.Tag.Get("pii"); tag == "true" {
			encrypted, err := Encrypt(entryField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", entryTypeField.Name, err)
			}
			dbField.SetString(encrypted)
		} else {
			if dbField.Type() == entryField.Type() {
				dbField.Set(entryField)
			}
		}
	}
	return dbEntry, nil
}

// fromDocumentationEntryDB converts a models.DocumentationEntryDB to a models.DocumentationEntry and decrypts PII fields.
func fromDocumentationEntryDB(dbEntry *models.DocumentationEntryDB, key []byte) (*models.DocumentationEntry, error) {
	entry := &models.DocumentationEntry{}

	dbEntryVal := reflect.ValueOf(dbEntry).Elem()
	entryVal := reflect.ValueOf(entry).Elem()
	entryType := entryVal.Type()

	for i := 0; i < dbEntryVal.NumField(); i++ {
		dbField := dbEntryVal.Field(i)
		dbTypeField := dbEntryVal.Type().Field(i)
		entryField := entryVal.FieldByName(dbTypeField.Name)

		if !entryField.IsValid() || !entryField.CanSet() {
			continue
		}

		structField, found := entryType.FieldByName(dbTypeField.Name)
		if !found {
			continue
		}

		if tag := structField.Tag.Get("pii"); tag == "true" {
			decrypted, err := Decrypt(dbField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", dbTypeField.Name, err)
			}
			entryField.SetString(decrypted)
		} else {
			if entryField.Type() == dbField.Type() {
				entryField.Set(dbField)
			}
		}
	}
	return entry, nil
}

// Create inserts a new documentation entry into the database.
func (s *SQLDocumentationEntryStore) Create(entry *models.DocumentationEntry) (int, error) {
	dbEntry, err := toDocumentationEntryDB(entry, s.encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO documentation_entries (child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved, approved_by_teacher_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, dbEntry.ChildID, dbEntry.TeacherID, dbEntry.CategoryID, dbEntry.ObservationDate, dbEntry.ObservationDescription, dbEntry.IsApproved, dbEntry.ApprovedByUserID, dbEntry.CreatedAt, dbEntry.UpdatedAt)
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
	query := `SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved, approved_by_teacher_id, created_at, updated_at FROM documentation_entries WHERE entry_id = ?`
	row := s.db.QueryRow(query, id)
	dbEntry := &models.DocumentationEntryDB{}
	err := row.Scan(&dbEntry.ID, &dbEntry.ChildID, &dbEntry.TeacherID, &dbEntry.CategoryID, &dbEntry.ObservationDate, &dbEntry.ObservationDescription, &dbEntry.IsApproved, &dbEntry.ApprovedByUserID, &dbEntry.CreatedAt, &dbEntry.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return fromDocumentationEntryDB(dbEntry, s.encryptionKey)
}

// Update updates an existing documentation entry in the database.
func (s *SQLDocumentationEntryStore) Update(entry *models.DocumentationEntry) error {
	dbEntry, err := toDocumentationEntryDB(entry, s.encryptionKey)
	if err != nil {
		return err
	}

	query := `UPDATE documentation_entries SET child_id = ?, documenting_teacher_id = ?, category_id = ?, observation_date = ?, observation_description = ?, approved = ?, approved_by_teacher_id = ?, updated_at = ? WHERE entry_id = ?`
	result, err := s.db.Exec(query, dbEntry.ChildID, dbEntry.TeacherID, dbEntry.CategoryID, dbEntry.ObservationDate, dbEntry.ObservationDescription, dbEntry.IsApproved, dbEntry.ApprovedByUserID, dbEntry.UpdatedAt, dbEntry.ID)
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
	query := `SELECT entry_id, child_id, documenting_teacher_id, category_id, observation_date, observation_description, approved, approved_by_teacher_id, created_at, updated_at FROM documentation_entries WHERE child_id = ? ORDER BY observation_date DESC`
	rows, err := s.db.Query(query, childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var entries []models.DocumentationEntry
	for rows.Next() {
		dbEntry := &models.DocumentationEntryDB{}
		err := rows.Scan(&dbEntry.ID, &dbEntry.ChildID, &dbEntry.TeacherID, &dbEntry.CategoryID, &dbEntry.ObservationDate, &dbEntry.ObservationDescription, &dbEntry.IsApproved, &dbEntry.ApprovedByUserID, &dbEntry.CreatedAt, &dbEntry.UpdatedAt)
		if err != nil {
			return nil, err
		}

		entry, err := fromDocumentationEntryDB(dbEntry, s.encryptionKey)
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

// ApproveEntry sets the approved_by_teacher_id for a documentation entry.
func (s *SQLDocumentationEntryStore) ApproveEntry(entryID int, approvedByTeacherID int) error {
	query := `UPDATE documentation_entries SET approved_by_teacher_id = ?, approved = 1, updated_at = CURRENT_TIMESTAMP WHERE entry_id = ?`
	result, err := s.db.Exec(query, approvedByTeacherID, entryID)
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
