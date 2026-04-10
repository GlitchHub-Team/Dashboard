package historical_data

import (
	"encoding/json"
	"errors"
	"time"

	sensorProfile "backend/internal/sensor/profile"

	"github.com/google/uuid"
)

const DefaultHistoricalDataLimit = 500

var (
	ErrInvalidDateRange = errors.New("invalid date range")
	ErrInvalidTimestamp = errors.New("invalid timestamp format, expected RFC3339")
)

type HistoricalSample struct {
	SensorId  uuid.UUID
	GatewayId uuid.UUID
	TenantId  uuid.UUID
	Profile   sensorProfile.SensorProfile
	Timestamp time.Time
	Data      json.RawMessage
}

type HistoricalDataFilter struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

func (f HistoricalDataFilter) Normalize() HistoricalDataFilter {
	if f.Limit <= 0 {
		f.Limit = DefaultHistoricalDataLimit
	}
	return f
}
