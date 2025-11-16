// internal/service/device_service.go
package service

import (
	"fmt"

	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
)

// i love how go auto matches interface with implementations
type DeviceServiceInterface interface {
	GetDevices(filter repository.DeviceFilter) ([]model.Device, error)
	GetDeviceByID(id string) (*model.Device, error)
	CreateDevice(device *model.Device) (*model.Device, error)
	UpdateDevice(device *model.Device) (*model.Device, error)
	DeleteDevice(id string) error
}

type DeviceService struct {
	repo repository.DeviceRepositoryInterface
}

func NewDeviceService(repo repository.DeviceRepositoryInterface) *DeviceService {
	return &DeviceService{repo: repo}
}

// GetDevices retrieves devices with optional filters
func (s *DeviceService) GetDevices(filter repository.DeviceFilter) ([]model.Device, error) {
	return s.repo.GetDevices(filter)
}

// GetDeviceByID retrieves a device by its ID
func (s *DeviceService) GetDeviceByID(id string) (*model.Device, error) {
	return s.repo.GetDeviceByID(id)
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(device *model.Device) (*model.Device, error) {
	if device.Name == "" {
		return nil, fmt.Errorf("device name cannot be empty")
	}
	if device.Brand == "" {
		return nil, fmt.Errorf("device brand cannot be empty")
	}

	return s.repo.CreateDevice(device)
}

// UpdateDevice updates an existing device
func (s *DeviceService) UpdateDevice(device *model.Device) (*model.Device, error) {
	existingDevice, err := s.repo.GetDeviceByID(device.ID.String())
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	if existingDevice.State == model.StateInUse {
		if device.Name != existingDevice.Name {
			return nil, fmt.Errorf("cannot update name: device is currently in use")
		}
		if device.Brand != existingDevice.Brand {
			return nil, fmt.Errorf("cannot update brand: device is currently in use")
		}
	}

	return s.repo.UpdateDevice(device)
}

// DeleteDevice deletes a device by its ID
func (s *DeviceService) DeleteDevice(id string) error {
	device, err := s.repo.GetDeviceByID(id)
	if err != nil {
		return fmt.Errorf("device not found")
	}

	if device.State == model.StateInUse {
		return fmt.Errorf("cannot delete device: device is currently in use")
	}

	return s.repo.DeleteDevice(id)
}
