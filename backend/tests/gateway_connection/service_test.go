package gateway_connection_test

import (
	"errors"
	"testing"

	"backend/internal/gateway"
	"backend/internal/gateway_connection"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func strPtr(v string) *string {
	return &v
}

// mockGetGatewayPort implementa gateway.GetGatewayPort
type mockGetGatewayPort struct {
	result       gateway.Gateway
	err          error
	calledGetBy  bool
	calledGetAll bool
	calledGetByT bool
}

func (m *mockGetGatewayPort) GetById(id uuid.UUID) (gateway.Gateway, error) {
	m.calledGetBy = true
	return m.result, m.err
}

func (m *mockGetGatewayPort) GetGatewayByTenantID(tenantId uuid.UUID, gatewayId uuid.UUID) (gateway.Gateway, error) {
	m.calledGetByT = true
	return m.result, m.err
}

func (m *mockGetGatewayPort) GetByTenantId(tenantId string) ([]gateway.Gateway, error) {
	m.calledGetByT = true
	return nil, m.err
}

func (m *mockGetGatewayPort) GetAll() ([]gateway.Gateway, error) {
	m.calledGetAll = true
	return nil, m.err
}

// mockSaveGatewayPort implementa gateway.SaveGatewayPort
type mockSaveGatewayPort struct {
	received gateway.Gateway
	err      error
	called   bool
}

func (m *mockSaveGatewayPort) Save(g gateway.Gateway) (gateway.Gateway, error) {
	m.called = true
	m.received = g
	if m.err != nil {
		return gateway.Gateway{}, m.err
	}
	return g, nil
}

func TestProcessHello_InvalidUUID(t *testing.T) {
	logger := zap.NewNop()
	get := &mockGetGatewayPort{}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: "not-a-uuid", PublicIdentifier: "p"}
	err := svc.ProcessHello(msg)
	if err == nil {
		t.Fatalf("expected error for invalid uuid, got nil")
	}
}

func TestProcessHello_MissingPublicIdentifier_Nak(t *testing.T) {
	logger := zap.NewNop()
	get := &mockGetGatewayPort{}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: uuid.New().String(), PublicIdentifier: ""}
	err := svc.ProcessHello(msg)
	if err == nil {
		t.Fatalf("expected error when public identifier is missing, got nil")
	}
	if !errors.Is(err, gateway_connection.ErrPublicIdentifierRequired) {
		t.Fatalf("expected ErrPublicIdentifierRequired, got %v", err)
	}
	if get.calledGetBy {
		t.Fatalf("did not expect GetById to be called when public identifier is missing")
	}
	if save.called {
		t.Fatalf("did not expect Save to be called when public identifier is missing")
	}
}

func TestProcessHello_GatewayNotFound_Nak(t *testing.T) {
	logger := zap.NewNop()
	// adapter returns zero-value Gateway and nil error when not found
	get := &mockGetGatewayPort{result: gateway.Gateway{}, err: nil}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: uuid.New().String(), PublicIdentifier: "p"}
	err := svc.ProcessHello(msg)
	if err == nil {
		t.Fatalf("expected error when gateway not found, got nil")
	}
	if !errors.Is(err, gateway.ErrGatewayNotFound) {
		t.Fatalf("expected gateway.ErrGatewayNotFound, got %v", err)
	}
	if !get.calledGetBy {
		t.Fatalf("expected GetById to be called")
	}
	if save.called {
		t.Fatalf("did not expect Save to be called when gateway not found")
	}
}

func TestProcessHello_GetByIdError_Nak(t *testing.T) {
	logger := zap.NewNop()
	get := &mockGetGatewayPort{err: errors.New("db down")}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: uuid.New().String(), PublicIdentifier: "p"}
	err := svc.ProcessHello(msg)
	if err == nil {
		t.Fatalf("expected error when GetById fails, got nil")
	}
}

func TestProcessHello_UpdatePublicIdentifier_SaveCalled(t *testing.T) {
	logger := zap.NewNop()
	id := uuid.New()
	existing := gateway.Gateway{
		Id:               id,
		Name:             "gw",
		Status:           gateway.GATEWAY_STATUS_INACTIVE,
		PublicIdentifier: strPtr("old-id"),
	}
	get := &mockGetGatewayPort{result: existing, err: nil}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: id.String(), PublicIdentifier: "new-id"}
	err := svc.ProcessHello(msg)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !save.called {
		t.Fatalf("expected Save to be called")
	}
	if save.received.PublicIdentifier == nil || *save.received.PublicIdentifier != "new-id" {
		t.Fatalf("expected saved public identifier to be new-id")
	}
	if save.received.Status != gateway.GATEWAY_STATUS_INACTIVE {
		t.Fatalf("expected status to remain unchanged, got %v", save.received.Status)
	}
}

func TestProcessHello_PublicIdentifierNil_SaveCalled(t *testing.T) {
	logger := zap.NewNop()
	id := uuid.New()
	existing := gateway.Gateway{
		Id:               id,
		Name:             "gw",
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		PublicIdentifier: nil,
	}
	get := &mockGetGatewayPort{result: existing, err: nil}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: id.String(), PublicIdentifier: "new-id"}
	err := svc.ProcessHello(msg)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !save.called {
		t.Fatalf("expected Save to be called when PublicIdentifier is nil")
	}
	if save.received.PublicIdentifier == nil || *save.received.PublicIdentifier != "new-id" {
		t.Fatalf("expected saved public identifier to be new-id")
	}
}

func TestProcessHello_PublicIdentifierUnchanged_NoSave(t *testing.T) {
	logger := zap.NewNop()
	id := uuid.New()
	existing := gateway.Gateway{
		Id:               id,
		Name:             "gw",
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		PublicIdentifier: strPtr("same-id"),
	}
	get := &mockGetGatewayPort{result: existing, err: nil}
	save := &mockSaveGatewayPort{}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: id.String(), PublicIdentifier: "same-id"}
	err := svc.ProcessHello(msg)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if save.called {
		t.Fatalf("did not expect Save to be called when PublicIdentifier is unchanged")
	}
}

func TestProcessHello_SaveError_Nak(t *testing.T) {
	logger := zap.NewNop()
	id := uuid.New()
	existing := gateway.Gateway{
		Id:               id,
		Name:             "gw",
		Status:           gateway.GATEWAY_STATUS_INACTIVE,
		PublicIdentifier: nil,
	}
	get := &mockGetGatewayPort{result: existing, err: nil}
	save := &mockSaveGatewayPort{err: errors.New("save failed")}
	svc := gateway_connection.NewGatewayHelloService(get, save, logger)

	msg := gateway_connection.GatewayHelloMessage{GatewayId: id.String(), PublicIdentifier: "new-id"}
	err := svc.ProcessHello(msg)
	if err == nil {
		t.Fatalf("expected error when Save fails, got nil")
	}
}
