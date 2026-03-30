package sensor

import (
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

func (repo *sensorPostgreRepository) CreateSensor(entity *SensorEntity) error {
	err := repo.db.
		Clauses(clause.Returning{}).
		Create(entity).
		Error
	if err != nil {
		repo.log.Error("Failed to create sensor", zap.Error(err))
		return err
	}
	return nil
}

func (repo *sensorPostgreRepository) DeleteSensor(entity *SensorEntity) error {
	err := repo.db.
		Clauses(clause.Returning{}).
		Delete(entity).
		Error
	if err != nil {
		repo.log.Error("Failed to delete sensor", zap.Error(err))
		return err
	}
	return nil
}

func (repo *sensorNatsRepository) SendCreateSensorCmd(cmd *CreateSensorCmdEntity) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		repo.log.Error("Failed to marshal create sensor command", zap.Error(err))
		return err
	}

	msg, err := repo.nc.Request(CREATE_SENSOR_CMD_SUBJECT, data, TIMEOUT_DURATION)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			repo.log.Error("Request has timed out while waiting for a reply from NATS", zap.Error(err))
			return err
		}
		repo.log.Error("Failed to publish create sensor command", zap.Error(err))
		return err
	}

	var resp CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		repo.log.Error("Failed to unmarshal reply from NATS", zap.Error(err))
		return err
	}

	if !resp.Success {
		repo.log.Error("Create sensor command failed", zap.String("message", resp.Message))
		return errors.New(resp.Message)
	}

	return nil
}

func (repo *sensorNatsRepository) SendDeleteSensorCmd(cmd *DeleteSensorCmdEntity) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		repo.log.Error("Failed to marshal delete sensor command", zap.Error(err))
		return err
	}

	msg, err := repo.nc.Request(DELETE_SENSOR_CMD_SUBJECT, data, TIMEOUT_DURATION)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			repo.log.Error("Request has timed out while waiting for a reply from NATS", zap.Error(err))
			return err
		}
		repo.log.Error("Failed to publish delete sensor command", zap.Error(err))
		return err
	}

	var resp CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		repo.log.Error("Failed to unmarshal reply from NATS", zap.Error(err))
		return err
	}

	if !resp.Success {
		repo.log.Error("Delete sensor command failed", zap.String("message", resp.Message))
		return errors.New(resp.Message)
	}

	return nil
}
