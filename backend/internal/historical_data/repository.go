package historical_data

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type historicalDataTimescaleRepository struct {
	log *zap.Logger
	db  *sql.DB
}

func newHistoricalDataTimescaleRepository(
	log *zap.Logger,
	db *sql.DB,
) *historicalDataTimescaleRepository {
	return &historicalDataTimescaleRepository{
		log: log,
		db:  db,
	}
}

func (repo *historicalDataTimescaleRepository) GetSensorHistoricalData(
	tenantId uuid.UUID,
	sensorId uuid.UUID,
	filter HistoricalDataFilter,
) ([]HistoricalSample, error) {
	filter = filter.Normalize()

	query, args := buildHistoricalDataQuery(tenantId, sensorId, filter)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	samples := make([]HistoricalSample, 0)
	for rows.Next() {
		var sample HistoricalSample
		var data []byte

		if err := rows.Scan(
			&sample.SensorId,
			&sample.GatewayId,
			&sample.TenantId,
			&sample.Profile,
			&sample.Timestamp,
			&data,
		); err != nil {
			return nil, err
		}

		sample.Data = data
		samples = append(samples, sample)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}

func buildHistoricalDataQuery(
	tenantId uuid.UUID,
	sensorId uuid.UUID,
	filter HistoricalDataFilter,
) (string, []any) {
	var query strings.Builder

	query.WriteString(`SELECT sensor_id, gateway_id, tenant_id, profile, timestamp, data FROM "`)
	query.WriteString(tenantId.String())
	query.WriteString(`".sensor_data WHERE tenant_id = $1 AND sensor_id = $2`)

	args := []any{tenantId, sensorId}
	argPos := 3

	if filter.From != nil {
		query.WriteString(fmt.Sprintf(" AND timestamp >= $%d", argPos))
		args = append(args, *filter.From)
		argPos++
	}

	if filter.To != nil {
		query.WriteString(fmt.Sprintf(" AND timestamp <= $%d", argPos))
		args = append(args, *filter.To)
		argPos++
	}

	query.WriteString(" ORDER BY timestamp ASC")
	query.WriteString(fmt.Sprintf(" LIMIT $%d", argPos))
	args = append(args, filter.Limit)

	return query.String(), args
}
