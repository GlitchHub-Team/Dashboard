package email

import (
	gomail "gopkg.in/mail.v2"

	"backend/internal/shared/config"
)

//go:generate mockgen -destination=../../tests/email/mocks/sender.go -package=mocks . smtpSender

/*
Interfaccia usata per astrarre gomail.Dialer in modo tale da rendere testabile il mail adapter
*/
type smtpSender interface {
	DialAndSend(m ...*gomail.Message) error
}

func newDialer(cfg *config.Config) *gomail.Dialer {
	return gomail.NewDialer(
		cfg.SMTPHost,
		int(cfg.SMTPPort),
		cfg.SMTPUser,
		cfg.SMTPPass,
	)
}
