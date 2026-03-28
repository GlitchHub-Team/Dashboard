package gateway_connection_test

import (
	"encoding/json"
	"errors"
	"testing"

	"backend/internal/gateway_connection"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- MOCK MANUALE (Per non dipendere da mockgen) ---
type mockHelloService struct {
	calledWith  gateway_connection.GatewayHelloMessage
	errToReturn error
}

func (m *mockHelloService) ProcessHello(msg gateway_connection.GatewayHelloMessage) error {
	m.calledWith = msg
	return m.errToReturn
}

// --- TEST CASE ---
func TestNATSWorker_ProcessMsg(t *testing.T) {
	cases := []struct {
		name          string
		payload       gateway_connection.GatewayHelloMessage
		serviceErr    error
		expectedError bool
	}{
		{
			name: "Success: valid message",
			payload: gateway_connection.GatewayHelloMessage{
				GatewayId:        "00000000-0000-0000-0000-000000000001",
				PublicIdentifier: "GW-001",
			},
			serviceErr:    nil,
			expectedError: false,
		},
		{
			name: "Error: service fails",
			payload: gateway_connection.GatewayHelloMessage{
				GatewayId: "00000000-0000-0000-0000-000000000001",
			},
			serviceErr:    errors.New("service error"), // Un errore qualsiasi
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockSvc := &mockHelloService{errToReturn: tc.serviceErr}
			logger := zap.NewNop()

			// NewNATSWorker(js, service, logger)
			worker := gateway_connection.NewNATSWorker(nil, mockSvc, logger)

			// Prepariamo il messaggio finto (Niente Docker!)
			data, _ := json.Marshal(tc.payload)
			msg := &nats.Msg{
				Data: data,
				Sub:  nil, // Questo fa saltare l'Ack() grazie alla tua modifica
			}

			// Esecuzione
			err := worker.ProcessMsg(msg)

			// Verifiche
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.payload.PublicIdentifier, mockSvc.calledWith.PublicIdentifier)
			}
		})
	}
}
