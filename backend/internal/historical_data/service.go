package historical_data

import (
	"backend/internal/shared/identity"
	"backend/internal/tenant"
)

type GetHistoricalDataService struct {
	getHistoricalDataPort GetHistoricalDataPort
	getTenantPort         tenant.GetTenantPort
}

func NewGetHistoricalDataService(
	getHistoricalDataPort GetHistoricalDataPort,
	getTenantPort tenant.GetTenantPort,
) *GetHistoricalDataService {
	return &GetHistoricalDataService{
		getHistoricalDataPort: getHistoricalDataPort,
		getTenantPort:         getTenantPort,
	}
}

func (service *GetHistoricalDataService) GetSensorHistoricalData(
	cmd GetSensorHistoricalDataCommand,
) ([]HistoricalSample, error) {
	if (cmd.From == nil) != (cmd.To == nil) {
		return nil, ErrInvalidDateRange
	}

	if cmd.From != nil && cmd.To != nil && cmd.From.After(*cmd.To) {
		return nil, ErrInvalidDateRange
	}

	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return nil, err
	}
	if tenantFound.IsZero() {
		return nil, tenant.ErrTenantNotFound
	}

	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate
	tenantAdminAccess := cmd.Requester.CanTenantAdminAccess(cmd.TenantId)
	tenantUserAccess := cmd.Requester.CanTenantUserAccess(cmd.TenantId)

	if !superAdminAccess && !tenantAdminAccess && !tenantUserAccess {
		return nil, identity.ErrUnauthorizedAccess
	}

	filter := HistoricalDataFilter{
		From:  cmd.From,
		To:    cmd.To,
		Limit: cmd.Limit,
	}.Normalize()

	return service.getHistoricalDataPort.GetSensorHistoricalData(cmd.TenantId, cmd.SensorId, filter)
}

var _ GetSensorHistoricalDataUseCase = (*GetHistoricalDataService)(nil)
