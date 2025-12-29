package handlers

import (
	"encoding/json"
	"net/http"

	"kitadoc-backend/models"
	"kitadoc-backend/services"
)

// KitaMasterdataHandler handles Kita master data-related HTTP requests.
type KitaMasterdataHandler struct {
	KitaMasterdataService services.KitaMasterdataService
}

// NewKitaMasterdataHandler creates a new KitaMasterdataHandler.
func NewKitaMasterdataHandler(kitaMasterdataService services.KitaMasterdataService) *KitaMasterdataHandler {
	return &KitaMasterdataHandler{KitaMasterdataService: kitaMasterdataService}
}

// GetKitaMasterdata handles fetching the Kita master data.
func (handler *KitaMasterdataHandler) GetKitaMasterdata(writer http.ResponseWriter, request *http.Request) {
	masterdata, err := handler.KitaMasterdataService.GetKitaMasterdata()
	if err != nil {
		if err == services.ErrNotFound {
			http.Error(writer, "Kita master data not found", http.StatusNotFound)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(masterdata); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateKitaMasterdata handles updating the Kita master data.
func (handler *KitaMasterdataHandler) UpdateKitaMasterdata(writer http.ResponseWriter, request *http.Request) {
	var masterdata models.KitaMasterdata
	if err := json.NewDecoder(request.Body).Decode(&masterdata); err != nil {
		http.Error(writer, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := handler.KitaMasterdataService.UpdateKitaMasterdata(&masterdata)
	if err != nil {
		if err == services.ErrInvalidInput {
			http.Error(writer, "Invalid Kita master data provided", http.StatusBadRequest)
			return
		}
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(map[string]string{"message": "Kita master data updated successfully"}); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
