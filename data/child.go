package data

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

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
	db            *sql.DB
	encryptionKey []byte
}

// NewSQLChildStore creates a new SQLChildStore.
func NewSQLChildStore(db *sql.DB, encryptionKey []byte) *SQLChildStore {
	return &SQLChildStore{db: db, encryptionKey: encryptionKey}
}

// toChildDB converts a models.Child to a models.ChildDB and encrypts PII fields.
func toChildDB(child *models.Child, key []byte) (*models.ChildDB, error) {
	dbChild := &models.ChildDB{}

	// Use reflection to iterate over the fields of the input struct.
	childVal := reflect.ValueOf(child).Elem()
	dbChildVal := reflect.ValueOf(dbChild).Elem()

	for i := 0; i < childVal.NumField(); i++ {
		childField := childVal.Field(i)
		childTypeField := childVal.Type().Field(i)
		// Find the corresponding field in the destination struct by name.
		dbField := dbChildVal.FieldByName(childTypeField.Name)

		// Skip if the field doesn't exist in the destination or cannot be set.
		if !dbField.IsValid() || !dbField.CanSet() {
			continue
		}

		// Check for the `pii:"true"` tag.
		if tag := childTypeField.Tag.Get("pii"); tag == "true" {
			var rawValue string
			// Convert the field's value to a string for encryption.
			switch childField.Kind() {
			case reflect.String:
				rawValue = childField.String()
			case reflect.Struct:
				// Handle time.Time fields specifically.
				if childField.Type() == reflect.TypeOf(time.Time{}) {
					rawValue = childField.Interface().(time.Time).Format(time.RFC3339Nano)
				}
			}

			// Encrypt the string value.
			encrypted, err := Encrypt(rawValue, key)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", childTypeField.Name, err)
			}
			// Set the encrypted value on the destination struct's field.
			dbField.SetString(encrypted)
		} else {
			// If not a PII field, copy the value directly if the types match.
			if dbField.Type() == childField.Type() {
				dbField.Set(childField)
			}
		}
	}
	return dbChild, nil
}

// fromChildDB converts a models.ChildDB to a models.Child and decrypts PII fields.
func fromChildDB(dbChild *models.ChildDB, key []byte) (*models.Child, error) {
	child := &models.Child{}

	// Use reflection to iterate over the fields of the input struct.
	dbChildVal := reflect.ValueOf(dbChild).Elem()
	childVal := reflect.ValueOf(child).Elem()
	childType := childVal.Type()

	for i := 0; i < dbChildVal.NumField(); i++ {
		dbField := dbChildVal.Field(i)
		dbTypeField := dbChildVal.Type().Field(i)
		// Find the corresponding field in the destination struct by name.
		childField := childVal.FieldByName(dbTypeField.Name)

		// Skip if the field doesn't exist in the destination or cannot be set.
		if !childField.IsValid() || !childField.CanSet() {
			continue
		}

		// We need to check the tag on the destination struct (models.Child).
		structField, found := childType.FieldByName(dbTypeField.Name)
		if !found {
			continue
		}

		// Check for the `pii:"true"` tag.
		if tag := structField.Tag.Get("pii"); tag == "true" {
			// Decrypt the string value from the database struct.
			decrypted, err := Decrypt(dbField.String(), key)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", dbTypeField.Name, err)
			}

			// Convert the decrypted string back to the correct type.
			switch childField.Kind() {
			case reflect.String:
				childField.SetString(decrypted)
			case reflect.Struct:
				// Handle time.Time fields specifically.
				if childField.Type() == reflect.TypeOf(time.Time{}) {
					parsedTime, err := time.Parse(time.RFC3339Nano, decrypted)
					if err != nil {
						return nil, fmt.Errorf("failed to parse decrypted time for field %s: %w", dbTypeField.Name, err)
					}
					childField.Set(reflect.ValueOf(parsedTime))
				}
			}
		} else {
			// If not a PII field, copy the value directly if the types match.
			if childField.Type() == dbField.Type() {
				childField.Set(dbField)
			}
		}
	}
	return child, nil
}

// Create inserts a new child into the database.
func (s *SQLChildStore) Create(child *models.Child) (int, error) {
	dbChild, err := toChildDB(child, s.encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO children (first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, dbChild.FirstName, dbChild.LastName, dbChild.Birthdate, dbChild.Gender, dbChild.FamilyLanguage, dbChild.MigrationBackground, dbChild.AdmissionDate, dbChild.ExpectedSchoolEnrollment, dbChild.Address, dbChild.Parent1Name, dbChild.Parent2Name)
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
	dbChild := &models.ChildDB{}
	err := row.Scan(&dbChild.ID, &dbChild.FirstName, &dbChild.LastName, &dbChild.Birthdate, &dbChild.Gender, &dbChild.FamilyLanguage, &dbChild.MigrationBackground, &dbChild.AdmissionDate, &dbChild.ExpectedSchoolEnrollment, &dbChild.Address, &dbChild.Parent1Name, &dbChild.Parent2Name, &dbChild.CreatedAt, &dbChild.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return fromChildDB(dbChild, s.encryptionKey)
}

// Update updates an existing child in the database.
func (s *SQLChildStore) Update(child *models.Child) error {
	dbChild, err := toChildDB(child, s.encryptionKey)
	if err != nil {
		return err
	}

	query := `UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, gender = ?, family_language = ?, migration_background = ?, admission_date = ?, expected_school_enrollment = ?, address = ?, parent1_name = ?, parent2_name = ? WHERE child_id = ?`
	result, err := s.db.Exec(query, dbChild.FirstName, dbChild.LastName, dbChild.Birthdate, dbChild.Gender, dbChild.FamilyLanguage, dbChild.MigrationBackground, dbChild.AdmissionDate, dbChild.ExpectedSchoolEnrollment, dbChild.Address, dbChild.Parent1Name, dbChild.Parent2Name, dbChild.ID)
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
		dbChild := &models.ChildDB{}
		err := rows.Scan(&dbChild.ID, &dbChild.FirstName, &dbChild.LastName, &dbChild.Birthdate, &dbChild.Gender, &dbChild.FamilyLanguage, &dbChild.MigrationBackground, &dbChild.AdmissionDate, &dbChild.ExpectedSchoolEnrollment, &dbChild.Address, &dbChild.Parent1Name, &dbChild.Parent2Name, &dbChild.CreatedAt, &dbChild.UpdatedAt)
		if err != nil {
			return nil, err
		}

		child, err := fromChildDB(dbChild, s.encryptionKey)
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
