package tenant

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tenant",
	fx.Provide(
		NewTenantController,
		NewCreateTenantService,
		NewTenantPostgreAdapter,
	),
	fx.Provide(
		fx.Private,
		NewTenantPostgreRepository,
	),
)
