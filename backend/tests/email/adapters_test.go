package email_test

import (
	"fmt"
	"testing"

	"backend/internal/email"
	"backend/internal/shared/config"
	"backend/tests/email/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	gomail "gopkg.in/mail.v2"
)

func TestSendEmailSMTPAdapter_SendConfirmAccountEmail(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Setup dipendenze
	mockSender := mocks.NewMocksmtpSender(ctrl)
	mockStrategy := mocks.NewMockcreateMessageStrategy(ctrl)

	targetAppUrl := "https://localhost"
	targetFrom := "from@m31.com"

	adapter := email.NewSendEmailSMTPAdapter(
		&config.Config{SMTPFrom: targetFrom, AppURL: targetAppUrl},
		mockSender,
		mockStrategy,
	)

	t.Run("Super Admin Case", func(t *testing.T) {
		targetTo := "super@example.com"
		targetToken := "token-123"

		// 1. Calcola URL atteso
		expectedUrl := targetAppUrl + "/confirm_account/token-123"
		expectedBody := fmt.Sprintf(email.CONFIRM_ACCOUNT_MAIL_TEMPLATE, expectedUrl)

		dummyMsg := &gomail.Message{} // NOTA: Struct vuoto perché non interessa mandare messaggio

		// 2. Parameter matching su strategy
		mockStrategy.EXPECT().
			CreateMessage(
				targetFrom,
				targetTo,
				"Conferma il tuo account",
				expectedBody, // NOTA: Qui viene testata la logica
			).
			Return(dummyMsg).
			Times(1)

		// 3. Controllo invio messaggio al sender
		mockSender.EXPECT().
			DialAndSend(dummyMsg).
			Return(nil).
			Times(1)

		err := adapter.SendConfirmAccountEmail(targetTo, nil, targetToken)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Tenant Member Case", func(t *testing.T) {
		targetTo := "tenant@example.com"
		targetToken := "token-456"
		targetTenantId := uuid.New()

		// 1. Calcola URL
		expectedUrl := fmt.Sprintf("%s/confirm_account/%s?tid=%s", targetAppUrl, targetToken, targetTenantId.String())
		expectedBody := fmt.Sprintf(email.CONFIRM_ACCOUNT_MAIL_TEMPLATE, expectedUrl)

		dummyMsg := &gomail.Message{}

		// 2. Parameter matching
		// NOTA: Questo assert assicura che non ci siano regressioni sull'URL
		mockStrategy.EXPECT().
			CreateMessage(
				targetFrom,
				targetTo,
				"Conferma il tuo account",
				expectedBody,
			).
			Return(dummyMsg).
			Times(1)

			// 3. Controlla invio messaggio a sender
		mockSender.EXPECT().
			DialAndSend(dummyMsg).
			Return(nil).
			Times(1)

		err := adapter.SendConfirmAccountEmail(targetTo, &targetTenantId, targetToken)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestSendEmailSMTPAdapter_SendForgotPasswordEmail(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Setup dipendenze
	mockSender := mocks.NewMocksmtpSender(ctrl)
	mockStrategy := mocks.NewMockcreateMessageStrategy(ctrl)

	targetAppUrl := "https://localhost"
	targetFrom := "from@m31.com"

	adapter := email.NewSendEmailSMTPAdapter(
		&config.Config{SMTPFrom: targetFrom, AppURL: targetAppUrl},
		mockSender,
		mockStrategy,
	)

	t.Run("Super Admin Case", func(t *testing.T) {
		targetTo := "super@example.com"
		targetToken := "token-123"

		// 1. Calcola URL atteso
		expectedUrl := fmt.Sprintf("%s/forgot_password/%s", targetAppUrl, targetToken)
		expectedBody := fmt.Sprintf(email.FORGOT_PASSWORD_MAIL_TEMPLATE, expectedUrl)

		dummyMsg := &gomail.Message{} // NOTA: Struct vuoto perché non interessa mandare messaggio

		// 2. Parameter matching su strategy
		mockStrategy.EXPECT().
			CreateMessage(
				targetFrom,
				targetTo,
				"Cambio password dimenticata richiesto",
				expectedBody, // NOTA: Qui viene testata la logica
			).
			Return(dummyMsg).
			Times(1)

		// 3. Controllo invio messaggio al sender
		mockSender.EXPECT().
			DialAndSend(dummyMsg).
			Return(nil).
			Times(1)

		err := adapter.SendForgotPasswordEmail(targetTo, nil, targetToken)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Tenant Member Case", func(t *testing.T) {
		targetTo := "tenant@example.com"
		targetToken := "token-456"
		targetTenantId := uuid.New()

		// 1. Calcola URL
		expectedUrl := fmt.Sprintf("%s/forgot_password/%s?tid=%s", targetAppUrl, targetToken, targetTenantId.String())
		expectedBody := fmt.Sprintf(email.FORGOT_PASSWORD_MAIL_TEMPLATE, expectedUrl)

		dummyMsg := &gomail.Message{}

		// 2. Parameter matching
		// NOTA: Questo assert assicura che non ci siano regressioni sull'URL
		mockStrategy.EXPECT().
			CreateMessage(
				targetFrom,
				targetTo,
				"Cambio password dimenticata richiesto",
				expectedBody,
			).
			Return(dummyMsg).
			Times(1)

			// 3. Controlla invio messaggio a sender
		mockSender.EXPECT().
			DialAndSend(dummyMsg).
			Return(nil).
			Times(1)

		err := adapter.SendForgotPasswordEmail(targetTo, &targetTenantId, targetToken)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestSendEmailTerminalAdapter_SendConfirmAccountEmail(t *testing.T) {
	// Input
	targetTo := "admin@example.com"
	targetTenantId := uuid.New()
	targetToken := "forgot-token-456"

	expectedMessage := "Invio mail di conferma account"

	type testCase struct {
		name,
		inputTo string
		inputTenantId *uuid.UUID
		inputToken    string
	}

	cases := []testCase{
		{
			name:          "Super Admin Case",
			inputTo:       targetTo,
			inputTenantId: nil,
			inputToken:    targetToken,
		},
		{
			name:          "Tenant Member Case",
			inputTo:       targetTo,
			inputTenantId: &targetTenantId,
			inputToken:    targetToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Observer zap
			core, recordedLogs := observer.New(zap.DebugLevel)
			testLogger := zap.New(core)

			// Adapter
			adapter := email.NewSendEmailTerminalAdapter(testLogger)

			// 1. Esecuzione metodo
			err := adapter.SendConfirmAccountEmail(targetTo, &targetTenantId, targetToken)
			// 2. Assert error
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// 3. Assert logs
			logs := recordedLogs.All()

			// - Check lunghezza
			if len(logs) != 1 {
				t.Fatalf("expected exactly 1 log entry, got %d", len(logs))
			}

			// - Assert struttura e contenuti
			entry := logs[0]

			if entry.Level != zap.DebugLevel {
				t.Errorf("expected log level %v, got %v", zap.DebugLevel, entry.Level)
			}

			if entry.Message != expectedMessage {
				t.Errorf("expected message %q, got %q", expectedMessage, entry.Message)
			}

			// - Assert strutturato sui campi loggati
			contextMap := entry.ContextMap()

			if contextMap["toAddr"] != tc.inputTo {
				t.Errorf("expected toAddr %q, got %v", tc.inputTo, contextMap["toAddr"])
			}

			contextTenantIdStr, ok := (contextMap["tenantId"]).(string)
			t.Logf("%v, %v", contextTenantIdStr, ok)
			if !ok {
				t.Fatalf("expected tenantId convertible to string")
			}

			if contextTenantIdStr == "" && tc.inputTenantId != nil {
				t.Errorf("expected tenantId nil, got %+#v", contextMap["tenantId"])
			} else if contextTenantIdStr != "" && tc.inputTenantId != nil && contextTenantIdStr != targetTenantId.String() {
				t.Errorf("expected tokenString %+#v, got %+#v", targetToken, contextMap["tenantId"])
			}

			if contextMap["tokenString"] != tc.inputToken {
				t.Errorf("expected tokenString %q, got %v", tc.inputToken, contextMap["tokenString"])
			}
		})
	}
}

func TestSendEmailTerminalAdapter_SendForgotPasswordEmail(t *testing.T) {
	// Input
	targetTo := "admin@example.com"
	targetTenantId := uuid.New()
	targetToken := "forgot-token-456"

	expectedMessage := "Invio mail di cambio password"

	type testCase struct {
		name,
		inputTo string
		inputTenantId *uuid.UUID
		inputToken    string
	}

	cases := []testCase{
		{
			name:          "Super Admin Case",
			inputTo:       targetTo,
			inputTenantId: nil,
			inputToken:    targetToken,
		},
		{
			name:          "Tenant Member Case",
			inputTo:       targetTo,
			inputTenantId: &targetTenantId,
			inputToken:    targetToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Observer zap
			core, recordedLogs := observer.New(zap.DebugLevel)
			testLogger := zap.New(core)

			// Adapter
			adapter := email.NewSendEmailTerminalAdapter(testLogger)

			// 1. Esecuzione metodo
			err := adapter.SendForgotPasswordEmail(targetTo, &targetTenantId, targetToken)
			// 2. Assert error
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// 3. Assert logs
			logs := recordedLogs.All()

			// - Check lunghezza
			if len(logs) != 1 {
				t.Fatalf("expected exactly 1 log entry, got %d", len(logs))
			}

			// - Assert struttura e contenuti
			entry := logs[0]

			if entry.Level != zap.DebugLevel {
				t.Errorf("expected log level %v, got %v", zap.DebugLevel, entry.Level)
			}

			if entry.Message != expectedMessage {
				t.Errorf("expected message %q, got %q", expectedMessage, entry.Message)
			}

			// - Assert strutturato sui campi loggati
			contextMap := entry.ContextMap()

			if contextMap["toAddr"] != tc.inputTo {
				t.Errorf("expected toAddr %q, got %v", tc.inputTo, contextMap["toAddr"])
			}

			contextTenantIdStr, ok := (contextMap["tenantId"]).(string)
			t.Logf("%v, %v", contextTenantIdStr, ok)
			if !ok {
				t.Fatalf("expected tenantId convertible to string")
			}

			if contextTenantIdStr == "" && tc.inputTenantId != nil {
				t.Errorf("expected tenantId nil, got %+#v", contextMap["tenantId"])
			} else if contextTenantIdStr != "" && tc.inputTenantId != nil && contextTenantIdStr != targetTenantId.String() {
				t.Errorf("expected tokenString %+#v, got %+#v", targetToken, contextMap["tenantId"])
			}

			if contextMap["tokenString"] != tc.inputToken {
				t.Errorf("expected tokenString %q, got %v", tc.inputToken, contextMap["tokenString"])
			}
		})
	}
}
