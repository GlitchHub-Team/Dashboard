package tenant

import (
	"encoding/json"
	"testing"

	httpdto "backend/internal/infra/transport/http/dto"

	"github.com/google/uuid"
)

func TestCreateTenantDTOUnmarshalSnakeCaseCanImpersonate(t *testing.T) {
	body := []byte(`{"tenant_name":"Acme","can_impersonate":true}`)

	var request CreateTenantDTO
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if request.TenantName != "Acme" {
		t.Fatalf("expected tenant_name to be Acme, got %q", request.TenantName)
	}

	if !request.CanImpersonate {
		t.Fatalf("expected can_impersonate to be true")
	}
}

func TestCreateTenantDTOUnmarshalLegacyCanImpersonateKeyIsIgnored(t *testing.T) {
	body := []byte(`{"tenant_name":"Acme","canimpersonate":true}`)

	var request CreateTenantDTO
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if request.CanImpersonate {
		t.Fatalf("expected can_impersonate to stay false when legacy key is used")
	}
}

func TestTenantResponseDTOMarshalUsesSnakeCase(t *testing.T) {
	tenantID := uuid.NewString()

	response := TenantResponseDTO{
		TenantIdField:   httpdto.TenantIdField{TenantId: tenantID},
		TenantNameField: httpdto.TenantNameField{TenantName: "Acme"},
		CanImpersonate:  true,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if value, ok := body["can_impersonate"]; !ok || value != true {
		t.Fatalf("expected can_impersonate=true in response body, got: %#v", body)
	}

	if _, exists := body["canimpersonate"]; exists {
		t.Fatalf("did not expect legacy key canimpersonate in response body")
	}
}

func TestNewTenantResponseDTOMapsCanImpersonate(t *testing.T) {
	tenantID := uuid.New()

	tenantModel := Tenant{
		Id:             tenantID,
		Name:           "Acme",
		CanImpersonate: true,
	}

	response := NewTenantResponseDTO(tenantModel)

	if response.TenantId != tenantID.String() {
		t.Fatalf("expected tenant_id %q, got %q", tenantID.String(), response.TenantId)
	}

	if response.TenantName != "Acme" {
		t.Fatalf("expected tenant_name Acme, got %q", response.TenantName)
	}

	if !response.CanImpersonate {
		t.Fatalf("expected can_impersonate=true")
	}
}
