package hello_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"backend/internal/gateway/hello"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// --- mock service semplice ---
type mockService struct {
	shouldErr bool
	called    bool
	received  hello.GatewayHelloMessageCommand
}

func (m *mockService) ProcessHello(msg hello.GatewayHelloMessageCommand) error {
	m.called = true
	m.received = msg
	if m.shouldErr {
		return errors.New("service error")
	}
	return nil
}

// --- mockMsgSimple: solo i 4 metodi che ci interessano ---
type mockMsgSimple struct {
	data   []byte
	acked  bool
	nacked bool
	termed bool

	ackErr  error
	nakErr  error
	termErr error
}

func (m *mockMsgSimple) Data() []byte { return m.data }
func (m *mockMsgSimple) Ack() error   { m.acked = true; return m.ackErr }
func (m *mockMsgSimple) Nak() error   { m.nacked = true; return m.nakErr }
func (m *mockMsgSimple) Term() error  { m.termed = true; return m.termErr }

// --- wrapper snello che implementa jetstream.Msg delegando a mockMsgSimple ---
type testJetMsg struct {
	inner *mockMsgSimple
}

func newTestJetMsgFromSimple(s *mockMsgSimple) *testJetMsg {
	return &testJetMsg{inner: s}
}

// metodi usati dalla logica del worker (delegati)
func (t *testJetMsg) Data() []byte { return t.inner.Data() }
func (t *testJetMsg) Ack() error   { return t.inner.Ack() }
func (t *testJetMsg) Nak() error   { return t.inner.Nak() }
func (t *testJetMsg) Term() error  { return t.inner.Term() }

// metodi richiesti dall'interfaccia jetstream.Msg (stubs minimi)
func (t *testJetMsg) Reply() string                             { return "" }
func (t *testJetMsg) Subject() string                           { return "" }
func (t *testJetMsg) DoubleAck(ctx context.Context) error       { return nil }
func (t *testJetMsg) Headers() nats.Header                      { return nil }
func (t *testJetMsg) InProgress() error                         { return nil }
func (t *testJetMsg) Metadata() (*jetstream.MsgMetadata, error) { return &jetstream.MsgMetadata{}, nil }
func (t *testJetMsg) NakWithDelay(delay time.Duration) error    { return nil }
func (t *testJetMsg) TermWithReason(reason string) error        { return nil }

// --- TESTS ---

func TestProcessMsg_MalformedJSON_TermCalled(t *testing.T) {
	logger := zap.NewNop()
	svc := &mockService{shouldErr: false}
	worker := hello.NewNATSWorker(jetstream.Consumer(nil), svc, logger)

	simple := &mockMsgSimple{data: []byte("not-json")}
	msg := newTestJetMsgFromSimple(simple)

	worker.ProcessMsg(msg)

	if !simple.termed {
		t.Fatalf("expected Term to be called for malformed payload")
	}
	if svc.called {
		t.Fatalf("did not expect service.ProcessHello to be called for malformed payload")
	}
}

func TestProcessMsg_ServiceError_NakCalled(t *testing.T) {
	logger := zap.NewNop()
	svc := &mockService{shouldErr: true}
	worker := hello.NewNATSWorker(jetstream.Consumer(nil), svc, logger)

	hello := hello.GatewayHelloMessageDTO{
		GatewayId:        "00000000-0000-0000-0000-000000000000",
		PublicIdentifier: "pub",
	}
	b, _ := json.Marshal(hello)

	simple := &mockMsgSimple{data: b}
	msg := newTestJetMsgFromSimple(simple)

	worker.ProcessMsg(msg)

	if !svc.called {
		t.Fatalf("expected service.ProcessHello to be called")
	}
	if !simple.nacked {
		t.Fatalf("expected Nak to be called when service returns error")
	}
	if simple.acked {
		t.Fatalf("did not expect Ack after Nak")
	}
}

func TestProcessMsg_Success_AckCalled(t *testing.T) {
	logger := zap.NewNop()
	svc := &mockService{shouldErr: false}
	worker := hello.NewNATSWorker(jetstream.Consumer(nil), svc, logger)

	hello := hello.GatewayHelloMessageDTO{
		GatewayId:        "00000000-0000-0000-0000-000000000000",
		PublicIdentifier: "pub",
	}
	b, _ := json.Marshal(hello)

	simple := &mockMsgSimple{data: b}
	msg := newTestJetMsgFromSimple(simple)

	worker.ProcessMsg(msg)

	if !svc.called {
		t.Fatalf("expected service.ProcessHello to be called")
	}
	if !simple.acked {
		t.Fatalf("expected Ack to be called on success")
	}
	if simple.nacked || simple.termed {
		t.Fatalf("did not expect Nak or Term on success")
	}
}
