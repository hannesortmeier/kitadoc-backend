package data_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSQLAudioRecordingStore_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAudioRecordingStore(db)

	recording := &models.AudioRecording{
		DocumentationEntryID: 1,
		FilePath:             "/path/to/audio.mp3",
		DurationSeconds:      120,
		CreatedAt:            time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audio_recordings (documentation_entry_id, file_path, duration_seconds, created_at) VALUES (?, ?, ?, ?)`)).
			WithArgs(recording.DocumentationEntryID, recording.FilePath, recording.DurationSeconds, recording.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		id, err := store.Create(recording)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO audio_recordings (documentation_entry_id, file_path, duration_seconds, created_at) VALUES (?, ?, ?, ?)`)).
			WithArgs(recording.DocumentationEntryID, recording.FilePath, recording.DurationSeconds, recording.CreatedAt).
			WillReturnError(errors.New("db error"))

		id, err := store.Create(recording)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAudioRecordingStore_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAudioRecordingStore(db)

	recordingID := 1
	expectedRecording := &models.AudioRecording{
		ID:                   recordingID,
		DocumentationEntryID: 1,
		FilePath:             "/path/to/audio.mp3",
		DurationSeconds:      120,
		CreatedAt:            time.Now().Truncate(time.Second),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "documentation_entry_id", "file_path", "duration_seconds", "created_at"}).
			AddRow(expectedRecording.ID, expectedRecording.DocumentationEntryID, expectedRecording.FilePath, expectedRecording.DurationSeconds, expectedRecording.CreatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, documentation_entry_id, file_path, duration_seconds, created_at FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnRows(rows)

		recording, err := store.GetByID(recordingID)
		assert.NoError(t, err)
		assert.NotNil(t, recording)
		assert.Equal(t, expectedRecording.ID, recording.ID)
		assert.Equal(t, expectedRecording.DocumentationEntryID, recording.DocumentationEntryID)
		assert.Equal(t, expectedRecording.FilePath, recording.FilePath)
		assert.Equal(t, expectedRecording.DurationSeconds, recording.DurationSeconds)
		assert.WithinDuration(t, expectedRecording.CreatedAt, recording.CreatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, documentation_entry_id, file_path, duration_seconds, created_at FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnError(sql.ErrNoRows)

		recording, err := store.GetByID(recordingID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.Nil(t, recording)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, documentation_entry_id, file_path, duration_seconds, created_at FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnError(errors.New("db error"))

		recording, err := store.GetByID(recordingID)
		assert.Error(t, err)
		assert.Nil(t, recording)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLAudioRecordingStore_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close() //nolint:errcheck

	store := data.NewSQLAudioRecordingStore(db)

	recordingID := 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Delete(recordingID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(recordingID)
		assert.Error(t, err)
		assert.Equal(t, data.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM audio_recordings WHERE id = ?`)).
			WithArgs(recordingID).
			WillReturnError(errors.New("db error"))

		err := store.Delete(recordingID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
