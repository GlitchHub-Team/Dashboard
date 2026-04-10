package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type TimeoutNATSClient time.Duration

type GatewayCommandNATSAdapter struct {
	nc      *nats.Conn
	timeout time.Duration
}

func NewGatewayCommandNATSAdapter(nc *nats.Conn, timeout TimeoutNATSClient) *GatewayCommandNATSAdapter {
	return &GatewayCommandNATSAdapter{
		nc:      nc,
		timeout: time.Duration(timeout),
	}
}

type gatewayCommandPayload struct {
	Action    string `json:"action"`
	TenantId  string `json:"tenant_id,omitempty"`
	Frequency int    `json:"frequency,omitempty"`
	Token     string `json:"token,omitempty"`
}

func (a *GatewayCommandNATSAdapter) sendCommand(gatewayId uuid.UUID, payload gatewayCommandPayload) error {
	subject := fmt.Sprintf("gateway.%s.command", gatewayId.String())

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal command payload: %w", err)
	}

	msg, err := a.nc.Request(subject, data, a.timeout)
	if err != nil {
		return fmt.Errorf("NATS request failed for gateway %s: %w", gatewayId, err)
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return fmt.Errorf("invalid NATS response for gateway %s: %w", gatewayId, err)
	}

	if response.Status != "OK" {
		return fmt.Errorf("gateway %s returned error: %s", gatewayId, response.Message)
	}

	return nil
}

func (a *GatewayCommandNATSAdapter) SendCommission(gatewayId uuid.UUID, tenantId uuid.UUID, token string) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{
		Action:   "COMMISSION",
		TenantId: tenantId.String(),
		Token:    token,
	})
}

func (a *GatewayCommandNATSAdapter) SendDecommission(gatewayId uuid.UUID) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{Action: "DECOMMISSION"})
}

func (a *GatewayCommandNATSAdapter) SendInterrupt(gatewayId uuid.UUID) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{Action: "INTERRUPT"})
}

func (a *GatewayCommandNATSAdapter) SendResume(gatewayId uuid.UUID) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{Action: "RESUME"})
}

func (a *GatewayCommandNATSAdapter) SendReset(gatewayId uuid.UUID) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{Action: "RESET"})
}

func (a *GatewayCommandNATSAdapter) SendReboot(gatewayId uuid.UUID) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{Action: "REBOOT"})
}

func (a *GatewayCommandNATSAdapter) SendSetFrequency(gatewayId uuid.UUID, frequency int) error {
	return a.sendCommand(gatewayId, gatewayCommandPayload{
		Action:    "SET_FREQUENCY",
		Frequency: frequency,
	})
}
