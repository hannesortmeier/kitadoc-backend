package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"kitadoc-backend/data"
	"kitadoc-backend/models"

	"github.com/go-playground/validator/v10"
)

// DocumentationEntryService defines the interface for documentation entry-related business logic operations.
type DocumentationEntryService interface {
	CreateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) (*models.DocumentationEntry, error)
	GetDocumentationEntryByID(logger *logrus.Entry, ctx context.Context, id int) (*models.DocumentationEntry, error)
	UpdateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) error
	DeleteDocumentationEntry(logger *logrus.Entry, ctx context.Context, id int) error
	GetAllDocumentationForChild(logger *logrus.Entry, ctx context.Context, childID int) ([]models.DocumentationEntry, error)
	ApproveDocumentationEntry(logger *logrus.Entry, ctx context.Context, entryID int, approvedByUserID int) error
	GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int) ([]byte, error) // Returns a byte slice representing the Word document
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
	validate.RegisterValidation("iso8601date", models.ValidateISO8601Date) //nolint:errcheck
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
func (service *DocumentationEntryServiceImpl) CreateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) (*models.DocumentationEntry, error) {
	if err := service.validate.Struct(entry); err != nil {
		logger.WithError(err).Error("Invalid input for CreateDocumentationEntry")
		return nil, ErrInvalidInput
	}

	// Validate ChildID
	_, err := service.childStore.GetByID(entry.ChildID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("child_id", entry.ChildID).Warn("Child not found for documentation entry creation")
			return nil, errors.New("child not found")
		}
		logger.WithError(err).WithField("child_id", entry.ChildID).Error("Error fetching child by ID for documentation entry creation")
		return nil, ErrInternal
	}

	// Validate TeacherID
	_, err = service.teacherStore.GetByID(entry.TeacherID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("teacher_id", entry.TeacherID).Warn("Teacher not found for documentation entry creation")
			return nil, errors.New("teacher not found")
		}
		logger.WithError(err).WithField("teacher_id", entry.TeacherID).Error("Error fetching teacher by ID for documentation entry creation")
		return nil, ErrInternal
	}

	// Validate CategoryID
	_, err = service.categoryStore.GetByID(entry.CategoryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("category_id", entry.CategoryID).Warn("Category not found for documentation entry creation")
			return nil, errors.New("category not found")
		}
		logger.WithError(err).WithField("category_id", entry.CategoryID).Error("Error fetching category by ID for documentation entry creation")
		return nil, ErrInternal
	}

	// Business rule: EntryDate cannot be in the future.
	if entry.ObservationDate.After(time.Now()) {
		logger.WithField("observation_date", entry.ObservationDate).Warn("Observation date cannot be in the future")
		return nil, errors.New("observation date cannot be in the future")
	}

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	id, err := service.documentationEntryStore.Create(entry)
	if err != nil {
		logger.WithError(err).Error("Error creating documentation entry in store")
		return nil, ErrInternal
	}
	entry.ID = id
	logger.WithField("entry_id", entry.ID).Info("Documentation entry created successfully")
	return entry, nil
}

// GetDocumentationEntryByID fetches a documentation entry by ID.
func (service *DocumentationEntryServiceImpl) GetDocumentationEntryByID(logger *logrus.Entry, ctx context.Context, id int) (*models.DocumentationEntry, error) {
	entry, err := service.documentationEntryStore.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("entry_id", id).Warn("Documentation entry not found")
			return nil, ErrNotFound
		}
		logger.WithError(err).WithField("entry_id", id).Error("Error fetching documentation entry by ID")
		return nil, ErrInternal
	}
	logger.WithField("entry_id", id).Info("Documentation entry fetched successfully")
	return entry, nil
}

// UpdateDocumentationEntry updates an existing documentation entry.
func (service *DocumentationEntryServiceImpl) UpdateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) error {
	if err := service.validate.Struct(entry); err != nil {
		logger.WithError(err).Warn("Invalid input for UpdateDocumentationEntry")
		return ErrInvalidInput
	}

	// Validate ChildID
	_, err := service.childStore.GetByID(entry.ChildID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("child_id", entry.ChildID).Warn("Child not found for documentation entry update")
			return errors.New("child not found")
		}
		logger.WithError(err).WithField("child_id", entry.ChildID).Error("Error fetching child by ID for documentation entry update")
		return ErrInternal
	}

	// Validate TeacherID
	_, err = service.teacherStore.GetByID(entry.TeacherID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("teacher_id", entry.TeacherID).Warn("Teacher not found for documentation entry update")
			return errors.New("teacher not found")
		}
		logger.WithError(err).WithField("teacher_id", entry.TeacherID).Error("Error fetching teacher by ID for documentation entry update")
		return ErrInternal
	}

	// Validate CategoryID
	_, err = service.categoryStore.GetByID(entry.CategoryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("category_id", entry.CategoryID).Warn("Category not found for documentation entry update")
			return errors.New("category not found")
		}
		logger.WithError(err).WithField("category_id", entry.CategoryID).Error("Error fetching category by ID for documentation entry update")
		return ErrInternal
	}

	// Business rule: EntryDate cannot be in the future.
	if entry.ObservationDate.After(time.Now()) {
		logger.WithField("observation_date", entry.ObservationDate).Warn("Observation date cannot be in the future for update")
		return errors.New("entry date cannot be in the future")
	}

	entry.UpdatedAt = time.Now()
	err = service.documentationEntryStore.Update(entry)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("entry_id", entry.ID).Warn("Documentation entry not found for update")
			return ErrNotFound
		}
		logger.WithError(err).WithField("entry_id", entry.ID).Error("Error updating documentation entry in store")
		return ErrInternal
	}
	logger.WithField("entry_id", entry.ID).Info("Documentation entry updated successfully")
	return nil
}

