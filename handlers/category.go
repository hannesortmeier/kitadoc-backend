package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// CategoryHandler handles category-related HTTP requests.
type CategoryHandler struct {
	CategoryService services.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{CategoryService: categoryService}
}

// CreateCategory handles creating a new category.
func (handler *CategoryHandler) CreateCategory(writer http.ResponseWriter, request *http.Request) {
	var category models.Category
	if err := json.NewDecoder(request.Body).Decode(&category); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}


	createdCategory, err := handler.CategoryService.CreateCategory(&category)
	if err != nil {
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid category data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(createdCategory)
}

// GetAllCategories handles fetching all categories.
func (handler *CategoryHandler) GetAllCategories(writer http.ResponseWriter, request *http.Request) {
	categories, err := handler.CategoryService.GetAllCategories()
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(writer).Encode(categories)
}

// UpdateCategory handles updating an existing category.
func (handler *CategoryHandler) UpdateCategory(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("category_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category models.Category
	if err := json.NewDecoder(request.Body).Decode(&category); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	category.ID = id

	err = handler.CategoryService.UpdateCategory(&category)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Category not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid category data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Category updated successfully"})
}

// DeleteCategory handles deleting a category.
func (handler *CategoryHandler) DeleteCategory(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("category_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = handler.CategoryService.DeleteCategory(id)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Category not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"message": "Category deleted successfully"})
}