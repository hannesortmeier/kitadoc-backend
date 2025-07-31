package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kitadoc-backend/middleware"
	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// TeacherHandler handles teacher-related HTTP requests.
type TeacherHandler struct {
	TeacherService services.TeacherService
}

// NewTeacherHandler creates a new TeacherHandler.
func NewTeacherHandler(teacherService services.TeacherService) *TeacherHandler {
	return &TeacherHandler{TeacherService: teacherService}
}

// CreateTeacher handles creating a new teacher.
func (teacherHandler *TeacherHandler) CreateTeacher(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	var teacher models.Teacher
	if err := json.NewDecoder(request.Body).Decode(&teacher); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	teacher.CreatedAt = time.Now()
	teacher.UpdatedAt = time.Now()

	createdTeacher, err := teacherHandler.TeacherService.CreateTeacher(&teacher)
	if err != nil {
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid teacher data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(writer).Encode(createdTeacher); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllTeachers handles fetching all teachers.
func (teacherHandler *TeacherHandler) GetAllTeachers(writer http.ResponseWriter, request *http.Request) {
	logger := middleware.GetLoggerWithReqID(request.Context())
	teachers, err := teacherHandler.TeacherService.GetAllTeachers()
	if err != nil {
		logger.Errorf("Error fetching all teachers: %v", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(teachers); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetTeacherByID handles fetching a teacher by ID.
func (teacherHandler *TeacherHandler) GetTeacherByID(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("teacher_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	teacher, err := teacherHandler.TeacherService.GetTeacherByID(id)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(teacher); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateTeacher handles updating an existing teacher.
func (teacherHandler *TeacherHandler) UpdateTeacher(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("teacher_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var teacher models.Teacher
	if err := json.NewDecoder(request.Body).Decode(&teacher); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	teacher.ID = id
	teacher.UpdatedAt = time.Now()

	err = teacherHandler.TeacherService.UpdateTeacher(&teacher)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Teacher not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid teacher data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Teacher updated successfully"}); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteTeacher handles deleting a teacher.
func (teacherHandler *TeacherHandler) DeleteTeacher(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("teacher_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	err = teacherHandler.TeacherService.DeleteTeacher(id)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Teacher deleted successfully"}); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
