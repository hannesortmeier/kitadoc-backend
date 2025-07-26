package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/wml/stypes"
	"github.com/sirupsen/logrus"
	"kitadoc-backend/data"
	"kitadoc-backend/models"
)

// DocumentationEntryService defines the interface for documentation entry-related business logic operations.
type DocumentationEntryService interface {
	CreateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) (*models.DocumentationEntry, error)
	GetDocumentationEntryByID(logger *logrus.Entry, ctx context.Context, id int) (*models.DocumentationEntry, error)
	UpdateDocumentationEntry(logger *logrus.Entry, ctx context.Context, entry *models.DocumentationEntry) error
	DeleteDocumentationEntry(logger *logrus.Entry, ctx context.Context, id int) error
	GetAllDocumentationForChild(logger *logrus.Entry, ctx context.Context, childID int) ([]models.DocumentationEntry, error)
	ApproveDocumentationEntry(logger *logrus.Entry, ctx context.Context, entryID int, approvedByUserID int) error
	GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int, assignments []models.Assignment) ([]byte, error) // Returns a byte slice representing the Word document
	GetDocumentName(ctx context.Context, childID int) (string, error)                                                            // Returns the document name for a child report
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

// GenerateChildReport generates a Word document with the child's documentation entries.
func (service *DocumentationEntryServiceImpl) GenerateChildReport(logger *logrus.Entry, ctx context.Context, childID int, assignments []models.Assignment) ([]byte, error) {
	logger.WithField("child_id", childID).Info("Generating child report")

	child, err := service.childStore.GetByID(childID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			logger.WithField("child_id", childID).Warn("Child not found for report generation")
			return nil, ErrNotFound
		}
		logger.WithError(err).WithField("child_id", childID).Error("Error fetching child for report generation")
		return nil, ErrInternal
	}

	entries, err := service.documentationEntryStore.GetAllForChild(childID)
	if err != nil {
		logger.WithError(err).WithField("child_id", childID).Error("Error fetching documentation entries for report generation")
		return nil, ErrInternal
	}

	document, err := godocx.NewDocument()
	if err != nil {
		logger.WithError(err).Error("Error creating new Word document for child report")
		return nil, ErrChildReportGenerationFailed
	}

	assignmentsText, err := service.FormatChildTeacherAssignments(assignments)
	if err != nil {
		logger.WithError(err).WithField("child_id", childID).Error("Error formatting child teacher assignments for report")
		return nil, ErrChildReportGenerationFailed
	}

	breaktype := stypes.BreakTypeTextWrapping

	// Add a title
	document.AddHeading("Dokumentation", 0)
	document.AddParagraph(
		"des Bildungsprozesses im Rahmen der Grundsätze zur Bildungsförderung für Kinder von 0 bis 10 Jahren in Kindertageseinrichtungen und Schulen im Primarbereich in NRW",
	).Justification(stypes.JustificationCenter)

	document.AddEmptyParagraph()

	// TODO(hannes) This needs to be dynamically set based on the kindergarten's details
	addressParagraph := document.AddEmptyParagraph()
	addressParagraph.AddText("Familienzentrum St. Jakobus Emsdetten").AddBreak(&breaktype)
	addressParagraph.AddText("Heidberge 1").AddBreak(&breaktype)
	addressParagraph.AddText("48282 Emsdetten").AddBreak(&breaktype)
	addressParagraph.AddText("Telefonnummer: 02572-5671").AddBreak(&breaktype)
	addressParagraph.AddText("E-Mail-Adresse: kita.stjakobus-emsdetten@bistum-muenster.de")

	document.AddEmptyParagraph()

	childInformationParagraph := document.AddEmptyParagraph()
	childInformationParagraph.AddText(fmt.Sprintf("Name des Kindes: %s %s", child.FirstName, child.LastName)).AddBreak(&breaktype)
	childInformationParagraph.AddText(fmt.Sprintf("Geburtsdatum: %s", child.Birthdate.Format("02.01.2006"))).AddBreak(&breaktype)
	childInformationParagraph.AddText(fmt.Sprintf("Familiensprache: %s", child.FamilyLanguage)).AddBreak(&breaktype)
	if child.MigrationBackground {
		childInformationParagraph.AddText("Migrationshintergrund: Ja").AddBreak(&breaktype)
	} else {
		childInformationParagraph.AddText("Migrationshintergrund: Nein").AddBreak(&breaktype)
	}
	childInformationParagraph.AddText(fmt.Sprintf("Aufnahmedatum: %s", child.AdmissionDate.Format("02.01.2006"))).AddBreak(&breaktype)
	childInformationParagraph.AddText(fmt.Sprintf("Voraussichtliche Einschulung: %s", child.ExpectedSchoolEnrollment.Format("02.01.2006"))).AddBreak(&breaktype)
	childInformationParagraph.AddText(fmt.Sprintf("Adresse: %s", child.Address)).AddBreak(&breaktype)
	childInformationParagraph.AddText(fmt.Sprintf("Namen der Erziehungsberechtigten: %s, %s", child.Parent1Name, child.Parent2Name)).AddBreak(&breaktype)
	childInformationParagraph.AddText("Entwicklungsbegleiter/-innen, Fachkräfte (von - bis):").AddBreak(&breaktype)
	for _, assignmentText := range assignmentsText {
		childInformationParagraph.AddText(assignmentText).Style("List Bullet").AddBreak(&breaktype)
	}

	document.AddPageBreak()

	document.AddHeading("Kindbeobachtungen", 1)

	// Group entries by category
	entriesByCategory := make(map[string][]models.DocumentationEntry)
	for _, entry := range entries {
		if entry.IsApproved {
			category, err := service.categoryStore.GetByID(entry.CategoryID)
			if err != nil {
				logger.WithError(err).WithField("category_id", entry.CategoryID).Warn("Category not found for entry")
				continue
			}
			entriesByCategory[category.Name] = append(entriesByCategory[category.Name], entry)
		}
	}

	// Add entries to the document
	for categoryName, entries := range entriesByCategory {
		document.AddHeading(fmt.Sprintf("Bildungsbereich: %s", categoryName), 2)
		for _, entry := range entries {
			document.AddParagraph(entry.ObservationDescription).Style("List Bullet")
		}
	}

	var buf bytes.Buffer
	if err := document.Write(&buf); err != nil {
		logger.WithError(err).Error("Error saving generated document")
		return nil, ErrChildReportGenerationFailed
	}

	logger.WithField("child_id", childID).Info("Child report generated successfully")
	return buf.Bytes(), nil
}

