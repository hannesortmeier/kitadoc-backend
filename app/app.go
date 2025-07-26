package app

import (
	"encoding/json"
	"net/http"
	"time"

	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/handlers"
	"kitadoc-backend/middleware"
	"kitadoc-backend/services"
)

// Application holds the application's services and router.
type Application struct {
	AuthHandler               *handlers.AuthHandler
	ChildHandler              *handlers.ChildHandler
	TeacherHandler            *handlers.TeacherHandler
	CategoryHandler           *handlers.CategoryHandler
	AssignmentHandler         *handlers.AssignmentHandler
	DocumentationEntryHandler *handlers.DocumentationEntryHandler
	AudioRecordingHandler     *handlers.AudioRecordingHandler
	DocumentGenerationHandler *handlers.DocumentGenerationHandler
	BulkOperationsHandler     *handlers.BulkOperationsHandler
	Router                    *http.ServeMux
	Config                    config.Config
}

// NewApplication initializes a new Application with all handlers and services.
func NewApplication(cfg config.Config, dal *data.DAL) *Application {
	// Initialize Services
	userService := services.NewUserService(dal.Users, &cfg)
	childService := services.NewChildService(dal.Children)
	teacherService := services.NewTeacherService(dal.Teachers)
	categoryService := services.NewCategoryService(dal.Categories)
	assignmentService := services.NewAssignmentService(dal.Assignments, dal.Children, dal.Teachers)
	documentationEntryService := services.NewDocumentationEntryService(dal.DocumentationEntries, dal.Children, dal.Teachers, dal.Categories, dal.Users)
	audioAnalysisService := services.NewAudioAnalysisService(&http.Client{Timeout: 10 * time.Minute}, cfg.AudioProcServiceURL)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(userService)
	childHandler := handlers.NewChildHandler(childService)
	teacherHandler := handlers.NewTeacherHandler(teacherService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	documentationEntryHandler := handlers.NewDocumentationEntryHandler(documentationEntryService)
	audioRecordingHandler := handlers.NewAudioRecordingHandler(audioAnalysisService, documentationEntryService, &cfg)
	documentGenerationHandler := handlers.NewDocumentGenerationHandler(documentationEntryService, assignmentService)
	bulkOperationsHandler := handlers.NewBulkOperationsHandler(childService)

	app := &Application{
		AuthHandler:               authHandler,
		ChildHandler:              childHandler,
		TeacherHandler:            teacherHandler,
		CategoryHandler:           categoryHandler,
		AssignmentHandler:         assignmentHandler,
		DocumentationEntryHandler: documentationEntryHandler,
		AudioRecordingHandler:     audioRecordingHandler,
		DocumentGenerationHandler: documentGenerationHandler,
		BulkOperationsHandler:     bulkOperationsHandler,
		Router:                    http.NewServeMux(),
		Config:                    cfg,
	}

	// Don't set up routes automatically here
	return app
}

// GetRouter returns the router with all routes set up
func (app *Application) GetRouter() http.Handler {
	// Just return the router without applying CORS again
	return app.Router
}

// Routes sets up all the HTTP routes and applies middleware.
func (app *Application) Routes() http.Handler {
	// Public routes
	app.Router.Handle("POST /api/v1/auth/register", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.RegisterUser)))))
	app.Router.Handle("POST /api/v1/auth/login", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.Login)))))
	app.Router.Handle("GET /health", middleware.RequestIDMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(healthCheckHandler)))))

	// Add a generic OPTIONS handler for all paths that need CORS
	// This handler will be wrapped by the CORS middleware later
	app.Router.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		// The CORS middleware will handle setting the appropriate headers
		// and writing the status. We just need to ensure this handler is called.
		w.WriteHeader(http.StatusOK)
	})

	// Authenticated routes (Teacher and Admin roles)
	authMiddleware := middleware.Authenticate(app.AuthHandler.UserService, &app.Config)

	// Auth Endpoints
	app.Router.Handle("POST /api/v1/auth/logout", middleware.RequestIDMiddleware(authMiddleware(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AuthHandler.Logout))))))
	app.Router.Handle("GET /api/v1/auth/me", middleware.RequestIDMiddleware(authMiddleware(middleware.RequestLogger(http.HandlerFunc(app.AuthHandler.GetMe)))))

	// Children Management Endpoints
	app.Router.Handle("POST /api/v1/children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.CreateChild)))))))
	app.Router.Handle("GET /api/v1/children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.GetAllChildren)))))))
	app.Router.Handle("GET /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.GetChildByID)))))))
	app.Router.Handle("PUT /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.UpdateChild)))))))
	app.Router.Handle("DELETE /api/v1/children/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.ChildHandler.DeleteChild)))))))

	// Teachers Management Endpoints
	app.Router.Handle("POST /api/v1/teachers", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.CreateTeacher)))))))
	app.Router.Handle("GET /api/v1/teachers", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.GetAllTeachers)))))))
	app.Router.Handle("GET /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.GetTeacherByID)))))))
	app.Router.Handle("PUT /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.UpdateTeacher)))))))
	app.Router.Handle("DELETE /api/v1/teachers/{teacher_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.TeacherHandler.DeleteTeacher)))))))

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
	app.Router.Handle("DELETE /api/v1/documentation/{entry_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.DeleteDocumentationEntry)))))))
	app.Router.Handle("PUT /api/v1/documentation/{entry_id}/approve", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentationEntryHandler.ApproveDocumentationEntry)))))))

	// Audio Recordings Endpoints
	app.Router.Handle("POST /api/v1/audio/upload", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.AudioRecordingHandler.UploadAudio)))))))

	// Document Generation Endpoints
	app.Router.Handle("GET /api/v1/documents/child-report/{child_id}", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleTeacher)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.DocumentGenerationHandler.GenerateChildReport)))))))

	// Bulk Operations Endpoints
	app.Router.Handle("POST /api/v1/bulk/import-children", middleware.RequestIDMiddleware(authMiddleware(middleware.Authorize(data.RoleAdmin)(middleware.RequestLogger(middleware.Recovery(http.HandlerFunc(app.BulkOperationsHandler.ImportChildren)))))))

	// Apply CORS middleware globally
	return middleware.CORS(app.Router)
}

// healthCheckHandler provides a simple health check endpoint.
func healthCheckHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(writer, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
