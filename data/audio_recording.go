package data

import (
	"database/sql"
	"errors" // Import the errors package

	"kitadoc-backend/models"
)

// AudioRecordingStore defines the interface for AudioRecording data operations.
type AudioRecordingStore interface {
	Create(recording *models.AudioRecording) (int, error)
	GetByID(id int) (*models.AudioRecording, error)
	Delete(id int) error
}

// SQLAudioRecordingStore implements AudioRecordingStore using database/sql.
type SQLAudioRecordingStore struct {
	db *sql.DB
}

// NewSQLAudioRecordingStore creates a new SQLAudioRecordingStore.
func NewSQLAudioRecordingStore(db *sql.DB) *SQLAudioRecordingStore {
	return &SQLAudioRecordingStore{db: db}
}

// Create inserts a new audio recording into the database.
func (store *SQLAudioRecordingStore) Create(recording *models.AudioRecording) (int, error) {
	query := `INSERT INTO audio_recordings (documentation_entry_id, file_path, duration_seconds, created_at) VALUES (?, ?, ?, ?)`
	result, err := store.db.Exec(query, recording.DocumentationEntryID, recording.FilePath, recording.DurationSeconds, recording.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID fetches an audio recording by ID from the database.
func (store *SQLAudioRecordingStore) GetByID(id int) (*models.AudioRecording, error) {
	query := `SELECT id, documentation_entry_id, file_path, duration_seconds, created_at FROM audio_recordings WHERE id = ?`
	row := store.db.QueryRow(query, id)
	recording := &models.AudioRecording{}
	err := row.Scan(&recording.ID, &recording.DocumentationEntryID, &recording.FilePath, &recording.DurationSeconds, &recording.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return recording, nil
}

// Delete deletes an audio recording by ID from the database.
func (store *SQLAudioRecordingStore) Delete(id int) error {
	query := `DELETE FROM audio_recordings WHERE id = ?`
	result, err := store.db.Exec(query, id)
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