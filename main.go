package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

func main() {
	// Delete the existing database file if it exists
	_, err := os.Stat("./data/kindergarten.db")
	if err == nil {
		err = os.Remove("./data/kindergarten.db")
		if err != nil {
			log.Fatalf("Failed to delete existing database file: %v", err)
		}
		fmt.Println("Existing database file deleted successfully.")
	} else if !os.IsNotExist(err) {
		log.Fatalf("Failed to check existing database file: %v", err)
	}

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/kindergarten.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Successfully connected to the database!")

	// Execute the schema to create tables if they don't exist
	schemaSQL, err := data.ReadSQLSchema("database/data_model.sql")
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	_, err = db.Exec(schemaSQL)
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}
	fmt.Println("Database schema initialized successfully!")

	// Initialize DAL
	dal := data.NewDAL(db)

	// Initialize Services
	userService := services.NewUserService(dal.Users)
	childService := services.NewChildService(dal.Children, dal.Groups)
	teacherService := services.NewTeacherService(dal.Teachers)
	groupService := services.NewGroupService(dal.Groups)
	categoryService := services.NewCategoryService(dal.Categories)
	assignmentService := services.NewAssignmentService(dal.Assignments, dal.Children, dal.Teachers)
	documentationEntryService := services.NewDocumentationEntryService(dal.DocumentationEntries, dal.Children, dal.Teachers, dal.Categories, dal.Users)
	audioRecordingService := services.NewAudioRecordingService(dal.AudioRecordings, dal.DocumentationEntries)

	// --- Demonstration of Service Layer operations ---

	// 1. Register a new user
	fmt.Println("\n--- User Service Demonstration ---")
	newUser, err := userService.RegisterUser("testuser", "password123", "teacher")
	if err != nil {
		if errors.Is(err, services.ErrAlreadyExists) {
			fmt.Println("User 'testuser' already exists. Skipping registration.")
			token, err := userService.LoginUser("testuser", "password123") // Attempt to log in if exists
			if err != nil {
				log.Fatalf("Failed to log in existing user: %v", err)
			}
			newUser, err = userService.GetCurrentUser(token)
			if err != nil {
				log.Fatalf("Failed to get user details for existing user: %v", err)
			}
		} else {
			log.Fatalf("Failed to register user: %v", err)
		}
	} else {
		fmt.Printf("Registered user: %+v\n", newUser)
	}

	// 2. Login the user
	token, err := userService.LoginUser("testuser", "password123")
	if err != nil {
		log.Fatalf("Failed to login user: %v", err)
	}
	fmt.Printf("Logged in successfully. JWT Token: %s\n", token)

	// 3. Get current user from token
	currentUser, err := userService.GetCurrentUser(token)
	if err != nil {
		log.Fatalf("Failed to get current user from token: %v", err)
	}
	fmt.Printf("Current user from token: %+v\n", currentUser)

	// 4. Create a new group
	fmt.Println("\n--- Group Service Demonstration ---")
	newGroup := &models.Group{
		Name: "Demonstration Group",
	}
	createdGroup, err := groupService.CreateGroup(newGroup)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyExists) {
			fmt.Println("Group 'Demonstration Group' already exists. Skipping creation.")
			// Attempt to fetch the existing group
			allGroups, err := groupService.GetAllGroups()
			if err != nil {
				log.Fatalf("Failed to get all groups: %v", err)
			}
			for _, g := range allGroups {
				if g.Name == "Demonstration Group" {
					createdGroup = &g
					break
				}
			}
			if createdGroup == nil {
				log.Fatalf("Failed to find existing 'Demonstration Group'")
			}
		} else {
			log.Fatalf("Failed to create group: %v", err)
		}
	} else {
		fmt.Printf("Created group: %+v\n", createdGroup)
	}

	// 5. Create a new child
	fmt.Println("\n--- Child Service Demonstration ---")
	newChild := &models.Child{
		FirstName:     "Service",
		LastName:      "Child",
		Birthdate:     time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC),
		GroupID:       &createdGroup.ID,
		AdmissionDate: time.Now(),
	}
	createdChild, err := childService.CreateChild(newChild)
	if err != nil {
		log.Fatalf("Failed to create child via service: %v", err)
	}
	fmt.Printf("Created child via service: %+v\n", createdChild)

	// 6. Get child by ID
	fetchedChild, err := childService.GetChildByID(createdChild.ID)
	if err != nil {
		log.Fatalf("Failed to get child by ID via service: %v", err)
	}
	fmt.Printf("Fetched child via service: %+v\n", fetchedChild)

	// 7. Create a new teacher
	fmt.Println("\n--- Teacher Service Demonstration ---")
	newTeacher := &models.Teacher{
		FirstName: "Service",
		LastName:  "Teacher",
	}
	createdTeacher, err := teacherService.CreateTeacher(newTeacher)
	if err != nil {
		log.Fatalf("Failed to create teacher via service: %v", err)
	}
	fmt.Printf("Created teacher via service: %+v\n", createdTeacher)

	// 8. Create a new category
	fmt.Println("\n--- Category Service Demonstration ---")
	newCategory := &models.Category{
		Name: "Service Category",
	}
	createdCategory, err := categoryService.CreateCategory(newCategory)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyExists) {
			fmt.Println("Category 'Service Category' already exists. Skipping creation.")
			// Attempt to fetch the existing category
			allCategories, err := categoryService.GetAllCategories()
			if err != nil {
				log.Fatalf("Failed to get all categories: %v", err)
			}
			for _, c := range allCategories {
				if c.Name == "Service Category" {
					createdCategory = &c
					break
				}
			}
			if createdCategory == nil {
				log.Fatalf("Failed to find existing 'Service Category'")
			}
		} else {
			log.Fatalf("Failed to create category: %v", err)
		}
	} else {
		fmt.Printf("Created category via service: %+v\n", createdCategory)
	}

	// 9. Create an assignment
	fmt.Println("\n--- Assignment Service Demonstration ---")
	newAssignment := &models.Assignment{
		ChildID:   createdChild.ID,
		TeacherID: createdTeacher.ID,
		StartDate: time.Now(),
	}
	createdAssignment, err := assignmentService.CreateAssignment(newAssignment)
	if err != nil {
		log.Fatalf("Failed to create assignment via service: %v", err)
	}
	fmt.Printf("Created assignment via service: %+v\n", createdAssignment)

	// 10. Create a documentation entry
	fmt.Println("\n--- Documentation Entry Service Demonstration ---")
	newEntry := &models.DocumentationEntry{
		ChildID:                createdChild.ID,
		TeacherID:              createdTeacher.ID,
		CategoryID:             createdCategory.ID,
		ObservationDate:        time.Now(),
		ObservationDescription: "This is a test documentation entry created via the service layer.",
	}
	createdEntry, err := documentationEntryService.CreateDocumentationEntry(newEntry)
	if err != nil {
		log.Fatalf("Failed to create documentation entry via service: %v", err)
	}
	fmt.Printf("Created documentation entry via service: %+v\n", createdEntry)

	// 11. Approve documentation entry
	err = documentationEntryService.ApproveDocumentationEntry(createdEntry.ID, currentUser.ID)
	if err != nil {
		log.Fatalf("Failed to approve documentation entry: %v", err)
	}
	fmt.Printf("Documentation entry %d approved by user %d.\n", createdEntry.ID, currentUser.ID)

	// 12. Upload audio recording (placeholder)
	fmt.Println("\n--- Audio Recording Service Demonstration ---")
	newAudioRecording := &models.AudioRecording{
		DocumentationEntryID: createdEntry.ID,
		DurationSeconds:      60,
		FilePath:             "/path/to/audio/file.mp3", // Placeholder path
	}
	// Simulate file content
	dummyFileContent := []byte("dummy audio data")
	uploadedRecording, err := audioRecordingService.UploadAudioRecording(newAudioRecording, dummyFileContent)
	if err != nil {
		log.Fatalf("Failed to upload audio recording: %v", err)
	}
	fmt.Printf("Uploaded audio recording: %+v\n", uploadedRecording)

	// 13. Generate child report (placeholder)
	fmt.Println("\n--- Report Generation Demonstration ---")
	reportContent, err := documentationEntryService.GenerateChildReport(createdChild.ID)
	if err != nil {
		fmt.Printf("Failed to generate child report (expected placeholder error): %v\n", err)
	} else {
		fmt.Printf("Generated child report with %d bytes of content.\n", len(reportContent))
	}

	// 14. Bulk import children (placeholder)
	fmt.Println("\n--- Bulk Import Demonstration ---")
	dummyImportContent := []byte("child1,child2,child3")
	err = childService.BulkImportChildren(dummyImportContent)
	if err != nil {
		fmt.Printf("Failed to bulk import children (expected placeholder error): %v\n", err)
	} else {
		fmt.Println("Bulk import of children successful.")
	}

	fmt.Println("\nService layer demonstration complete.")
}
