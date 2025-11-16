// internal/repository/device_repository.go
package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
)

type DeviceRepositoryInterface interface {
	GetDevices(filter DeviceFilter) ([]model.Device, error)
	GetDeviceByID(id string) (*model.Device, error)
	CreateDevice(device *model.Device) (*model.Device, error)
	UpdateDevice(device *model.Device) (*model.Device, error)
	DeleteDevice(id string) error
}

type DeviceRepository struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// DeviceFilter holds the query filters
type DeviceFilter struct {
	Brand *string
	State *model.DeviceState
}

// GetDevices retrieves devices with optional filters
func (r *DeviceRepository) GetDevices(filter DeviceFilter) ([]model.Device, error) {
	var devices []model.Device

	query := "SELECT * FROM devices WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.Brand != nil {
		query += fmt.Sprintf(" AND brand = $%d", argCount)
		args = append(args, *filter.Brand)
		argCount++
	}

	if filter.State != nil {
		query += fmt.Sprintf(" AND state = $%d", argCount)
		args = append(args, *filter.State)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	err := r.db.Select(&devices, query, args...)
	return devices, err
}

// GetDeviceByID retrieves a device by its ID
func (r *DeviceRepository) GetDeviceByID(id string) (*model.Device, error) {
	var device model.Device
	query := "SELECT * FROM devices WHERE id = $1"
	err := r.db.Get(&device, query, id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// CreateDevice creates a new device
func (r *DeviceRepository) CreateDevice(device *model.Device) (*model.Device, error) {
	query := `
        INSERT INTO devices (name, brand, state)
        VALUES ($1, $2, $3)
        RETURNING id, name, brand, state, created_at, updated_at
    `

	err := r.db.QueryRowx(query, device.Name, device.Brand, device.State).StructScan(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// UpdateDevice updates an existing device
func (r *DeviceRepository) UpdateDevice(device *model.Device) (*model.Device, error) {
	query := `
        UPDATE devices 
        SET name = $1, brand = $2, state = $3
        WHERE id = $4
        RETURNING id, name, brand, state, created_at, updated_at
    `

	err := r.db.QueryRowx(query, device.Name, device.Brand, device.State, device.ID).StructScan(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// DeleteDevice deletes a device by its ID
func (r *DeviceRepository) DeleteDevice(id string) error {
	query := "DELETE FROM devices WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}
