package email

import (
	"backend/internal/auth"
	"backend/internal/shared/config"
	"backend/internal/user"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewEmailAdapterFactory(
	cfg *config.Config,
	sender smtpSender,
	msgStrategy createMessageStrategy,
	log *zap.Logger,
) (SendEmailPort, error) {
	// NOTA: Si può avere solo un email sender alla volta, quindi è importante usare questo
	// factory per crearlo
	// Bisogna inserire tutte le dipendenze di tutti i possibili adapter
	switch cfg.MailAdapter {
	case "terminal":
		return NewSendEmailTerminalAdapter(log), nil
	case "smtp":
		return NewSendEmailSMTPAdapter(cfg, sender, msgStrategy), nil
	default:
		return nil, fmt.Errorf("mail adapter '%v' does not exist", cfg.MailAdapter)
	}
}

var Module = fx.Module(
	"email",
	fx.Provide(
		fx.Annotate(
			NewEmailAdapterFactory,
			fx.As(new(user.SendConfirmAccountEmailPort)),
			fx.As(new(auth.SendForgotPasswordEmailPort)),
		),
	),
	fx.Provide(
		fx.Private,

		// Dialer SMTP
		fx.Annotate(
			newDialer,
			fx.As(new(smtpSender)),
		),

		// NOTA: Qui si può cambiare strategia di creazione messaggi mail
		fx.Annotate(
			newPlainTextMessageStrategy,
			fx.As(new(createMessageStrategy)),
		),
	),
)
