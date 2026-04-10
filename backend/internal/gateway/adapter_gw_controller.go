package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/infra/transport/http/dto"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type TimeoutNATSClient time.Duration

type GatewayCommandNATSRepository struct {
	nc      *nats.Conn
	timeout time.Duration
}

func NewGatewayCommandNATSRepository(nc *nats.Conn, timeout TimeoutNATSClient) *GatewayCommandNATSRepository {
	return &GatewayCommandNATSRepository{
		nc:      nc,
		timeout: time.Duration(timeout),
	}
}

func (a *GatewayCommandNATSRepository) sendCommand(subject string, payload []byte) error {
	msg, err := a.nc.Request(subject, payload, a.timeout)
	if err != nil {
		return fmt.Errorf("NATS request failed for subject %s: %w", subject, err)
	}

	var response dto.CommandResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return fmt.Errorf("invalid NATS response for subject %s, error: %v", subject, err)
	}

	if !response.Success {
		return fmt.Errorf("command failed for subject %s, message: %s", subject, response.Message)
	}

	return nil
}

func (a *GatewayCommandNATSRepository) SendCreateGateway(gatewayId uuid.UUID, interval int64) error {
	payload := createGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
		Interval:  interval,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal create gateway command payload: %w", err)
	}

	return a.sendCommand(CREATE_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendDeleteGateway(gatewayId uuid.UUID) error {
	payload := deleteGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete gateway command payload: %w", err)
	}

	return a.sendCommand(DELETE_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendCommission(gatewayId uuid.UUID, tenantId uuid.UUID, token string) error {
	payload := commissionGatewayCommandPayloadDTO{
		GatewayId:       gatewayId.String(),
		TenantId:        tenantId.String(),
		CommissionToken: token,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal commission gateway command payload: %w", err)
	}

	return a.sendCommand(COMMISSION_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendDecommission(gatewayId uuid.UUID) error {
	payload := decommissionGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal decommission gateway command payload: %w", err)
	}

	return a.sendCommand(DECOMMISSION_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendInterrupt(gatewayId uuid.UUID) error {
	payload := interruptGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal interrupt gateway command payload: %w", err)
	}

	return a.sendCommand(INTERRUPT_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendResume(gatewayId uuid.UUID) error {
	payload := resumeGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal resume gateway command payload: %w", err)
	}

	return a.sendCommand(RESUME_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendReset(gatewayId uuid.UUID) error {
	payload := resetGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal reset gateway command payload: %w", err)
	}

	return a.sendCommand(RESET_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

func (a *GatewayCommandNATSRepository) SendReboot(gatewayId uuid.UUID) error {
	payload := rebootGatewayCommandPayloadDTO{
		GatewayId: gatewayId.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal reboot gateway command payload: %w", err)
	}

	return a.sendCommand(REBOOT_GATEWAY_COMMAND_SUBJECT, payloadBytes)
}

var _ GatewayCommandPort = (*GatewayCommandNATSRepository)(nil)
