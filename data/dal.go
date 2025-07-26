package data

import (
	"database/sql"
	"os"
)

// DAL represents the Data Access Layer.
type DAL struct {
	Users                UserStore
	Children             ChildStore
	Teachers             TeacherStore
	Categories           CategoryStore
	Assignments          AssignmentStore
	DocumentationEntries DocumentationEntryStore
	AudioRecordings      AudioRecordingStore
}

// NewDAL creates a new DAL instance.
func NewDAL(db *sql.DB) *DAL {
	return &DAL{
		Users:                NewSQLUserStore(db),
		Children:             NewSQLChildStore(db),
		Teachers:             NewSQLTeacherStore(db),
		Categories:           NewSQLCategoryStore(db),
		Assignments:          NewSQLAssignmentStore(db),
		DocumentationEntries: NewSQLDocumentationEntryStore(db),
		AudioRecordings:      NewSQLAudioRecordingStore(db),
	}
}

// ReadSQLSchema reads the content of an SQL schema file.
func ReadSQLSchema(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
