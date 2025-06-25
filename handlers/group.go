package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// GroupHandler handles group-related HTTP requests.
type GroupHandler struct {
	GroupService services.GroupService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(groupService services.GroupService) *GroupHandler {
	return &GroupHandler{GroupService: groupService}
}

// CreateGroup handles creating a new group.
func (groupHandler *GroupHandler) CreateGroup(writer http.ResponseWriter, request *http.Request) {
	var group models.Group
	if err := json.NewDecoder(request.Body).Decode(&group); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	createdGroup, err := groupHandler.GroupService.CreateGroup(&group)
	if err != nil {
		switch err {
		case services.ErrInvalidInput:
			http.Error(writer, "Invalid group data provided", http.StatusBadRequest)
			return
		case services.ErrAlreadyExists:
			http.Error(writer, "Group already exists", http.StatusConflict)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(writer).Encode(createdGroup); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllGroups handles fetching all groups.
func (groupHandler *GroupHandler) GetAllGroups(writer http.ResponseWriter, request *http.Request) {
	groups, err := groupHandler.GroupService.GetAllGroups()
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(groups); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetGroupByID handles fetching a group by ID.
func (groupHandler *GroupHandler) GetGroupByID(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("group_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := groupHandler.GroupService.GetGroupByID(id)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(group); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateGroup handles updating an existing group.
func (groupHandler *GroupHandler) UpdateGroup(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("group_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var group models.Group
	if err := json.NewDecoder(request.Body).Decode(&group); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	group.ID = id
	group.UpdatedAt = time.Now()

	err = groupHandler.GroupService.UpdateGroup(&group)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Group not found", http.StatusNotFound)
			return
		}
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid group data provided", http.StatusBadRequest)
			return
		}
		if err == services.ErrAlreadyExists {
			http.Error(writer, "Group with the same name already exists", http.StatusConflict)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Group updated successfully"}); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteGroup handles deleting a group.
func (groupHandler *GroupHandler) DeleteGroup(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("group_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid group ID", http.StatusBadRequest)
		return
	}

	err = groupHandler.GroupService.DeleteGroup(id)
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Group deleted successfully"}); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
