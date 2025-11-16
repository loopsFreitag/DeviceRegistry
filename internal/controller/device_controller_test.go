// internal/controller/device_controller_test.go
package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceService is a mock implementation of the device service
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) GetDevices(filter repository.DeviceFilter) ([]model.Device, error) {
	args := m.Called(filter)
	return args.Get(0).([]model.Device), args.Error(1)
}

func (m *MockDeviceService) GetDeviceByID(id string) (*model.Device, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) CreateDevice(device *model.Device) (*model.Device, error) {
	args := m.Called(device)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) UpdateDevice(device *model.Device) (*model.Device, error) {
	args := m.Called(device)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceService) DeleteDevice(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Test CreateDevice

func TestCreateDevice_Success(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService) // Use the testing constructor

	deviceID := uuid.New()
	requestDevice := model.Device{
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	responseDevice := model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockService.On("CreateDevice", mock.AnythingOfType("*model.Device")).Return(&responseDevice, nil)

	body, _ := json.Marshal(requestDevice)
	req := httptest.NewRequest("POST", "/api/devices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.CreateDevice(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var result model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "Apple", result.Brand)
	mockService.AssertExpectations(t)
}

func TestCreateDevice_InvalidJSON(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	req := httptest.NewRequest("POST", "/api/devices", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.CreateDevice(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.Equal(t, "error", errorResp.Status)
	assert.Equal(t, "Invalid request body", errorResp.Message)
}

func TestCreateDevice_ServiceError(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	requestDevice := model.Device{
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockService.On("CreateDevice", mock.AnythingOfType("*model.Device")).Return(nil, errors.New("database error"))

	body, _ := json.Marshal(requestDevice)
	req := httptest.NewRequest("POST", "/api/devices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.CreateDevice(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Test UpdateDevice

func TestUpdateDevice_Success(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	requestDevice := model.Device{
		ID:    deviceID,
		Name:  "iPhone 15 Pro",
		Brand: "Apple",
		State: model.StateInUse,
	}

	mockService.On("UpdateDevice", mock.AnythingOfType("*model.Device")).Return(&requestDevice, nil)

	body, _ := json.Marshal(requestDevice)
	req := httptest.NewRequest("PUT", "/api/devices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.UpdateDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "iPhone 15 Pro", result.Name)
	mockService.AssertExpectations(t)
}

func TestUpdateDevice_InvalidJSON(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	req := httptest.NewRequest("PUT", "/api/devices", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.UpdateDevice(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDevice_NotFound(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	requestDevice := model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockService.On("UpdateDevice", mock.AnythingOfType("*model.Device")).Return(nil, errors.New("device not found"))

	body, _ := json.Marshal(requestDevice)
	req := httptest.NewRequest("PUT", "/api/devices", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.UpdateDevice(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Test GetDevices

func TestGetDevices_Success(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	devices := []model.Device{
		{
			ID:    uuid.New(),
			Name:  "iPhone 15",
			Brand: "Apple",
			State: model.StateAvailable,
		},
		{
			ID:    uuid.New(),
			Name:  "Galaxy S23",
			Brand: "Samsung",
			State: model.StateInUse,
		},
	}

	mockService.On("GetDevices", mock.AnythingOfType("repository.DeviceFilter")).Return(devices, nil)

	req := httptest.NewRequest("GET", "/api/devices", nil)
	w := httptest.NewRecorder()

	controller.GetDevices(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Len(t, result, 2)
	mockService.AssertExpectations(t)
}

func TestGetDevices_WithBrandFilter(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	devices := []model.Device{
		{
			ID:    uuid.New(),
			Name:  "iPhone 15",
			Brand: "Apple",
			State: model.StateAvailable,
		},
	}

	mockService.On("GetDevices", mock.MatchedBy(func(filter repository.DeviceFilter) bool {
		return filter.Brand != nil && *filter.Brand == "Apple"
	})).Return(devices, nil)

	req := httptest.NewRequest("GET", "/api/devices?brand=Apple", nil)
	w := httptest.NewRecorder()

	controller.GetDevices(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Apple", result[0].Brand)
	mockService.AssertExpectations(t)
}

func TestGetDevices_WithStateFilter(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	devices := []model.Device{
		{
			ID:    uuid.New(),
			Name:  "iPhone 15",
			Brand: "Apple",
			State: model.StateInUse,
		},
	}

	mockService.On("GetDevices", mock.MatchedBy(func(filter repository.DeviceFilter) bool {
		return filter.State != nil && *filter.State == model.StateInUse
	})).Return(devices, nil)

	req := httptest.NewRequest("GET", "/api/devices?state=in-use", nil)
	w := httptest.NewRecorder()

	controller.GetDevices(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Len(t, result, 1)
	mockService.AssertExpectations(t)
}

func TestGetDevices_InvalidStateParameter(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	req := httptest.NewRequest("GET", "/api/devices?state=invalid", nil)
	w := httptest.NewRecorder()

	controller.GetDevices(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.Contains(t, errorResp.Message, "Invalid state parameter")
}

func TestGetDevices_ServiceError(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	mockService.On("GetDevices", mock.AnythingOfType("repository.DeviceFilter")).Return([]model.Device{}, errors.New("database error"))

	req := httptest.NewRequest("GET", "/api/devices", nil)
	w := httptest.NewRecorder()

	controller.GetDevices(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// Test GetDevice

func TestGetDevice_Success(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	device := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockService.On("GetDeviceByID", deviceID.String()).Return(device, nil)

	req := httptest.NewRequest("GET", "/api/devices/"+deviceID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": deviceID.String()})
	w := httptest.NewRecorder()

	controller.GetDevice(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Device
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, deviceID, result.ID)
	assert.Equal(t, "iPhone 15", result.Name)
	mockService.AssertExpectations(t)
}

func TestGetDevice_NotFound(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	mockService.On("GetDeviceByID", deviceID.String()).Return(nil, errors.New("device not found"))

	req := httptest.NewRequest("GET", "/api/devices/"+deviceID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": deviceID.String()})
	w := httptest.NewRecorder()

	controller.GetDevice(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// Test DeleteDevice

func TestDeleteDevice_Success(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	mockService.On("DeleteDevice", deviceID.String()).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/devices/"+deviceID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": deviceID.String()})
	w := httptest.NewRecorder()

	controller.DeleteDevice(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteDevice_NotFound(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	mockService.On("DeleteDevice", deviceID.String()).Return(errors.New("device not found"))

	req := httptest.NewRequest("DELETE", "/api/devices/"+deviceID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": deviceID.String()})
	w := httptest.NewRecorder()

	controller.DeleteDevice(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteDevice_InUse(t *testing.T) {
	mockService := new(MockDeviceService)
	controller := NewDeviceControllerWithService(mockService)

	deviceID := uuid.New()
	mockService.On("DeleteDevice", deviceID.String()).Return(errors.New("cannot delete device: device is currently in use"))

	req := httptest.NewRequest("DELETE", "/api/devices/"+deviceID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": deviceID.String()})
	w := httptest.NewRecorder()

	controller.DeleteDevice(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.Contains(t, errorResp.Message, "in use")
	mockService.AssertExpectations(t)
}
