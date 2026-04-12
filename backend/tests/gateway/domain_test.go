package gateway_test

import (
	"testing"
	"time"

	"backend/internal/gateway"

	"github.com/google/uuid"
)

func TestGatewayDomain_IsZero(t *testing.T) {
	t.Run("true when gateway is zero value", func(t *testing.T) {
		var g gateway.Gateway

		if !g.IsZero() {
			t.Fatalf("expected true for zero value gateway")
		}
	})

	t.Run("false when gateway has id", func(t *testing.T) {
		g := gateway.Gateway{Id: uuid.New()}

		if g.IsZero() {
			t.Fatalf("expected false for non-zero gateway")
		}
	})
}

func TestGatewayDomain_IsCommissioned(t *testing.T) {
	t.Run("true when tenant id is set", func(t *testing.T) {
		tenantID := uuid.New()
		g := gateway.Gateway{TenantId: &tenantID}

		if !g.IsCommissioned() {
			t.Fatalf("expected gateway to be commissioned")
		}
	})

	t.Run("false when tenant id is nil", func(t *testing.T) {
		g := gateway.Gateway{}

		if g.IsCommissioned() {
			t.Fatalf("expected gateway to be not commissioned")
		}
	})
}

func TestGatewayDomain_GetId(t *testing.T) {
	expectedID := uuid.New()
	g := gateway.Gateway{Id: expectedID}

	if got := g.GetId(); got != expectedID {
		t.Fatalf("expected id %s, got %s", expectedID, got)
	}
}

func TestGatewayDomain_BelongsToTenant(t *testing.T) {
	t.Run("true when tenant matches", func(t *testing.T) {
		tenantID := uuid.New()
		g := gateway.Gateway{TenantId: &tenantID}

		if !g.BelongsToTenant(tenantID) {
			t.Fatalf("expected gateway to belong to tenant")
		}
	})

	t.Run("false when tenant does not match", func(t *testing.T) {
		tenantID := uuid.New()
		otherTenantID := uuid.New()
		g := gateway.Gateway{TenantId: &tenantID}

		if g.BelongsToTenant(otherTenantID) {
			t.Fatalf("expected gateway to not belong to tenant")
		}
	})

	t.Run("false when tenant id is nil", func(t *testing.T) {
		g := gateway.Gateway{}

		if g.BelongsToTenant(uuid.New()) {
			t.Fatalf("expected gateway with nil tenant to not belong to tenant")
		}
	})
}

func TestGatewayDomain_Constants(t *testing.T) {
	t.Run("gateway statuses", func(t *testing.T) {
		if gateway.GATEWAY_STATUS_ACTIVE != "active" {
			t.Fatalf("unexpected active status: %s", gateway.GATEWAY_STATUS_ACTIVE)
		}
		if gateway.GATEWAY_STATUS_INACTIVE != "inactive" {
			t.Fatalf("unexpected inactive status: %s", gateway.GATEWAY_STATUS_INACTIVE)
		}
		if gateway.GATEWAY_STATUS_DECOMMISSIONED != "decommissioned" {
			t.Fatalf("unexpected decommissioned status: %s", gateway.GATEWAY_STATUS_DECOMMISSIONED)
		}
	})

	t.Run("default interval limit", func(t *testing.T) {
		if gateway.DEFAULT_INTERVAL_LIMIT != 5*time.Second {
			t.Fatalf("unexpected default interval limit: %v", gateway.DEFAULT_INTERVAL_LIMIT)
		}
	})

	t.Run("command subjects", func(t *testing.T) {
		type subjectCase struct {
			name     string
			expected string
			got      string
		}

		cases := []subjectCase{
			{name: "create", expected: "commands.creategateway", got: gateway.CREATE_GATEWAY_COMMAND_SUBJECT},
			{name: "delete", expected: "commands.deletegateway", got: gateway.DELETE_GATEWAY_COMMAND_SUBJECT},
			{name: "commission", expected: "commands.commissiongateway", got: gateway.COMMISSION_GATEWAY_COMMAND_SUBJECT},
			{name: "decommission", expected: "commands.decommissiongateway", got: gateway.DECOMMISSION_GATEWAY_COMMAND_SUBJECT},
			{name: "interrupt", expected: "commands.interruptgateway", got: gateway.INTERRUPT_GATEWAY_COMMAND_SUBJECT},
			{name: "resume", expected: "commands.resumegateway", got: gateway.RESUME_GATEWAY_COMMAND_SUBJECT},
			{name: "reset", expected: "commands.resetgateway", got: gateway.RESET_GATEWAY_COMMAND_SUBJECT},
			{name: "reboot", expected: "commands.rebootgateway", got: gateway.REBOOT_GATEWAY_COMMAND_SUBJECT},
		}

		for _, tc := range cases {
			if tc.got != tc.expected {
				t.Fatalf("unexpected %s command subject: expected %s, got %s", tc.name, tc.expected, tc.got)
			}
		}
	})
}