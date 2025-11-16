package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/stretchr/testify/assert"
)

// Test CreateDevice

func TestDeviceRepository_CreateDevice(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)

	deviceID := uuid.New()
	now := time.Now()
	device := &model.Device{
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateAvailable,
	}

	t.Run("successful creation", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(deviceID, device.Name, device.Brand, device.State, now, now)

		mock.ExpectQuery(`INSERT INTO devices`).
			WithArgs(device.Name, device.Brand, device.State).
			WillReturnRows(rows)

		result, err := repo.CreateDevice(device)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, deviceID, result.ID)
		assert.Equal(t, "iPhone 15", result.Name)
		assert.Equal(t, "Apple", result.Brand)
		assert.Equal(t, model.StateAvailable, result.State)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO devices`).
			WithArgs(device.Name, device.Brand, device.State).
			WillReturnError(fmt.Errorf("database error"))

		result, err := repo.CreateDevice(device)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test GetDeviceByID

func TestDeviceRepository_GetDeviceByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)

	deviceID := uuid.New()
	now := time.Now()

	t.Run("device found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(deviceID, "iPhone 15", "Apple", model.StateAvailable, now, now)

		mock.ExpectQuery(`SELECT \* FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnRows(rows)

		device, err := repo.GetDeviceByID(deviceID.String())

		assert.NoError(t, err)
		assert.NotNil(t, device)
		assert.Equal(t, deviceID, device.ID)
		assert.Equal(t, "iPhone 15", device.Name)
		assert.Equal(t, "Apple", device.Brand)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("device not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnError(sql.ErrNoRows)

		device, err := repo.GetDeviceByID(deviceID.String())

		assert.Error(t, err)
		assert.Nil(t, device)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnError(fmt.Errorf("database error"))

		device, err := repo.GetDeviceByID(deviceID.String())

		assert.Error(t, err)
		assert.Nil(t, device)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test GetDevices

func TestDeviceRepository_GetDevices(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)

	now := time.Now()

	t.Run("get all devices", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(uuid.New(), "iPhone 15", "Apple", model.StateAvailable, now, now).
			AddRow(uuid.New(), "Galaxy S23", "Samsung", model.StateInUse, now, now)

		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 ORDER BY created_at DESC`).
			WillReturnRows(rows)

		filter := DeviceFilter{}
		devices, err := repo.GetDevices(filter)

		assert.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by brand", func(t *testing.T) {
		brand := "Apple"
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(uuid.New(), "iPhone 15", "Apple", model.StateAvailable, now, now)

		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 AND brand = \$1 ORDER BY created_at DESC`).
			WithArgs(brand).
			WillReturnRows(rows)

		filter := DeviceFilter{Brand: &brand}
		devices, err := repo.GetDevices(filter)

		assert.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, "Apple", devices[0].Brand)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by state", func(t *testing.T) {
		state := model.StateInUse
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Galaxy S23", "Samsung", model.StateInUse, now, now)

		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 AND state = \$1 ORDER BY created_at DESC`).
			WithArgs(state).
			WillReturnRows(rows)

		filter := DeviceFilter{State: &state}
		devices, err := repo.GetDevices(filter)

		assert.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, model.StateInUse, devices[0].State)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filter by brand and state", func(t *testing.T) {
		brand := "Apple"
		state := model.StateInUse
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(uuid.New(), "iPhone 14", "Apple", model.StateInUse, now, now)

		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 AND brand = \$1 AND state = \$2 ORDER BY created_at DESC`).
			WithArgs(brand, state).
			WillReturnRows(rows)

		filter := DeviceFilter{Brand: &brand, State: &state}
		devices, err := repo.GetDevices(filter)

		assert.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, "Apple", devices[0].Brand)
		assert.Equal(t, model.StateInUse, devices[0].State)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 ORDER BY created_at DESC`).
			WillReturnRows(rows)

		filter := DeviceFilter{}
		devices, err := repo.GetDevices(filter)

		assert.NoError(t, err)
		assert.Len(t, devices, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM devices WHERE 1=1 ORDER BY created_at DESC`).
			WillReturnError(fmt.Errorf("database error"))

		filter := DeviceFilter{}
		devices, err := repo.GetDevices(filter)

		assert.Error(t, err)
		assert.Len(t, devices, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test UpdateDevice

func TestDeviceRepository_UpdateDevice(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)

	deviceID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	device := &model.Device{
		ID:    deviceID,
		Name:  "iPhone 15",
		Brand: "Apple",
		State: model.StateInUse,
	}

	t.Run("successful update", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "brand", "state", "created_at", "updated_at"}).
			AddRow(deviceID, device.Name, device.Brand, device.State, createdAt, updatedAt)

		mock.ExpectQuery(`UPDATE devices SET name = \$1, brand = \$2, state = \$3 WHERE id = \$4`).
			WithArgs(device.Name, device.Brand, device.State, device.ID).
			WillReturnRows(rows)

		result, err := repo.UpdateDevice(device)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, deviceID, result.ID)
		assert.Equal(t, "iPhone 15", result.Name)
		assert.Equal(t, model.StateInUse, result.State)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("device not found", func(t *testing.T) {
		mock.ExpectQuery(`UPDATE devices SET name = \$1, brand = \$2, state = \$3 WHERE id = \$4`).
			WithArgs(device.Name, device.Brand, device.State, device.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.UpdateDevice(device)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`UPDATE devices SET name = \$1, brand = \$2, state = \$3 WHERE id = \$4`).
			WithArgs(device.Name, device.Brand, device.State, device.ID).
			WillReturnError(fmt.Errorf("database error"))

		result, err := repo.UpdateDevice(device)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Test DeleteDevice

func TestDeviceRepository_DeleteDevice(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewDeviceRepository(db)

	deviceID := uuid.New()

	t.Run("successful deletion", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteDevice(deviceID.String())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("device not found", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteDevice(deviceID.String())

		assert.Error(t, err)
		assert.Equal(t, "device not found", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnError(fmt.Errorf("database error"))

		err := repo.DeleteDevice(deviceID.String())

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error getting rows affected", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM devices WHERE id`).
			WithArgs(deviceID.String()).
			WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

		err := repo.DeleteDevice(deviceID.String())

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
