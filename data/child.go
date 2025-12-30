package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"kitadoc-backend/models"

	"modernc.org/sqlite"
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
	encryptedFirstName, err := Encrypt(child.FirstName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt FirstName: %w", err)
	}

	encryptedLastName, err := Encrypt(child.LastName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt LastName: %w", err)
	}

	encryptedBirthdate, err := Encrypt(child.Birthdate.Format(time.RFC3339Nano), key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt Birthdate: %w", err)
	}

	dbChild := &models.ChildDB{
		ID:        child.ID,
		FirstName: encryptedFirstName,
		LastName:  encryptedLastName,
		Birthdate: encryptedBirthdate,
		CreatedAt: child.CreatedAt,
		UpdatedAt: child.UpdatedAt,
	}

	if child.AdmissionDate != nil {
		dbChild.AdmissionDate = sql.NullTime{Time: *child.AdmissionDate, Valid: true}
	} else {
		dbChild.AdmissionDate = sql.NullTime{Valid: false}
	}

	if child.ExpectedSchoolEnrollment != nil {
		dbChild.ExpectedSchoolEnrollment = sql.NullTime{Time: *child.ExpectedSchoolEnrollment, Valid: true}
	} else {
		dbChild.ExpectedSchoolEnrollment = sql.NullTime{Valid: false}
	}

	return dbChild, nil
}

// fromChildDB converts a models.ChildDB to a models.Child and decrypts PII fields.
func fromChildDB(dbChild *models.ChildDB, key []byte) (*models.Child, error) {
	decryptedFirstName, err := Decrypt(dbChild.FirstName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt FirstName: %w", err)
	}

	decryptedLastName, err := Decrypt(dbChild.LastName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt LastName: %w", err)
	}

	decryptedBirthdate, err := Decrypt(dbChild.Birthdate, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt Birthdate: %w", err)
	}

	parsedBirthdate, err := time.Parse(time.RFC3339Nano, decryptedBirthdate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Birthdate: %w", err)
	}

	child := &models.Child{
		ID:        dbChild.ID,
		FirstName: decryptedFirstName,
		LastName:  decryptedLastName,
		Birthdate: parsedBirthdate,
		CreatedAt: dbChild.CreatedAt,
		UpdatedAt: dbChild.UpdatedAt,
	}

	if dbChild.AdmissionDate.Valid {
		child.AdmissionDate = &dbChild.AdmissionDate.Time
	}

	if dbChild.ExpectedSchoolEnrollment.Valid {
		child.ExpectedSchoolEnrollment = &dbChild.ExpectedSchoolEnrollment.Time
	}

	return child, nil
}

// Create inserts a new child into the database.
func (s *SQLChildStore) Create(child *models.Child) (int, error) {
	dbChild, err := toChildDB(child, s.encryptionKey)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO children (first_name, last_name, birthdate, admission_date, expected_school_enrollment) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, dbChild.FirstName, dbChild.LastName, dbChild.Birthdate, dbChild.AdmissionDate, dbChild.ExpectedSchoolEnrollment)
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
	query := `SELECT child_id, first_name, last_name, birthdate, admission_date, expected_school_enrollment, created_at, updated_at FROM children WHERE child_id = ?`
	row := s.db.QueryRow(query, id)
	dbChild := &models.ChildDB{}
	err := row.Scan(&dbChild.ID, &dbChild.FirstName, &dbChild.LastName, &dbChild.Birthdate, &dbChild.AdmissionDate, &dbChild.ExpectedSchoolEnrollment, &dbChild.CreatedAt, &dbChild.UpdatedAt)
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

	query := `UPDATE children SET first_name = ?, last_name = ?, birthdate = ?, admission_date = ?, expected_school_enrollment = ? WHERE child_id = ?`
	result, err := s.db.Exec(query, dbChild.FirstName, dbChild.LastName, dbChild.Birthdate, dbChild.AdmissionDate, dbChild.ExpectedSchoolEnrollment, dbChild.ID)
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
		// Check for foreign key constraint violation
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == 1811 || code == 787 {
				return ErrForeignKeyConstraint
			}
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

// GetAll fetches all children with pagination and filtering options.
func (s *SQLChildStore) GetAll() ([]models.Child, error) {
	query := `SELECT child_id, first_name, last_name, birthdate, admission_date, expected_school_enrollment, created_at, updated_at FROM children`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var children []models.Child
	for rows.Next() {
		dbChild := &models.ChildDB{}
		err := rows.Scan(&dbChild.ID, &dbChild.FirstName, &dbChild.LastName, &dbChild.Birthdate, &dbChild.AdmissionDate, &dbChild.ExpectedSchoolEnrollment, &dbChild.CreatedAt, &dbChild.UpdatedAt)
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