// DeleteDocumentationEntry deletes a documentation entry by ID.
func (service *DocumentationEntryServiceImpl) DeleteDocumentationEntry(logger *logrus.Entry, ctx context.Context, id int) error {
	err := service.documentationEntryStore.Delete(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("entry_id", id).Warn("Documentation entry not found for deletion")
			return ErrNotFound
		}
		logger.WithError(err).WithField("entry_id", id).Error("Error deleting documentation entry from store")
		return ErrInternal
	}
	logger.WithField("entry_id", id).Info("Documentation entry deleted successfully")
	return nil
}

// GetAllDocumentationForChild fetches all documentation entries for a specific child.
func (service *DocumentationEntryServiceImpl) GetAllDocumentationForChild(logger *logrus.Entry, ctx context.Context, childID int) ([]models.DocumentationEntry, error) {
	// Validate ChildID
	_, err := service.childStore.GetByID(childID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("child_id", childID).Warn("Child not found for fetching documentation entries")
			return nil, errors.New("child not found")
		}
		logger.WithError(err).WithField("child_id", childID).Error("Error fetching child by ID for documentation entries")
		return nil, ErrInternal
	}

	entries, err := service.documentationEntryStore.GetAllForChild(childID)
	if err != nil {
		logger.WithError(err).WithField("child_id", childID).Error("Error fetching documentation entries for child ID")
		return nil, ErrInternal
	}
	logger.WithField("child_id", childID).Info("Documentation entries fetched successfully for child")
	return entries, nil
}

// ApproveDocumentationEntry approves a documentation entry.
func (service *DocumentationEntryServiceImpl) ApproveDocumentationEntry(logger *logrus.Entry, ctx context.Context, entryID int, approvedByUserID int) error {
	// Check if the entry exists
	entry, err := service.documentationEntryStore.GetByID(entryID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("entry_id", entryID).Warn("Documentation entry not found for approval")
			return ErrNotFound
		}
		logger.WithError(err).WithField("entry_id", entryID).Error("Error fetching documentation entry by ID for approval")
		return ErrInternal
	}

	// Check if the approving user exists
	_, err = service.userStore.GetByID(approvedByUserID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("user_id", approvedByUserID).Warn("Approving user not found")
			return errors.New("approving user not found")
		}
		logger.WithError(err).WithField("user_id", approvedByUserID).Error("Error fetching user by ID for approval")
		return ErrInternal
	}

	// Business rule: Only unapproved entries can be approved.
	if entry.IsApproved {
		logger.WithField("entry_id", entryID).Warn("Documentation entry is already approved")
		return errors.New("documentation entry is already approved")
	}

	err = service.documentationEntryStore.ApproveEntry(entryID, approvedByUserID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("entry_id", entryID).Warn("Documentation entry not found during approval process")
			return ErrNotFound
		}
		logger.WithError(err).WithField("entry_id", entryID).Error("Error approving documentation entry in store")
		return ErrInternal
	}
	logger.WithField("entry_id", entryID).Info("Documentation entry approved successfully")
	return nil
}

// GenerateChildReport generates a Word document report for a child.
// This is a placeholder for actual document generation logic.
func (service *DocumentationEntryServiceImpl) GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int) ([]byte, error) {
	logger.WithField("child_id", childID).Info("Attempting to generate child report")

	// In a real implementation, you would fetch child data and related documentation entries,
	// then use a library (e.g., go-docx, unioffice) to generate a Word document.
	// For now, we return a placeholder error and an empty byte slice.
	// Simulate some work
	select {
	case <-ctx.Done():
		logger.Warn("Child report generation cancelled due to context cancellation")
		return nil, ctx.Err()
	case <-time.After(2 * time.Second): // Simulate document generation time
		// Fetch child data and documentation entries
		child, err := service.childStore.GetByID(childID)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				logger.WithField("child_id", childID).Warn("Child not found for report generation")
				return nil, errors.New("child not found for report generation")
			}
			logger.WithError(err).WithField("child_id", childID).Error("Error fetching child for report generation")
			return nil, ErrInternal
		}

		entries, err := service.documentationEntryStore.GetAllForChild(childID)
		if err != nil {
			logger.WithError(err).WithField("child_id", childID).Error("Error fetching documentation entries for report generation")
			return nil, ErrInternal
		}

		// Placeholder for actual document generation logic
		// For demonstration, create a simple text document
		content := fmt.Sprintf("Child Report for: %s %s\n\n", child.FirstName, child.LastName)
		if len(entries) > 0 {
			content += "Documentation Entries:\n"
			for _, entry := range entries {
				content += fmt.Sprintf("- Date: %s, Category: %d, Observation: %s\n", entry.ObservationDate.Format("2006-01-02"), entry.CategoryID, entry.ObservationDescription)
			}
		} else {
			content += "No documentation entries found for this child.\n"
		}

		// Simulate DOCX content (very basic, not a real DOCX)
		// A real DOCX would require a proper library like github.com/unidoc/unioffice
		docxContent := []byte(content)
		logger.WithField("child_id", childID).Info("Child report generated successfully (simulated)")
		return docxContent, nil
	}
}

// DownloadDocument is no longer needed as reports are directly returned.
// This function has been removed as per the scope.
