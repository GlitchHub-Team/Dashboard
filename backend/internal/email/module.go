package email

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"backend/internal/config"
)

func NewEmailAdapterFactory(
	cfg *config.Config,
	log *zap.Logger,
) SendEmailPort {
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
		NewEmailAdapterFactory,
	),
)
