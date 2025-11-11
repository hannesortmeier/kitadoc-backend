package data

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"kitadoc-backend/models"

	"github.com/mattn/go-sqlite3"
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
	db            *sql.DB
	encryptionKey []byte
}

// NewSQLTeacherStore creates a new SQLTeacherStore.
func NewSQLTeacherStore(db *sql.DB, encryptionKey []byte) *SQLTeacherStore {
	return &SQLTeacherStore{db: db, encryptionKey: encryptionKey}
}

// toTeacherDB converts a models.Teacher to a models.TeacherDB and encrypts PII fields.
func toTeacherDB(teacher *models.Teacher, key []byte) (*models.TeacherDB, error) {
	dbTeacher := &models.TeacherDB{}

	teacherVal := reflect.ValueOf(teacher).Elem()
	dbTeacherVal := reflect.ValueOf(dbTeacher).Elem()

	for i := 0; i < teacherVal.NumField(); i++ {
		teacherField := teacherVal.Field(i)
		teacherTypeField := teacherVal.Type().Field(i)
		dbField := dbTeacherVal.FieldByName(teacherTypeField.Name)

		if !dbField.IsValid() || !dbField.CanSet() {
			continue
		}

		if tag := teacherTypeField.Tag.Get("pii"); tag == "true" {
			encrypted, err := Encrypt(teacherField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", teacherTypeField.Name, err)
			}
			dbField.SetString(encrypted)
		} else {
			if dbField.Type() == teacherField.Type() {
				dbField.Set(teacherField)
			}
		}
	}
	return dbTeacher, nil
}

// fromTeacherDB converts a models.TeacherDB to a models.Teacher and decrypts PII fields.
func fromTeacherDB(dbTeacher *models.TeacherDB, key []byte) (*models.Teacher, error) {
	teacher := &models.Teacher{}

	dbTeacherVal := reflect.ValueOf(dbTeacher).Elem()
	teacherVal := reflect.ValueOf(teacher).Elem()
	teacherType := teacherVal.Type()

	for i := 0; i < dbTeacherVal.NumField(); i++ {
		dbField := dbTeacherVal.Field(i)
		dbTypeField := dbTeacherVal.Type().Field(i)
		teacherField := teacherVal.FieldByName(dbTypeField.Name)

		if !teacherField.IsValid() || !teacherField.CanSet() {
			continue
		}

		structField, found := teacherType.FieldByName(dbTypeField.Name)
		if !found {
			continue
		}

		if tag := structField.Tag.Get("pii"); tag == "true" {
			decrypted, err := Decrypt(dbField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", dbTypeField.Name, err)
			}
			teacherField.SetString(decrypted)
		} else {
			if teacherField.Type() == dbField.Type() {
				teacherField.Set(dbField)
			}
		}
	}
	return teacher, nil
}

// Create inserts a new teacher into the database.
func (s *SQLTeacherStore) Create(teacher *models.Teacher) (int, error) {
	dbTeacher, err := toTeacherDB(teacher, s.encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO teachers (first_name, last_name, username, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, dbTeacher.FirstName, dbTeacher.LastName, dbTeacher.Username, teacher.CreatedAt, teacher.UpdatedAt)
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
	query := `SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers WHERE teacher_id = ?`
	row := s.db.QueryRow(query, id)
	dbTeacher := &models.TeacherDB{}
	err := row.Scan(&dbTeacher.ID, &dbTeacher.FirstName, &dbTeacher.LastName, &dbTeacher.Username, &dbTeacher.CreatedAt, &dbTeacher.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return fromTeacherDB(dbTeacher, s.encryptionKey)
}

// Update updates an existing teacher in the database.
func (s *SQLTeacherStore) Update(teacher *models.Teacher) error {
	dbTeacher, err := toTeacherDB(teacher, s.encryptionKey)
	if err != nil {
		return err
	}

	query := `UPDATE teachers SET first_name = ?, last_name = ?, username = ?, updated_at = ? WHERE teacher_id = ?`
	result, err := s.db.Exec(query, dbTeacher.FirstName, dbTeacher.LastName, dbTeacher.Username, teacher.UpdatedAt, dbTeacher.ID)
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

// GetAll fetches all teachers from the database.
func (s *SQLTeacherStore) GetAll() ([]models.Teacher, error) {
	query := `SELECT teacher_id, first_name, last_name, username, created_at, updated_at FROM teachers`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var teachers []models.Teacher
	for rows.Next() {
		dbTeacher := &models.TeacherDB{}
		err := rows.Scan(&dbTeacher.ID, &dbTeacher.FirstName, &dbTeacher.LastName, &dbTeacher.Username, &dbTeacher.CreatedAt, &dbTeacher.UpdatedAt)
		if err != nil {
			return nil, err
		}

		teacher, err := fromTeacherDB(dbTeacher, s.encryptionKey)
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
