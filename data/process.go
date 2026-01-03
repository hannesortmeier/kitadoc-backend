package data

import (
	"database/sql"
	"errors"

	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

// ProcessStore defines the interface for Process data operations.
type ProcessStore interface {
	Create(process *models.Process) (*models.Process, error)
	GetByID(id int) (*models.Process, error)
	Update(process *models.Process) error
	Delete(id int) error
}

// SQLProcessStore implements ProcessStore using database/sql.
type SQLProcessStore struct {
	db *sql.DB
}

// NewSQLProcessStore creates a new SQLProcessStore.
func NewSQLProcessStore(db *sql.DB) *SQLProcessStore {
	return &SQLProcessStore{db: db}
}

// Creates a new process. Returns the new newly created process.
func (s *SQLProcessStore) Create(process *models.Process) (*models.Process, error) {
	query := `INSERT INTO processes (status) VALUES (?)`
	result, err := s.db.Exec(query, process.Status)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error creating process: %v", err)
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting last insert ID: %v", err)
		return nil, err
	}
	process.ProcessId = int(id)
	return process, nil
}

// GetByID fetches a process by ID from the database.
func (s *SQLProcessStore) GetByID(id int) (*models.Process, error) {
	query := `SELECT process_id, status FROM processes WHERE process_id = ?`
	row := s.db.QueryRow(query, id)
	process := &models.Process{}
	err := row.Scan(&process.ProcessId, &process.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.GetGlobalLogger().Errorf("Process not found: %d", id)
			return nil, ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Error fetching process: %v", err)
		return nil, err
	}

	return process, nil
}

// Update updates an existing process in the database.
func (s *SQLProcessStore) Update(process *models.Process) error {
	query := `UPDATE processes SET status = ? WHERE process_id = ?`
	result, err := s.db.Exec(query, process.Status, process.ProcessId)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error updating process: %v", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error updating process: %v", err)
		return err
	}
	if rowsAffected == 0 {
		logger.GetGlobalLogger().Errorf("Process not found: %d", process.ProcessId)
		return ErrNotFound
	}
	return nil
}

// Delete deletes a process by ID from the database.
func (s *SQLProcessStore) Delete(id int) error {
	query := `DELETE FROM processes WHERE process_id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error deleting process: %v", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error deleting process: %v", err)
		return err
	}
	if rowsAffected == 0 {
		logger.GetGlobalLogger().Errorf("Process not found: %d", id)
		return ErrNotFound
	}
	return nil
}
