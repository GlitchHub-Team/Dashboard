package sensor

import (
	"encoding/json"
	"errors"

	"backend/internal/infra/transport/http/dto"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (repo *sensorPostgreRepository) UpdateSensor(sensorId string, status string) error {
	db := (*gorm.DB)(repo.db)
	return db.Model(&SensorEntity{}).
		Where("id = ?", sensorId).
		Update("status", status).Error
}

func (repo *sensorNatsRepository) SendInterruptSensorCmd(cmd *InterruptSensorCmdEntity) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		repo.log.Error("Failed to marshal interrupt sensor command", zap.Error(err))
		return err
	}

	msg, err := repo.nc.Request(INTERRUPT_SENSOR_CMD_SUBJECT, data, TIMEOUT_DURATION)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			repo.log.Error("Request has timed out while waiting for a reply from NATS", zap.Error(err))
			return err
		}
		repo.log.Error("Failed to publish interrupt sensor command", zap.Error(err))
		return err
	}

	var resp dto.CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		repo.log.Error("Failed to unmarshal reply from NATS", zap.Error(err))
		return err
	}

	if !resp.Success {
		repo.log.Error("Interrupt sensor command failed", zap.String("message", resp.Message))
		return errors.New(resp.Message)
	}

	return nil
}

func (repo *sensorNatsRepository) SendResumeSensorCmd(cmd *ResumeSensorCmdEntity) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		repo.log.Error("Failed to marshal resume sensor command", zap.Error(err))
		return err
	}

	msg, err := repo.nc.Request(RESUME_SENSOR_CMD_SUBJECT, data, TIMEOUT_DURATION)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			repo.log.Error("Request has timed out while waiting for a reply from NATS", zap.Error(err))
			return err
		}
		repo.log.Error("Failed to publish resume sensor command", zap.Error(err))
		return err
	}

	var resp dto.CommandResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		repo.log.Error("Failed to unmarshal reply from NATS", zap.Error(err))
		return err
	}

	if !resp.Success {
		repo.log.Error("Resume sensor command failed", zap.String("message", resp.Message))
		return errors.New(resp.Message)
	}

	return nil
}
