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
		NewTenantPostgreRepository,
	),
	fx.Provide(
		fx.Private,
	),
)
