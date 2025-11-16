package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceRepository is a mock implementation of the device repository
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) GetDevices(filter repository.DeviceFilter) ([]model.Device, error) {
	args := m.Called(filter)
	return args.Get(0).([]model.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetDeviceByID(id string) (*model.Device, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepository) CreateDevice(device *model.Device) (*model.Device, error) {
	args := m.Called(device)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepository) UpdateDevice(device *model.Device) (*model.Device, error) {
	args := m.Called(device)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceRepository) DeleteDevice(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Test CreateDevice

func TestCreateDevice_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	device := &model.Device{
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	expectedDevice := &model.Device{
		ID:        uuid.New(),
		Name:      "iPhone 15",
		Brand:     "Apple",
		State:     model.StateAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("CreateDevice", device).Return(expectedDevice, nil)

	result, err := service.CreateDevice(device)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDevice.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestCreateDevice_EmptyName(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	device := &model.Device{
		Name:  "",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	result, err := service.CreateDevice(device)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "device name cannot be empty", err.Error())
	mockRepo.AssertNotCalled(t, "CreateDevice")
}

func TestCreateDevice_EmptyBrand(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	device := &model.Device{
		Name:  "iPhone 15",
		Brand: "",
		State: model.StateAvailable,
	}

	result, err := service.CreateDevice(device)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "device brand cannot be empty", err.Error())
	mockRepo.AssertNotCalled(t, "CreateDevice")
}

// Test UpdateDevice

func TestUpdateDevice_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	existingDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	updatedDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateInUse,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(existingDevice, nil)
	mockRepo.On("UpdateDevice", updatedDevice).Return(updatedDevice, nil)

	result, err := service.UpdateDevice(updatedDevice)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "iPhone 15", result.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateDevice_CannotUpdateNameWhenInUse(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	existingDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateInUse,
	}

	updatedDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 15", // Trying to change name
		Brand: "Apple",
		State: model.StateInUse,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(existingDevice, nil)

	result, err := service.UpdateDevice(updatedDevice)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "cannot update name: device is currently in use", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateDevice")
}

func TestUpdateDevice_CannotUpdateBrandWhenInUse(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	existingDevice := &model.Device{
		ID:    deviceID,
		Name:  "Galaxy S23",
		Brand: "Samsung",
		State: model.StateInUse,
	}

	updatedDevice := &model.Device{
		ID:    deviceID,
		Name:  "Galaxy S23",
		Brand: "Apple", // Trying to change brand
		State: model.StateInUse,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(existingDevice, nil)

	result, err := service.UpdateDevice(updatedDevice)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "cannot update brand: device is currently in use", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateDevice")
}

func TestUpdateDevice_CanUpdateStateWhenInUse(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	existingDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateInUse,
	}

	updatedDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateAvailable, // Only changing state
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(existingDevice, nil)
	mockRepo.On("UpdateDevice", updatedDevice).Return(updatedDevice, nil)

	result, err := service.UpdateDevice(updatedDevice)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.StateAvailable, result.State)
	mockRepo.AssertExpectations(t)
}

func TestUpdateDevice_DeviceNotFound(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	updatedDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(nil, sql.ErrNoRows)

	result, err := service.UpdateDevice(updatedDevice)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "device not found", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateDevice")
}

// Test DeleteDevice

func TestDeleteDevice_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	device := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(device, nil)
	mockRepo.On("DeleteDevice", deviceID.String()).Return(nil)

	err := service.DeleteDevice(deviceID.String())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteDevice_CannotDeleteInUseDevice(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	device := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateInUse,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(device, nil)

	err := service.DeleteDevice(deviceID.String())

	assert.Error(t, err)
	assert.Equal(t, "cannot delete device: device is currently in use", err.Error())
	mockRepo.AssertNotCalled(t, "DeleteDevice")
}

func TestDeleteDevice_DeviceNotFound(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(nil, sql.ErrNoRows)

	err := service.DeleteDevice(deviceID.String())

	assert.Error(t, err)
	assert.Equal(t, "device not found", err.Error())
	mockRepo.AssertNotCalled(t, "DeleteDevice")
}

// Test GetDeviceByID

func TestGetDeviceByID_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()
	expectedDevice := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 14",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(expectedDevice, nil)

	result, err := service.GetDeviceByID(deviceID.String())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDevice.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetDeviceByID_NotFound(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	deviceID := uuid.New()

	mockRepo.On("GetDeviceByID", deviceID.String()).Return(nil, sql.ErrNoRows)

	result, err := service.GetDeviceByID(deviceID.String())

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// Test GetDevices

func TestGetDevices_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	expectedDevices := []model.Device{
		{
			ID:    uuid.New(),
			Name:  "iPhone 14",
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

	filter := repository.DeviceFilter{}
	mockRepo.On("GetDevices", filter).Return(expectedDevices, nil)

	result, err := service.GetDevices(filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetDevices_WithBrandFilter(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	brand := "Apple"
	expectedDevices := []model.Device{
		{
			ID:    uuid.New(),
			Name:  "iPhone 14",
			Brand: "Apple",
			State: model.StateAvailable,
		},
	}

	filter := repository.DeviceFilter{Brand: &brand}
	mockRepo.On("GetDevices", filter).Return(expectedDevices, nil)

	result, err := service.GetDevices(filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Apple", result[0].Brand)
	mockRepo.AssertExpectations(t)
}

func TestGetDevices_EmptyResult(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	expectedDevices := []model.Device{}
	filter := repository.DeviceFilter{}
	mockRepo.On("GetDevices", filter).Return(expectedDevices, nil)

	result, err := service.GetDevices(filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockRepo.AssertExpectations(t)
}

func TestGetDevices_RepositoryError(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	service := NewDeviceService(mockRepo)

	filter := repository.DeviceFilter{}
	mockRepo.On("GetDevices", filter).Return([]model.Device{}, errors.New("database error"))

	_, err := service.GetDevices(filter)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}
