package data

import (
	"database/sql"
	"errors"
	"strings"

	"kitadoc-backend/models"
)

// GroupStore defines the interface for Group data operations.
type GroupStore interface {
	Create(group *models.Group) (int, error)
	GetByID(id int) (*models.Group, error)
	Update(group *models.Group) error
	Delete(id int) error
	GetByName(name string) (*models.Group, error)
	GetAll() ([]models.Group, error)
}

// SQLGroupStore implements GroupStore using database/sql.
type SQLGroupStore struct {
	db *sql.DB
}

// NewSQLGroupStore creates a new SQLGroupStore.
func NewSQLGroupStore(db *sql.DB) *SQLGroupStore {
	return &SQLGroupStore{db: db}
}

// Create inserts a new group into the database.
func (s *SQLGroupStore) Create(group *models.Group) (int, error) {
	query := `INSERT INTO groups (group_name) VALUES (?)`
	result, err := s.db.Exec(query, group.Name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches a group by ID from the database.
func (s *SQLGroupStore) GetByID(id int) (*models.Group, error) {
	query := `SELECT group_id, group_name FROM groups WHERE group_id = ?`
	row := s.db.QueryRow(query, id)
	group := &models.Group{}
	err := row.Scan(&group.ID, &group.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return group, nil
}

// Update updates an existing group in the database.
func (s *SQLGroupStore) Update(group *models.Group) error {
	query := `UPDATE groups SET group_name = ? WHERE group_id = ?`
	result, err := s.db.Exec(query, group.Name, group.ID)
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

// Delete deletes a group by ID from the database.
func (s *SQLGroupStore) Delete(id int) error {
	query := `DELETE FROM groups WHERE group_id = ?`
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

// GetByName fetches a group by name from the database.
func (s *SQLGroupStore) GetByName(name string) (*models.Group, error) {
	query := `SELECT group_id, group_name FROM groups WHERE group_name = ?`
	row := s.db.QueryRow(query, name)
	group := &models.Group{}
	err := row.Scan(&group.ID, &group.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return group, nil
}

// GetAll fetches all groups from the database.
func (s *SQLGroupStore) GetAll() ([]models.Group, error) {
	query := `SELECT group_id, group_name FROM groups`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		group := &models.Group{}
		err := rows.Scan(&group.ID, &group.Name)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}
