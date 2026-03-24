package email

import (
	"backend/internal/auth"
	"backend/internal/shared/config"
	"backend/internal/user"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewEmailAdapterFactory(
	cfg *config.Config,
	log *zap.Logger,
) (
	SendEmailPort,
) {
	// NOTA: Si può avere solo un email sender alla vota, quindi è importante usare questo
	// factory per crearlo
	// Bisogna inserire tutte le dipendenze di tutti i possibili adapter
	if cfg.MailAdapter == "terminal" {
		return NewSendEmailTerminalAdapter(log)
	}

	return NewSendEmailMailtrapAdapter()
}

var Module = fx.Module(
	"email",
	fx.Provide(
		fx.Annotate(
			NewEmailAdapterFactory,
			fx.As(new(user.SendConfirmAccountEmailPort)),
			fx.As(new(auth.SendChangePasswordEmailPort)),
		),
	),
)
