package historical_data

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GetHistoricalDataPort interface {
	GetSensorHistoricalData(
		tenantId uuid.UUID,
		sensorId uuid.UUID,
		filter HistoricalDataFilter,
	) ([]HistoricalSample, error)
}

type HistoricalDataTimescaleAdapter struct {
	log  *zap.Logger
	repo *historicalDataTimescaleRepository
}

func NewHistoricalDataTimescaleAdapter(
	log *zap.Logger,
	repository *historicalDataTimescaleRepository,
) *HistoricalDataTimescaleAdapter {
	return &HistoricalDataTimescaleAdapter{
		log:  log,
		repo: repository,
	}
}

func (adapter *HistoricalDataTimescaleAdapter) GetSensorHistoricalData(
	tenantId uuid.UUID,
	sensorId uuid.UUID,
	filter HistoricalDataFilter,
) ([]HistoricalSample, error) {
	return adapter.repo.GetSensorHistoricalData(tenantId, sensorId, filter)
}

var _ GetHistoricalDataPort = (*HistoricalDataTimescaleAdapter)(nil)
