package auth_integration_test

import (
	"net/http"
	"testing"

	"backend/internal/shared/identity"
	"backend/tests/helper"

	"github.com/google/uuid"
)

func TestLogoutUserIntegration(t *testing.T) {
    deps := helper.SetupIntegrationTest(t)

    // create a JWT for a fake requester
    jwt, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 123, RequesterRole: identity.ROLE_TENANT_USER, RequesterTenantId: &uuid.UUID{}})
    if err != nil {
        t.Fatalf("cannot generate jwt: %v", err)
    }

    tests := []*helper.IntegrationTestCase{
        {
            PreSetups: []helper.IntegrationTestPreSetup{},
            Name:      "Success: logout with JWT",
            Method:    http.MethodPost,
            Path:      "/api/v1/auth/logout",
            Header:    http.Header{"Authorization": {"Bearer " + jwt}},
            Body:      nil,

            WantStatusCode:   http.StatusOK,
            WantResponseBody: `"result":"ok"`,
            ResponseChecks:   nil,
            PostSetups:       []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
        },
        {
            PreSetups: []helper.IntegrationTestPreSetup{},
            Name:      "Fail: logout no JWT",
            Method:    http.MethodPost,
            Path:      "/api/v1/auth/logout",
            Header:    http.Header{},
            Body:      nil,

            WantStatusCode:   http.StatusBadRequest,
            WantResponseBody: "error",
            ResponseChecks:   nil,
            PostSetups:       []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
        },
    }

    helper.RunIntegrationTests(t, tests, deps)
}
