package email_test

import (
	"testing"

	"backend/internal/email"
	"backend/internal/shared/config"
	"backend/tests/email/mocks"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestNewEmailAdapterFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mocks.NewMocksmtpSender(ctrl)
	mockMsgStrategy := mocks.NewMockcreateMessageStrategy(ctrl)
	mockLogger := zaptest.NewLogger(t)

	t.Run("Case terminal", func(t *testing.T) {
		port, err := email.NewEmailAdapterFactory(
			&config.Config{MailAdapter: "terminal"},
			mockSender,
			mockMsgStrategy,
			mockLogger,
		)

		if _, ok := port.(*email.SendEmailTerminalAdapter); !ok {
			t.Errorf("want port with conrete type email.SendEmailTerminalAdapter")
		}
		if err != nil {
			t.Errorf("want nil error, got: %v", err)
		}
	})

	t.Run("Case smtp", func(t *testing.T) {
		port, err := email.NewEmailAdapterFactory(
			&config.Config{MailAdapter: "smtp"},
			mockSender,
			mockMsgStrategy,
			mockLogger,
		)

		if _, ok := port.(*email.SendEmailSMTPAdapter); !ok {
			t.Errorf("Expected port to be of conrete type email.SendEmailTerminalAdapter")
		}
		if err != nil {
			t.Errorf("want nil error, got: %v", err)
		}
	})

	t.Run("Case unknown", func(t *testing.T) {
		_, err := email.NewEmailAdapterFactory(
			&config.Config{MailAdapter: "invalid-value"},
			mockSender,
			mockMsgStrategy,
			mockLogger,
		)

		if err == nil {
			t.Errorf("want error, got: %v", err)
		}
	})
}
