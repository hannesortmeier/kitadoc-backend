package services

import (
	"errors"
	"log"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// DocumentationEntryService defines the interface for documentation entry-related business logic operations.
type DocumentationEntryService interface {
	CreateDocumentationEntry(entry *models.DocumentationEntry) (*models.DocumentationEntry, error)
	GetDocumentationEntryByID(id int) (*models.DocumentationEntry, error)
	UpdateDocumentationEntry(entry *models.DocumentationEntry) error
	DeleteDocumentationEntry(id int) error
	GetAllDocumentationForChild(childID int) ([]models.DocumentationEntry, error)
	ApproveDocumentationEntry(entryID int, approvedByUserID int) error
	GenerateChildReport(childID int) ([]byte, error) // Returns a byte slice representing the Word document
	DownloadDocument(documentID int) ([]byte, error) // Placeholder for document download
}

// DocumentationEntryServiceImpl implements DocumentationEntryService.
type DocumentationEntryServiceImpl struct {
	documentationEntryStore data.DocumentationEntryStore
	childStore              data.ChildStore
	teacherStore            data.TeacherStore
	categoryStore           data.CategoryStore
	userStore               data.UserStore // For ApprovedByUserID validation
	validate                *validator.Validate
}

// NewDocumentationEntryService creates a new DocumentationEntryServiceImpl.
func NewDocumentationEntryService(
	documentationEntryStore data.DocumentationEntryStore,
	childStore data.ChildStore,
	teacherStore data.TeacherStore,
	categoryStore data.CategoryStore,
	userStore data.UserStore,
) *DocumentationEntryServiceImpl {
	validate := validator.New()
	validate.RegisterValidation("iso8601date", models.ValidateISO8601Date)
	return &DocumentationEntryServiceImpl{
		documentationEntryStore: documentationEntryStore,
		childStore:              childStore,
		teacherStore:            teacherStore,
		categoryStore:           categoryStore,
		userStore:               userStore,
		validate:                validate,
	}
}

// CreateDocumentationEntry creates a new documentation entry.
func (service *DocumentationEntryServiceImpl) CreateDocumentationEntry(entry *models.DocumentationEntry) (*models.DocumentationEntry, error) {
	if err := service.validate.Struct(entry); err != nil {
		return nil, ErrInvalidInput
	}

	// Validate ChildID
	_, err := service.childStore.GetByID(entry.ChildID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("child not found")
		}
		log.Printf("Error fetching child by ID %d: %v", entry.ChildID, err)
		return nil, ErrInternal
	}

	// Validate TeacherID
	_, err = service.teacherStore.GetByID(entry.TeacherID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("teacher not found")
		}
		log.Printf("Error fetching teacher by ID %d: %v", entry.TeacherID, err)
		return nil, ErrInternal
	}

	// Validate CategoryID
	_, err = service.categoryStore.GetByID(entry.CategoryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("category not found")
		}
		log.Printf("Error fetching category by ID %d: %v", entry.CategoryID, err)
		return nil, ErrInternal
	}

	// Business rule: EntryDate cannot be in the future.
	if entry.ObservationDate.After(time.Now()) {
		return nil, errors.New("observation date cannot be in the future")
	}

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	id, err := service.documentationEntryStore.Create(entry)
	if err != nil {
		log.Printf("Error creating documentation entry: %v", err)
		return nil, ErrInternal
	}
	entry.ID = id
	return entry, nil
}

// GetDocumentationEntryByID fetches a documentation entry by ID.
func (service *DocumentationEntryServiceImpl) GetDocumentationEntryByID(id int) (*models.DocumentationEntry, error) {
	entry, err := service.documentationEntryStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, ErrNotFound
		}
		log.Printf("Error fetching documentation entry by ID %d: %v", id, err)
		return nil, ErrInternal
	}
	return entry, nil
}

