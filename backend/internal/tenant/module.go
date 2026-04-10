package tenant

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tenant",
	fx.Provide(
		NewTenantController,

		fx.Annotate(
			NewCreateTenantService,
			fx.As(new(CreateTenantUseCase)),
			fx.As(new(DeleteTenantUseCase)),
			fx.As(new(GetTenantUseCase)),
			fx.As(new(GetTenantListUseCase)),
			fx.As(new(GetAllTenantsUseCase)),
		),
		
		fx.Annotate(
			NewTenantPostgreAdapter,
			fx.As(new(CreateTenantPort)),
			fx.As(new(DeleteTenantPort)),
			fx.As(new(GetTenantPort)),
			fx.As(new(GetTenantsPort)),
		),
		
		fx.Annotate(
			NewTenantPostgreRepository,
			fx.As(new(TenantRepository)),
		),
	),
	fx.Provide(
		fx.Private,
	),
)
