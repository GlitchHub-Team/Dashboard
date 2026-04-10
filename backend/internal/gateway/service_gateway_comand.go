package gateway

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type GatewayCommandPort interface {
	SendCommission(gatewayId uuid.UUID, tenantId uuid.UUID, token string) error
	SendDecommission(gatewayId uuid.UUID) error
	SendInterrupt(gatewayId uuid.UUID) error
	SendResume(gatewayId uuid.UUID) error
	SendReset(gatewayId uuid.UUID) error
	SendReboot(gatewayId uuid.UUID) error
	SendSetFrequency(gatewayId uuid.UUID, frequency int) error
}

type GatewayCommandService struct {
	getGatewayPort     GetGatewayPort
	saveGatewayPort    SaveGatewayPort
	gatewayCommandPort GatewayCommandPort
}

func NewGatewayCommandService(
	getGatewayPort GetGatewayPort,
	saveGatewayPort SaveGatewayPort,
	gatewayCommandPort GatewayCommandPort,
) *GatewayCommandService {
	return &GatewayCommandService{
		getGatewayPort:     getGatewayPort,
		saveGatewayPort:    saveGatewayPort,
		gatewayCommandPort: gatewayCommandPort,
	}
}

var BackendJWTSecret = []byte("aaaaaaa")

func GenerateCommissionedToken(gatewayId uuid.UUID, tenantId uuid.UUID, gatewaySecret []byte) (string, error) {
	claims := jwt.MapClaims{
		"tenant_id":  tenantId.String(),
		"gateway_id": gatewayId.String(),
		"role":       "gateway_device",
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(gatewaySecret)
}

func (s *GatewayCommandService) CommissionGateway(command CommissionGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.Status == GATEWAY_STATUS_COMMISSIONED {
		return Gateway{}, ErrGatewayAlreadyCommissioned
	}

	if gw.SigningSecret == "" {
		return Gateway{}, ErrMissingGatewaySecret
	}

	tokenString, err := GenerateCommissionedToken(
		gw.Id,
		command.TenantId,
		[]byte(gw.SigningSecret),
	)
	if err != nil {
		return Gateway{}, err
	}

	err = s.gatewayCommandPort.SendCommission(gw.Id, command.TenantId, tokenString)
	if err != nil {
		return Gateway{}, ErrComunicationNats
	}

	gw.Status = GATEWAY_STATUS_COMMISSIONED
	gw.TenantId = &command.TenantId

	savedGw, err := s.saveGatewayPort.Save(gw)

	if err != nil {
		return Gateway{}, err
	}

	savedGw.SigningSecret = tokenString
	return savedGw, nil
}

func (s *GatewayCommandService) DecommissionGateway(command DecommissionGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.Status == "DECOMMISSIONED" {
		return Gateway{}, ErrGatewayAlreadyDecommissioned
	}
	err = s.gatewayCommandPort.SendDecommission(gw.Id)
	if err != nil {
		return Gateway{}, ErrComunicationNats
	}

	gw.Status = "DECOMMISSIONED"
	gw.TenantId = nil

	savedGw, err := s.saveGatewayPort.Save(gw)
	if err != nil {
		return Gateway{}, err
	}

	return savedGw, nil
}

func (s *GatewayCommandService) InterruptGateway(command InterruptGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.Status != "COMMISSIONED" {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	err = s.gatewayCommandPort.SendInterrupt(gw.Id)
	if err != nil {
		return Gateway{}, err
	}

	gw.Status = "INTERRUPTED"
	savedGw, err := s.saveGatewayPort.Save(gw)
	if err != nil {
		return Gateway{}, err
	}

	return savedGw, nil
}
func (s *GatewayCommandService) ResumeGateway(command ResumeGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	if gw.Status != GATEWAY_STATUS_INTERRUPTED {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	if err := s.gatewayCommandPort.SendResume(gw.Id); err != nil {
		return Gateway{}, ErrComunicationNats
	}

	gw.Status = GATEWAY_STATUS_COMMISSIONED
	savedGw, err := s.saveGatewayPort.Save(gw)
	if err != nil {
		return Gateway{}, err
	}

	return savedGw, nil
}

func (s *GatewayCommandService) ResetGateway(command ResetGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	if gw.Status != GATEWAY_STATUS_COMMISSIONED {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	if err := s.gatewayCommandPort.SendReset(gw.Id); err != nil {
		return Gateway{}, ErrComunicationNats
	}

	return gw, nil
}

func (s *GatewayCommandService) RebootGateway(command RebootGatewayCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	if gw.Status != GATEWAY_STATUS_COMMISSIONED {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	if err := s.gatewayCommandPort.SendReboot(gw.Id); err != nil {
		return Gateway{}, ErrComunicationNats
	}

	return gw, nil
}

func (s *GatewayCommandService) SetGatewayIntervalLimit(command SetGatewayIntervalLimitCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	if gw.Status != GATEWAY_STATUS_COMMISSIONED {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	if err := s.gatewayCommandPort.SendSetFrequency(gw.Id, command.IntervalLimit); err != nil {
		return Gateway{}, ErrComunicationNats
	}

	gw.IntervalLimit = int64(command.IntervalLimit)
	savedGw, err := s.saveGatewayPort.Save(gw)
	if err != nil {
		return Gateway{}, err
	}

	return savedGw, nil
}
