package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/handlers"
	"kitadoc-backend/internal/logger"
	"kitadoc-backend/middleware"
	"kitadoc-backend/services"

	"github.com/sirupsen/logrus" // Keep logrus for now, as it's used in the logger package
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// App holds the application's services and router.
type App struct {
	AuthHandler               *handlers.AuthHandler
	ChildHandler              *handlers.ChildHandler
	TeacherHandler            *handlers.TeacherHandler
	GroupHandler              *handlers.GroupHandler
	CategoryHandler           *handlers.CategoryHandler
	AssignmentHandler         *handlers.AssignmentHandler
	DocumentationEntryHandler *handlers.DocumentationEntryHandler
	AudioRecordingHandler     *handlers.AudioRecordingHandler
	DocumentGenerationHandler *handlers.DocumentGenerationHandler
	BulkOperationsHandler *handlers.BulkOperationsHandler
	Router                *http.ServeMux
	Config                *config.Config // Add Config to App struct
}

// NewApp initializes a new App with all handlers and services.
func NewApp(db *sql.DB, cfg *config.Config) *App {
	// Initialize DAL
	dal := data.NewDAL(db)

	// Initialize Services
	userService := services.NewUserService(dal.Users, cfg)
	childService := services.NewChildService(dal.Children, dal.Groups)
	teacherService := services.NewTeacherService(dal.Teachers)
	groupService := services.NewGroupService(dal.Groups)
	categoryService := services.NewCategoryService(dal.Categories)
	assignmentService := services.NewAssignmentService(dal.Assignments, dal.Children, dal.Teachers)
	documentationEntryService := services.NewDocumentationEntryService(dal.DocumentationEntries, dal.Children, dal.Teachers, dal.Categories, dal.Users)
	audioRecordingService := services.NewAudioRecordingService(dal.AudioRecordings, dal.DocumentationEntries)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(userService)
	childHandler := handlers.NewChildHandler(childService)
	teacherHandler := handlers.NewTeacherHandler(teacherService)
	groupHandler := handlers.NewGroupHandler(groupService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	documentationEntryHandler := handlers.NewDocumentationEntryHandler(documentationEntryService)
	audioRecordingHandler := handlers.NewAudioRecordingHandler(audioRecordingService, cfg)
	documentGenerationHandler := handlers.NewDocumentGenerationHandler(documentationEntryService)
	bulkOperationsHandler := handlers.NewBulkOperationsHandler(childService)

	return &App{
		AuthHandler:               authHandler,
		ChildHandler:              childHandler,
		TeacherHandler:            teacherHandler,
		GroupHandler:              groupHandler,
		CategoryHandler:           categoryHandler,
		AssignmentHandler:         assignmentHandler,
		DocumentationEntryHandler: documentationEntryHandler,
		AudioRecordingHandler: audioRecordingHandler,
		DocumentGenerationHandler: documentGenerationHandler,
		BulkOperationsHandler: bulkOperationsHandler,
		Router:                http.NewServeMux(),
		Config:                cfg, // Assign config to App struct
	}
}

// routes sets up all the HTTP routes and applies middleware.
func (app *App) routes() {
	// Public routes
	app.Router.Handle("POST /api/v1/auth/register", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.RegisterUser)))))
	app.Router.Handle("POST /api/v1/auth/login", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.Login)))))
	app.Router.Handle("GET /health", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(healthCheckHandler)))))

	// Authenticated routes (Teacher and Admin roles)
	authMiddleware := middleware.Authenticate(app.AuthHandler.UserService, app.Config)

	// Auth Endpoints
	app.Router.Handle("POST /api/v1/auth/logout", middleware.RequestIDMiddleware(authMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.Logout))))))
	app.Router.Handle("GET /api/v1/auth/me", middleware.RequestIDMiddleware(authMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.GetMe))))))

	// Children Management Endpoints
	app.Router.Handle("POST /api/v1/children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.CreateChild)))))))
	app.Router.Handle("GET /api/v1/children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.GetAllChildren)))))))
	app.Router.Handle("GET /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.GetChildByID)))))))
	app.Router.Handle("PUT /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.UpdateChild)))))))
	app.Router.Handle("DELETE /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.DeleteChild)))))))

	// Teachers Management Endpoints
	app.Router.Handle("POST /api/v1/teachers", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.CreateTeacher)))))))
	app.Router.Handle("GET /api/v1/teachers", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.GetAllTeachers)))))))
	app.Router.Handle("GET /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.GetTeacherByID)))))))
	app.Router.Handle("PUT /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.UpdateTeacher)))))))
	app.Router.Handle("DELETE /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.DeleteTeacher)))))))

	// Groups Management Endpoints
	app.Router.Handle("POST /api/v1/groups", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.GroupHandler.CreateGroup)))))))
	app.Router.Handle("GET /api/v1/groups", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.GroupHandler.GetAllGroups)))))))
	app.Router.Handle("GET /api/v1/groups/{group_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.GroupHandler.GetGroupByID)))))))
	app.Router.Handle("PUT /api/v1/groups/{group_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.GroupHandler.UpdateGroup)))))))
	app.Router.Handle("DELETE /api/v1/groups/{group_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.GroupHandler.DeleteGroup)))))))

	// Categories Management Endpoints
	app.Router.Handle("POST /api/v1/categories", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.CategoryHandler.CreateCategory)))))))
	app.Router.Handle("GET /api/v1/categories", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.CategoryHandler.GetAllCategories)))))))
	app.Router.Handle("PUT /api/v1/categories/{category_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.CategoryHandler.UpdateCategory)))))))
	app.Router.Handle("DELETE /api/v1/categories/{category_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.CategoryHandler.DeleteCategory)))))))

	// Child-Teacher Assignments Endpoints
	app.Router.Handle("POST /api/v1/assignments", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AssignmentHandler.CreateAssignment)))))))
	app.Router.Handle("GET /api/v1/assignments/child/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AssignmentHandler.GetAssignmentsByChildID)))))))
	app.Router.Handle("PUT /api/v1/assignments/{assignment_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AssignmentHandler.UpdateAssignment)))))))
	app.Router.Handle("DELETE /api/v1/assignments/{assignment_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AssignmentHandler.DeleteAssignment)))))))

	// Documentation Entries Endpoints
	app.Router.Handle("POST /api/v1/documentation", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.CreateDocumentationEntry)))))))
	app.Router.Handle("GET /api/v1/documentation/child/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.GetDocumentationEntriesByChildID)))))))
	app.Router.Handle("PUT /api/v1/documentation/{entry_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.UpdateDocumentationEntry)))))))
	app.Router.Handle("DELETE /api/v1/documentation/{entry_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.DeleteDocumentationEntry)))))))
	app.Router.Handle("PUT /api/v1/documentation/{entry_id}/approve", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.ApproveDocumentationEntry)))))))

	// Audio Recordings Endpoints
	app.Router.Handle("POST /api/v1/audio/upload", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AudioRecordingHandler.UploadAudio)))))))

	// Document Generation Endpoints
	app.Router.Handle("GET /api/v1/documents/child-report/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentGenerationHandler.GenerateChildReport)))))))
	
	// Bulk Operations Endpoints
	app.Router.Handle("POST /api/v1/bulk/import-children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.BulkOperationsHandler.ImportChildren)))))))
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up structured logging
	logLevel, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logrus.Fatalf("Invalid log level in configuration: %v", err)
	}

	var logFormatter logrus.Formatter
	switch cfg.Log.Format {
	case "json":
		logFormatter = &logrus.JSONFormatter{}
	case "text":
		logFormatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}
	default:
		logrus.Fatalf("Unsupported log format: %s. Must be 'json' or 'text'.", cfg.Log.Format)
	}
	logger.InitGlobalLogger(logLevel, logFormatter)

	logger.GetGlobalLogger().Infof("Application starting in %s environment...", cfg.Environment)

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", cfg.Database.DSN)
	if err != nil {
		logger.GetGlobalLogger().Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.GetGlobalLogger().Errorf("Failed to close database connection: %v", err)
		}
	}()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		logger.GetGlobalLogger().Fatalf("Failed to connect to database: %v", err)
	}
	logger.GetGlobalLogger().Info("Successfully connected to the database!")

	// Initialize App
	app := NewApp(db, cfg)
	app.routes() // Setup routes

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      middleware.CORS(app.Router), // Apply CORS middleware globally
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.GetGlobalLogger().Infof("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.GetGlobalLogger().Fatalf("Could not listen on %s: %v", server.Addr, err)
		}
	}()

	<-done
	logger.GetGlobalLogger().Info("Attempting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.GetGlobalLogger().Fatalf("Server shutdown failed: %v", err)
	}
	logger.GetGlobalLogger().Info("Server gracefully shut down.")
}

// healthCheckHandler provides a simple health check endpoint.
func healthCheckHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"status": "ok"})
}
