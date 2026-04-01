package email

import gomail "gopkg.in/mail.v2"

const (
	CONFIRM_ACCOUNT_MAIL_TEMPLATE = `Conferma ora il tuo account cliccando il seguente link:
%s`
	FORGOT_PASSWORD_MAIL_TEMPLATE = `È stato richiesto un cambio password per il tuo account.
Clicca il seguente link per cambiare la password: %s

Se non sei stato tu a richiedere il cambio password, puoi ignorare questo messaggio.`
)

//go:generate mockgen -destination=../../tests/email/mocks/message_strategy.go -package=mocks . createMessageStrategy

/*
Interfaccia da rispettare per creare messaggi mail in maniere diverse (plain, html, etc.)
*/
type createMessageStrategy interface {
	/*
		Crea un messaggio email inviato da indirizzo fromAddress a indirizzo toAddress con oggetto subject
		e corpo body.
	*/
	CreateMessage(fromAddress, toAddress, subject, body string) *gomail.Message
}

type plainTextMessageStrategy struct{}

var _ createMessageStrategy = (*plainTextMessageStrategy)(nil) // Compile-time check

func newPlainTextMessageStrategy() *plainTextMessageStrategy {
	return &plainTextMessageStrategy{}
}

func (c *plainTextMessageStrategy) CreateMessage(fromAddress, toAddress, subject, body string) (
	message *gomail.Message,
) {
	message = gomail.NewMessage()
	message.SetHeader("From", fromAddress)
	message.SetHeader("To", toAddress)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return
}
