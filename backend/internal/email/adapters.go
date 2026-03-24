package email

import (
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/email/mocks/ports.go -package=mocks . SendEmailPort

type SendEmailPort interface {
	SendConfirmAccountEmail(toAddr string, token string) error
	SendChangePasswordEmail(toAddr string, token string) error
}

// ------------------------------------------------------------------------------------------------------

type SendEmailMailtrapAdapter struct{}

func NewSendEmailMailtrapAdapter() *SendEmailMailtrapAdapter {
	return &SendEmailMailtrapAdapter{}
}

func (adapter *SendEmailMailtrapAdapter) SendConfirmAccountEmail(toAddr string, token string) error {
	return nil
}

func (adapter *SendEmailMailtrapAdapter) SendChangePasswordEmail(toAddr string, token string) error {
	return nil
}

// Compile-time checks
var _ SendEmailPort = (*SendEmailMailtrapAdapter)(nil)

// ------------------------------------------------------------------------------------------------------
type SendEmailTerminalAdapter struct {
	log *zap.Logger
}

func NewSendEmailTerminalAdapter(log *zap.Logger) *SendEmailTerminalAdapter {
	return &SendEmailTerminalAdapter{log: log}
}

func (adapter *SendEmailTerminalAdapter) SendConfirmAccountEmail(toAddr string, token string) error {
	adapter.log.Debug(
		"Invio mail di conferma account",
		zap.String("to", toAddr),
		zap.String("token", token),
	)
	return nil
}

func (adapter *SendEmailTerminalAdapter) SendChangePasswordEmail(toAddr string, token string) error {
	adapter.log.Debug(
		"Invio mail di cambio password",
		zap.String("to", toAddr),
		zap.String("token", token),
	)
	return nil
}

// Compile-time checks
var _ SendEmailPort = (*SendEmailTerminalAdapter)(nil)