// UpdateDocumentationEntry updates an existing documentation entry.
func (service *DocumentationEntryServiceImpl) UpdateDocumentationEntry(entry *models.DocumentationEntry) error {
	if err := service.validate.Struct(entry); err != nil {
		return ErrInvalidInput
	}

	// Validate ChildID
	_, err := service.childStore.GetByID(entry.ChildID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return errors.New("child not found")
		}
		log.Printf("Error fetching child by ID %d: %v", entry.ChildID, err)
		return ErrInternal
	}

	// Validate TeacherID
	_, err = service.teacherStore.GetByID(entry.TeacherID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return errors.New("teacher not found")
		}
		log.Printf("Error fetching teacher by ID %d: %v", entry.TeacherID, err)
		return ErrInternal
	}

	// Validate CategoryID
	_, err = service.categoryStore.GetByID(entry.CategoryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return errors.New("category not found")
		}
		log.Printf("Error fetching category by ID %d: %v", entry.CategoryID, err)
		return ErrInternal
	}

	// Business rule: EntryDate cannot be in the future.
	if entry.ObservationDate.After(time.Now()) {
		return errors.New("entry date cannot be in the future")
	}

	entry.UpdatedAt = time.Now()
	err = service.documentationEntryStore.Update(entry)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		log.Printf("Error updating documentation entry with ID %d: %v", entry.ID, err)
		return ErrInternal
	}
	return nil
}

// DeleteDocumentationEntry deletes a documentation entry by ID.
func (service *DocumentationEntryServiceImpl) DeleteDocumentationEntry(id int) error {
	err := service.documentationEntryStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		log.Printf("Error deleting documentation entry with ID %d: %v", id, err)
		return ErrInternal
	}
	return nil
}

// GetAllDocumentationForChild fetches all documentation entries for a specific child.
func (service *DocumentationEntryServiceImpl) GetAllDocumentationForChild(childID int) ([]models.DocumentationEntry, error) {
	// Validate ChildID
	_, err := service.childStore.GetByID(childID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return nil, errors.New("child not found")
		}
		log.Printf("Error fetching child by ID %d: %v", childID, err)
		return nil, ErrInternal
	}

	entries, err := service.documentationEntryStore.GetAllForChild(childID)
	if err != nil {
		log.Printf("Error fetching documentation entries for child ID %d: %v", childID, err)
		return nil, ErrInternal
	}
	return entries, nil
}

// ApproveDocumentationEntry approves a documentation entry.
func (service *DocumentationEntryServiceImpl) ApproveDocumentationEntry(entryID int, approvedByUserID int) error {
	// Check if the entry exists
	entry, err := service.documentationEntryStore.GetByID(entryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		log.Printf("Error fetching documentation entry by ID %d: %v", entryID, err)
		return ErrInternal
	}

	// Check if the approving user exists
	_, err = service.userStore.GetByID(approvedByUserID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return errors.New("approving user not found")
		}
		log.Printf("Error fetching user by ID %d: %v", approvedByUserID, err)
		return ErrInternal
	}

	// Business rule: Only unapproved entries can be approved.
	if entry.IsApproved {
		return errors.New("documentation entry is already approved")
	}

	err = service.documentationEntryStore.ApproveEntry(entryID, approvedByUserID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return ErrNotFound
		}
		log.Printf("Error approving documentation entry with ID %d: %v", entryID, err)
		return ErrInternal
	}
	return nil
}

// GenerateChildReport generates a Word document report for a child.
// This is a placeholder for actual document generation logic.
func (service *DocumentationEntryServiceImpl) GenerateChildReport(childID int) ([]byte, error) {
	// In a real implementation, you would fetch child data and related documentation entries,
	// then use a library (e.g., go-docx, unioffice) to generate a Word document.
	// For now, we return a placeholder error and an empty byte slice.
	_ = childID // Suppress unused variable warning
	return nil, ErrChildReportGenerationFailed
}

// DownloadDocument is a placeholder for downloading a generated document.
func (service *DocumentationEntryServiceImpl) DownloadDocument(documentID int) ([]byte, error) {
	// In a real implementation, you would retrieve the document content based on documentID
	// (e.g., from a file storage or database) and return it.
	_ = documentID // Suppress unused variable warning
	return nil, errors.New("document download not implemented")
}