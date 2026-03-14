package email

import (
	"backend/internal/common"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"email",
	fx.Provide(
		common.FxAs[SendEmailPort](NewSendEmailMailtrapAdapter),
		common.FxAs[SendEmailPort](NewSendEmailTerminalAdapter),
	),
)
