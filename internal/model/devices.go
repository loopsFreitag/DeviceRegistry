// internal/model/device.go
package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DeviceState int

const (
	StateInactive  DeviceState = 0
	StateAvailable DeviceState = 1
	StateInUse     DeviceState = 2
)

func (s DeviceState) String() string {
	switch s {
	case StateInactive:
		return "inactive"
	case StateAvailable:
		return "available"
	case StateInUse:
		return "in-use"
	default:
		return "unknown"
	}
}

func (s DeviceState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *DeviceState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		var num int
		if err := json.Unmarshal(data, &num); err != nil {
			return err
		}
		*s = DeviceState(num)
		return nil
	}

	switch str {
	case "inactive":
		*s = StateInactive
	case "available":
		*s = StateAvailable
	case "in-use":
		*s = StateInUse
	default:
		return fmt.Errorf("invalid device state: %s", str)
	}

	return nil
}

// Device represents a device in the system
type Device struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string      `json:"name" db:"name" binding:"required"`
	Brand     string      `json:"brand" db:"brand" binding:"required"`
	State     DeviceState `json:"state" db:"state" binding:"required"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}
