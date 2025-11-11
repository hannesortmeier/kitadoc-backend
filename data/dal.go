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
}

// NewDAL creates a new DAL instance.
func NewDAL(db *sql.DB, encryptionKey []byte) *DAL {
	return &DAL{
		Users:                NewSQLUserStore(db, encryptionKey),
		Children:             NewSQLChildStore(db, encryptionKey),
		Teachers:             NewSQLTeacherStore(db, encryptionKey),
		Categories:           NewSQLCategoryStore(db),
		Assignments:          NewSQLAssignmentStore(db),
		DocumentationEntries: NewSQLDocumentationEntryStore(db, encryptionKey),
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
