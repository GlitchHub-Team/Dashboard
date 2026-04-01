package email

import (
	"fmt"

	"backend/internal/shared/config"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/email/mocks/ports.go -package=mocks . SendEmailPort

type SendEmailPort interface {
	SendConfirmAccountEmail(toAddr string, tenantId *uuid.UUID, tokenString string) error
	SendForgotPasswordEmail(toAddr string, tenantId *uuid.UUID, tokenString string) error
}

// ------------------------------------------------------------------------------------------------------

type SendEmailSMTPAdapter struct {
	fromAddress       string
	sender            smtpSender
	createMsgStrategy createMessageStrategy
	appUrl            string
}

var _ SendEmailPort = (*SendEmailSMTPAdapter)(nil)

func NewSendEmailSMTPAdapter(
	cfg *config.Config, sender smtpSender, createMsgStrategy createMessageStrategy,
) *SendEmailSMTPAdapter {
	return &SendEmailSMTPAdapter{
		fromAddress:       cfg.SMTPFrom,
		sender:            sender,
		createMsgStrategy: createMsgStrategy,
		appUrl:            cfg.AppURL,
	}
}

// TODO: Da vedere con front-end come organizzare questa parte
func (adapter *SendEmailSMTPAdapter) SendConfirmAccountEmail(toAddress string, tenantId *uuid.UUID, tokenString string) (err error) {
	confirmAccountUrl := fmt.Sprintf("%s/confirm_account/%s", adapter.appUrl, tokenString)
	if tenantId != nil {
		confirmAccountUrl += fmt.Sprintf("?tid=%s", tenantId.String())
	}

	// 1. Crea messaggio
	msg := adapter.createMsgStrategy.CreateMessage(
		adapter.fromAddress,
		toAddress,
		"Conferma il tuo account",
		fmt.Sprintf(CONFIRM_ACCOUNT_MAIL_TEMPLATE, confirmAccountUrl),
	)

	// 2. Invia messaggio
	err = adapter.sender.DialAndSend(msg)
	return
}

// TODO: Da vedere con front-end come organizzare questa parte
func (adapter *SendEmailSMTPAdapter) SendForgotPasswordEmail(toAddress string, tenantId *uuid.UUID, tokenString string) (err error) {
	confirmAccountUrl := fmt.Sprintf("%s/forgot_password/%s", adapter.appUrl, tokenString)
	if tenantId != nil {
		confirmAccountUrl += fmt.Sprintf("?tid=%s", tenantId.String())
	}

	// 1. Crea messaggio
	msg := adapter.createMsgStrategy.CreateMessage(
		adapter.fromAddress,
		toAddress,
		"Cambio password dimenticata richiesto",
		fmt.Sprintf(FORGOT_PASSWORD_MAIL_TEMPLATE, confirmAccountUrl),
	)

	// 2. Invia messaggio
	err = adapter.sender.DialAndSend(msg)
	return
}

// ------------------------------------------------------------------------------------------------------
type SendEmailTerminalAdapter struct {
	log *zap.Logger
}

var _ SendEmailPort = (*SendEmailTerminalAdapter)(nil) // Compile-time checks

func NewSendEmailTerminalAdapter(log *zap.Logger) *SendEmailTerminalAdapter {
	return &SendEmailTerminalAdapter{log: log}
}

func (adapter *SendEmailTerminalAdapter) SendConfirmAccountEmail(toAddr string, tenantId *uuid.UUID, tokenString string) error {
	var tenantIdStr string
	if tenantId != nil {
		tenantIdStr = tenantId.String()
	}
	adapter.log.Debug(
		"Invio mail di conferma account",
		zap.String("toAddr", toAddr),
		zap.Any("tenantId", tenantIdStr),
		zap.Any("tokenString", tokenString),
	)
	return nil
}

func (adapter *SendEmailTerminalAdapter) SendForgotPasswordEmail(toAddr string, tenantId *uuid.UUID, tokenString string) error {
	var tenantIdStr string
	if tenantId != nil {
		tenantIdStr = tenantId.String()
	}
	adapter.log.Debug(
		"Invio mail di cambio password",
		zap.String("toAddr", toAddr),
		zap.Any("tenantId", tenantIdStr),
		zap.Any("tokenString", tokenString),
	)
	return nil
}
