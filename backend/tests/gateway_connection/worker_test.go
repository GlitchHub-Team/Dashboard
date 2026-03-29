package gateway_connection_test

import (
	"errors"
	"testing"

	"backend/internal/gateway_connection"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// --- MOCK DEL MESSAGGIO JETSTREAM ---
type mockJetStreamMsg struct {
	jetstream.Msg
	data       []byte
	ackCalled  bool
	nakCalled  bool
	termCalled bool
}

func (m *mockJetStreamMsg) Data() []byte { return m.data }
func (m *mockJetStreamMsg) Ack() error   { m.ackCalled = true; return nil }
func (m *mockJetStreamMsg) Nak() error   { m.nakCalled = true; return nil }
func (m *mockJetStreamMsg) Term() error  { m.termCalled = true; return nil }

// --- MOCK DEL SERVIZIO ---
type mockHelloService struct {
	errToReturn error
}

func (m *mockHelloService) ProcessHello(msg gateway_connection.GatewayHelloMessage) error {
	return m.errToReturn
}

// --- TEST CASE ---
func TestNATSWorker_ProcessMsg(t *testing.T) {
	cases := []struct {
		name       string
		data       []byte
		serviceErr error
		expectAck  bool
		expectNak  bool
		expectTerm bool
	}{
		{
			name:       "Success: valid message",
			data:       []byte(`{"gateway_id":"001"}`),
			serviceErr: nil,
			expectAck:  true,
		},
		{
			name:       "Error: service fails",
			data:       []byte(`{"gateway_id":"001"}`),
			serviceErr: errors.New("db error"),
			expectNak:  true,
		},
		{
			name:       "Error: invalid JSON",
			data:       []byte(`{ broken-json }`),
			serviceErr: nil,
			expectTerm: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockSvc := &mockHelloService{errToReturn: tc.serviceErr}
			worker := gateway_connection.NewNATSWorker(nil, mockSvc, zap.NewNop())
			msg := &mockJetStreamMsg{data: tc.data}

			// Esecuzione
			worker.ProcessMsg(msg)

			// Verifiche
			if tc.expectAck {
				require.True(t, msg.ackCalled, "Dovrebbe aver chiamato Ack()")
			}
			if tc.expectNak {
				require.True(t, msg.nakCalled, "Dovrebbe aver chiamato Nak()")
			}
			if tc.expectTerm {
				require.True(t, msg.termCalled, "Dovrebbe aver chiamato Term()")
			}
		})
	}
}