func (service *DocumentationEntryServiceImpl) GetDocumentName(ctx context.Context, childID int) (string, error) {
	// Fetch child details to construct the document name
	child, err := service.childStore.GetByID(childID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return "", fmt.Errorf("child with ID %d not found", childID)
		}
		return "", fmt.Errorf("error fetching child details: %w", err)
	}

	documentName := fmt.Sprintf("Bildungsdokumentation_%s_%s_%s.docx", child.FirstName, child.LastName, child.Birthdate.Format("2006-01-02"))

	return documentName, nil
}

func (service *DocumentationEntryServiceImpl) FormatChildTeacherAssignments(assignments []models.Assignment) ([]string, error) {
	if len(assignments) == 0 {
		return []string{"Keine Zuordnungen gefunden"}, nil
	}

	var formattedAssignments []string
	for _, assignment := range assignments {
		// Lookup teacher name for teacher ID
		teacher, err := service.teacherStore.GetByID(assignment.TeacherID)
		if err != nil {
			return nil, err
		}
		assignmentStart := assignment.StartDate.Format("02.01.2006")
		var assignmentEnd string
		if assignment.EndDate == nil {
			assignmentEnd = "heute"
		} else {
			assignmentEnd = assignment.EndDate.Format("02.01.2006")
		}
		formattedAssignments = append(formattedAssignments, fmt.Sprintf("- %s %s (%s bis %s)", teacher.FirstName, teacher.LastName, assignmentStart, assignmentEnd))
	}

	return formattedAssignments, nil
}
