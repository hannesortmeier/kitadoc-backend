package data

import (
	"database/sql"
	"errors"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"
)

// AssignmentStore defines the interface for Assignment data operations.
type AssignmentStore interface {
	Create(assignment *models.Assignment) (int, error)
	GetByID(id int) (*models.Assignment, error)
	Update(assignment *models.Assignment) error
	Delete(id int) error
	GetAssignmentHistoryForChild(childID int) ([]models.Assignment, error)
	EndAssignment(assignmentID int) error
}

// Update updates an existing assignment in the database.
func (s *SQLAssignmentStore) Update(assignment *models.Assignment) error {
	query := `UPDATE child_teacher_assignments SET child_id = ?, teacher_id = ?, start_date = ?, end_date = ?, updated_at = ? WHERE assignment_id = ?`
	result, err := s.db.Exec(query, assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate, assignment.UpdatedAt, assignment.ID)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error updating assignment: %v", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting rows affected for assignment ID %d: %v", assignment.ID, err)
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SQLAssignmentStore implements AssignmentStore using database/sql.
type SQLAssignmentStore struct {
	db *sql.DB
}

// NewSQLAssignmentStore creates a new SQLAssignmentStore.
func NewSQLAssignmentStore(db *sql.DB) *SQLAssignmentStore {
	return &SQLAssignmentStore{db: db}
}

// Create inserts a new assignment into the database.
func (s *SQLAssignmentStore) Create(assignment *models.Assignment) (int, error) {
	query := `INSERT INTO child_teacher_assignments (child_id, teacher_id, start_date, end_date) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error inserting assignment: %v", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting last insert ID: %v", err)
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches an assignment by ID from the database.
func (s *SQLAssignmentStore) GetByID(id int) (*models.Assignment, error) {
	query := `SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE assignment_id = ?`
	row := s.db.QueryRow(query, id)
	assignment := &models.Assignment{}
	err := row.Scan(&assignment.ID, &assignment.ChildID, &assignment.TeacherID, &assignment.StartDate, &assignment.EndDate, &assignment.CreatedAt, &assignment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		logger.GetGlobalLogger().Errorf("Error fetching assignment by ID %d: %v", id, err)
		return nil, err
	}
	return assignment, nil
}

// Delete deletes an assignment by ID from the database.
func (s *SQLAssignmentStore) Delete(id int) error {
	query := `DELETE FROM child_teacher_assignments WHERE assignment_id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error deleting assignment with ID %d: %v", id, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting rows affected for assignment ID %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAssignmentHistoryForChild fetches all assignments for a specific child.
func (s *SQLAssignmentStore) GetAssignmentHistoryForChild(childID int) ([]models.Assignment, error) {
	query := `SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE child_id = ? ORDER BY start_date DESC`
	rows, err := s.db.Query(query, childID)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error fetching assignment history for child ID %d: %v", childID, err)
		return nil, err
	}
	defer rows.Close()

	var assignments []models.Assignment
	for rows.Next() {
		assignment := &models.Assignment{}
		err := rows.Scan(&assignment.ID, &assignment.ChildID, &assignment.TeacherID, &assignment.StartDate, &assignment.EndDate, &assignment.CreatedAt, &assignment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, *assignment)
	}

	if err = rows.Err(); err != nil {
		logger.GetGlobalLogger().Errorf("Error iterating over assignment history for child ID %d: %v", childID, err)
		return nil, err
	}

	return assignments, nil
}

// EndAssignment sets the end_date for an assignment to the current time.
func (s *SQLAssignmentStore) EndAssignment(assignmentID int) error {
	query := `UPDATE assignments SET end_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE assignment_id = ? AND end_date IS NULL`
	result, err := s.db.Exec(query, assignmentID)
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error ending assignment with ID %d: %v", assignmentID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.GetGlobalLogger().Errorf("Error getting rows affected for ending assignment ID %d: %v", assignmentID, err)
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
