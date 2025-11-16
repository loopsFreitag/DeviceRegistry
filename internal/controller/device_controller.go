package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"github.com/loopsFreitag/DeviceRegistry/internal/service"
)

type DeviceController struct {
	deviceService service.DeviceServiceInterface
}

func NewDeviceController() *DeviceController {
	deviceRepo := repository.NewDeviceRepository(model.DBX())
	deviceService := service.NewDeviceService(deviceRepo)
	return &DeviceController{
		deviceService: deviceService,
	}
}

// NewDeviceControllerWithService creates a controller with injected service (for testing)
func NewDeviceControllerWithService(deviceService service.DeviceServiceInterface) *DeviceController {
	return &DeviceController{
		deviceService: deviceService,
	}
}

func (dc *DeviceController) SetRoutes(r *mux.Router) {
	r.HandleFunc("/devices", dc.CreateDevice).Methods("POST")
	r.HandleFunc("/devices", dc.UpdateDevice).Methods("PUT")
	r.HandleFunc("/devices", dc.GetDevices).Methods("GET")
	r.HandleFunc("/devices/{id}", dc.GetDevice).Methods("GET")
	r.HandleFunc("/devices/{id}", dc.DeleteDevice).Methods("DELETE")
}

// Helper function to send error responses
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

// CreateDevice godoc
// @Summary      Create a new device
// @Description  Create a new device with the provided details
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device  body      model.Device  true  "Device details"
// @Success      201     {object}  model.Device
// @Failure      400     {object}  ErrorResponse
// @Router       /api/devices [post]
func (dc *DeviceController) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var device model.Device

	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdDevice, err := dc.deviceService.CreateDevice(&device)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdDevice)
}

// UpdateDevice godoc
// @Summary      Update an existing device
// @Description  Update the details of an existing device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device  body      model.Device  true  "Updated device details"
// @Success      200     {object}  model.Device
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Router       /api/devices [put]
func (dc *DeviceController) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	var device model.Device

	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedDevice, err := dc.deviceService.UpdateDevice(&device)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedDevice)
}

// GetDevices godoc
// @Summary      Get devices
// @Description  Retrieve devices with optional filters for brand and state
// @Tags         devices
// @Produce      json
// @Param        brand  query     string  false  "Filter by brand"
// @Param        state  query     string  false  "Filter by state (inactive, available, in-use)"
// @Success      200  {array}   model.Device
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/devices [get]
func (dc *DeviceController) GetDevices(w http.ResponseWriter, r *http.Request) {
	filter := repository.DeviceFilter{}

	if brand := r.URL.Query().Get("brand"); brand != "" {
		filter.Brand = &brand
	}

	if stateParam := r.URL.Query().Get("state"); stateParam != "" {
		var state model.DeviceState

		switch stateParam {
		case "inactive":
			state = model.StateInactive
		case "available":
			state = model.StateAvailable
		case "in-use":
			state = model.StateInUse
		default:
			stateVal, parseErr := strconv.Atoi(stateParam)
			if parseErr != nil {
				sendErrorResponse(w, http.StatusBadRequest, "Invalid state parameter. Use: inactive, available, or in-use")
				return
			}
			state = model.DeviceState(stateVal)
		}

		filter.State = &state
	}

	devices, err := dc.deviceService.GetDevices(filter)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(devices)
}

// GetDevice godoc
// @Summary      Get a device by ID
// @Description  Retrieve the details of a device by its ID
// @Tags         devices
// @Produce      json
// @Param        id   path      string  true  "Device ID"
// @Success      200  {object}  model.Device
// @Failure      404  {object}  ErrorResponse
// @Router       /api/devices/{id} [get]
func (dc *DeviceController) GetDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	device, err := dc.deviceService.GetDeviceByID(id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(device)
}

// DeleteDevice godoc
// @Summary      Delete a device by ID
// @Description  Delete a device using its ID
// @Tags         devices
// @Produce      json
// @Param        id   path      string  true  "Device ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /api/devices/{id} [delete]
func (dc *DeviceController) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := dc.deviceService.DeleteDevice(id)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
