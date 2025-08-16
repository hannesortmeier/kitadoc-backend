package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSQLAssignmentStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	assignment := &models.Assignment{
		ChildID:   1,
		TeacherID: 2,
		StartDate: time.Now(),
		EndDate:   nil,
	}

	log_level, _ := logrus.ParseLevel("debug")

	logger.InitGlobalLogger(
		log_level,
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO child_teacher_assignments (child_id, teacher_id, start_date, end_date) VALUES (?, ?, ?, ?)`)).
			WithArgs(assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(assignment)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO child_teacher_assignments (child_id, teacher_id, start_date, end_date) VALUES (?, ?, ?, ?)`)).
			WithArgs(assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(assignment)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	assignmentID := 1
	expectedAssignment := &models.Assignment{
		ID:        assignmentID,
		ChildID:   1,
		TeacherID: 2,
		StartDate: time.Now().Truncate(time.Second),
		EndDate:   nil,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"assignment_id", "child_id", "teacher_id", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(expectedAssignment.ID, expectedAssignment.ChildID, expectedAssignment.TeacherID, expectedAssignment.StartDate, expectedAssignment.EndDate, expectedAssignment.CreatedAt, expectedAssignment.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnRows(rows)

		assignment, err := store.GetByID(assignmentID)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)
		assert.Equal(t, expectedAssignment.ID, assignment.ID)
		assert.Equal(t, expectedAssignment.ChildID, assignment.ChildID)
		assert.Equal(t, expectedAssignment.TeacherID, assignment.TeacherID)
		assert.WithinDuration(t, expectedAssignment.StartDate, assignment.StartDate, time.Second)
		assert.Equal(t, expectedAssignment.EndDate, assignment.EndDate)
		assert.WithinDuration(t, expectedAssignment.CreatedAt, assignment.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedAssignment.UpdatedAt, assignment.UpdatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnError(sql.ErrNoRows)

		assignment, err := store.GetByID(assignmentID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, assignment)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnError(errors.New("db error"))

		assignment, err := store.GetByID(assignmentID)
		assert.Error(t, err)
		assert.Nil(t, assignment)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	assignment := &models.Assignment{
		ID:        1,
		ChildID:   1,
		TeacherID: 2,
		StartDate: time.Now().Add(-time.Hour),
		EndDate:   func() *time.Time { t := time.Now(); return &t }(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE child_teacher_assignments SET child_id = ?, teacher_id = ?, start_date = ?, end_date = ?, updated_at = ? WHERE assignment_id = ?`)).
			WithArgs(assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate, assignment.UpdatedAt, assignment.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Update(assignment)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE child_teacher_assignments SET child_id = ?, teacher_id = ?, start_date = ?, end_date = ?, updated_at = ? WHERE assignment_id = ?`)).
			WithArgs(assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate, assignment.UpdatedAt, assignment.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Update(assignment)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE child_teacher_assignments SET child_id = ?, teacher_id = ?, start_date = ?, end_date = ?, updated_at = ? WHERE assignment_id = ?`)).
			WithArgs(assignment.ChildID, assignment.TeacherID, assignment.StartDate, assignment.EndDate, assignment.UpdatedAt, assignment.ID).
			WillReturnError(errors.New("db error"))

		err := store.Update(assignment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	assignmentID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(assignmentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(assignmentID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM child_teacher_assignments WHERE assignment_id = ?`)).
			WithArgs(assignmentID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(assignmentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_GetAssignmentHistoryForChild(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	childID := 1
	now := time.Now().Truncate(time.Second)
	assignments := []models.Assignment{
		{ID: 1, ChildID: childID, TeacherID: 1, StartDate: now.Add(-time.Hour * 24), EndDate: &now, CreatedAt: now.Add(-time.Hour * 25), UpdatedAt: now},
		{ID: 2, ChildID: childID, TeacherID: 2, StartDate: now.Add(-time.Hour * 48), EndDate: nil, CreatedAt: now.Add(-time.Hour * 49), UpdatedAt: now.Add(-time.Hour * 49)},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "child_id", "teacher_id", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(assignments[0].ID, assignments[0].ChildID, assignments[0].TeacherID, assignments[0].StartDate, assignments[0].EndDate, assignments[0].CreatedAt, assignments[0].UpdatedAt).
			AddRow(assignments[1].ID, assignments[1].ChildID, assignments[1].TeacherID, assignments[1].StartDate, assignments[1].EndDate, assignments[1].CreatedAt, assignments[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE child_id = ? ORDER BY start_date DESC`)).
			WithArgs(childID).
			WillReturnRows(rows)

		fetchedAssignments, err := store.GetAssignmentHistoryForChild(childID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedAssignments)
		assert.Len(t, fetchedAssignments, 2)
		assert.Equal(t, assignments[0].ID, fetchedAssignments[0].ID)
		assert.Equal(t, assignments[1].ID, fetchedAssignments[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no assignments found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE child_id = ? ORDER BY start_date DESC`)).
			WithArgs(childID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "child_id", "teacher_id", "start_date", "end_date", "created_at", "updated_at"}))

		fetchedAssignments, err := store.GetAssignmentHistoryForChild(childID)
		assert.NoError(t, err)
		assert.Nil(t, fetchedAssignments)
		assert.Len(t, fetchedAssignments, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE child_id = ? ORDER BY start_date DESC`)).
			WithArgs(childID).
			WillReturnError(errors.New("db error"))

		fetchedAssignments, err := store.GetAssignmentHistoryForChild(childID)
		assert.Error(t, err)
		assert.Nil(t, fetchedAssignments)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"assignment_id", "child_id", "teacher_id", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(assignments[0].ID, assignments[0].ChildID, "not-an-int", assignments[0].StartDate, assignments[0].EndDate, assignments[0].CreatedAt, assignments[0].UpdatedAt) // Malformed row

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments WHERE child_id = ? ORDER BY start_date DESC`)).
			WithArgs(childID).
			WillReturnRows(rows)

		fetchedAssignments, err := store.GetAssignmentHistoryForChild(childID)
		assert.Error(t, err)
		assert.Nil(t, fetchedAssignments)
		assert.Contains(t, err.Error(), "converting driver.Value type string")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_EndAssignment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	assignmentID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE assignments SET end_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE assignment_id = ? AND end_date IS NULL`)).
			WithArgs(assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.EndAssignment(assignmentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or already ended", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE assignments SET end_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE assignment_id = ? AND end_date IS NULL`)).
			WithArgs(assignmentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.EndAssignment(assignmentID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE assignments SET end_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE assignment_id = ? AND end_date IS NULL`)).
			WithArgs(assignmentID).
			WillReturnError(errors.New("db error"))

		err := store.EndAssignment(assignmentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAssignmentStore_GetAllAssignments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAssignmentStore(db)

	now := time.Now().Truncate(time.Second)
	assignments := []models.Assignment{
		{ID: 1, ChildID: 1, TeacherID: 1, StartDate: now.Add(-time.Hour * 24), EndDate: &now, CreatedAt: now.Add(-time.Hour * 25), UpdatedAt: now},
		{ID: 2, ChildID: 2, TeacherID: 2, StartDate: now.Add(-time.Hour * 48), EndDate: nil, CreatedAt: now.Add(-time.Hour * 49), UpdatedAt: now.Add(-time.Hour * 49)},
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"assignment_id", "child_id", "teacher_id", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(assignments[0].ID, assignments[0].ChildID, assignments[0].TeacherID, assignments[0].StartDate, assignments[0].EndDate, assignments[0].CreatedAt, assignments[0].UpdatedAt).
			AddRow(assignments[1].ID, assignments[1].ChildID, assignments[1].TeacherID, assignments[1].StartDate, assignments[1].EndDate, assignments[1].CreatedAt, assignments[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments ORDER BY start_date DESC`)).
			WillReturnRows(rows)

		fetchedAssignments, err := store.GetAllAssignments()
		assert.NoError(t, err)
		assert.NotNil(t, fetchedAssignments)
		assert.Len(t, fetchedAssignments, 2)
		assert.Equal(t, assignments[0].ID, fetchedAssignments[0].ID)
		assert.Equal(t, assignments[1].ID, fetchedAssignments[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT assignment_id, child_id, teacher_id, start_date, end_date, created_at, updated_at FROM child_teacher_assignments ORDER BY start_date DESC`)).
			WillReturnError(errors.New("db error"))

		fetchedAssignments, err := store.GetAllAssignments()
		assert.Error(t, err)
		assert.Nil(t, fetchedAssignments)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
