package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"kitadoc-backend/data"
	"kitadoc-backend/services"
)

type ProcessHandler struct {
	processService services.ProcessService
}

func NewProcessHandler(service services.ProcessService) *ProcessHandler {
	return &ProcessHandler{
		processService: service,
	}
}

// GetStatus handles fetching process status.
func (handler *ProcessHandler) GetStatus(writer http.ResponseWriter, request *http.Request) {
	idStr := request.PathValue("process_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, "Invalid process ID", http.StatusBadRequest)
		return
	}
	process, err := handler.processService.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			http.Error(writer, "Process not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Failed to get process status", http.StatusInternalServerError)
		return
	}
	// Return the process status as a JSON response
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(process)
	if err != nil {
		http.Error(writer, "Failed to encode process status", http.StatusInternalServerError)
	}
}
