package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/models"
)

// KitaMasterdataStore defines the interface for KitaMasterdata data operations.
type KitaMasterdataStore interface {
	Get() (*models.KitaMasterdata, error)
	Update(data *models.KitaMasterdata) error
}

// SQLKitaMasterdataStore implements KitaMasterdataStore using database/sql.
type SQLKitaMasterdataStore struct {
	db *sql.DB
}

// NewSQLKitaMasterdataStore creates a new SQLKitaMasterdataStore.
func NewSQLKitaMasterdataStore(db *sql.DB) *SQLKitaMasterdataStore {
	return &SQLKitaMasterdataStore{db: db}
}

// Get fetches the master data from the database.
func (s *SQLKitaMasterdataStore) Get() (*models.KitaMasterdata, error) {
	query := `SELECT name, street, house_number, postal_code, city, phone_number, email, created_at, updated_at FROM kita_masterdata LIMIT 1`
	row := s.db.QueryRow(query)

	masterdata := &models.KitaMasterdata{}
	err := row.Scan(
		&masterdata.Name,
		&masterdata.Street,
		&masterdata.HouseNumber,
		&masterdata.PostalCode,
		&masterdata.City,
		&masterdata.PhoneNumber,
		&masterdata.Email,
		&masterdata.CreatedAt,
		&masterdata.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return masterdata, nil
}

// Update updates the master data. If no record exists, it creates one.
func (s *SQLKitaMasterdataStore) Update(data *models.KitaMasterdata) error {
	// First, try to update
	queryUpdate := `UPDATE kita_masterdata SET name = ?, street = ?, house_number = ?, postal_code = ?, city = ?, phone_number = ?, email = ?`
	result, err := s.db.Exec(queryUpdate, data.Name, data.Street, data.HouseNumber, data.PostalCode, data.City, data.PhoneNumber, data.Email)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// If no rows affected, insert
		queryInsert := `INSERT INTO kita_masterdata (name, street, house_number, postal_code, city, phone_number, email) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err := s.db.Exec(queryInsert, data.Name, data.Street, data.HouseNumber, data.PostalCode, data.City, data.PhoneNumber, data.Email)
		if err != nil {
			return err
		}
	}
	return nil
}
