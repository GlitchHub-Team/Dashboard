package tenant_test

import (
	"errors"
	"testing"

	"backend/internal/identity"
	"backend/internal/tenant"
	"backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestCreateTenant(t *testing.T) {
	targetTenantID := uuid.New()
	targetTenantName := "Stefano"
	targetCanImpersonate := true

	superAdminRequester := identity.Requester{
		RequesterRole: identity.ROLE_SUPER_ADMIN,
	}

	targetCreateTenant := tenant.Tenant{
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
	}

	expectedTenant := tenant.Tenant{
		Id:             targetTenantID,
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
	}

	type mockSetupFunc func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call

	type testCase struct {
		name           string
		input          tenant.CreateTenantCommand
		setupSteps     []mockSetupFunc
		expectedTenant tenant.Tenant
		expectedError  error
	}

	step1CreateTenantOK := func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call {
		return createTenantPort.EXPECT().
			CreateTenant(targetCreateTenant).
			Return(expectedTenant, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1CreateTenantError := func(
		createTenantPort *mocks.MockCreateTenantPort,
	) *gomock.Call {
		return createTenantPort.EXPECT().
			CreateTenant(targetCreateTenant).
			Return(tenant.Tenant{}, errMockStep1).
			Times(1)
	}

	baseInput := tenant.CreateTenantCommand{
		Name:           targetTenantName,
		CanImpersonate: targetCanImpersonate,
		Requester:      superAdminRequester,
	}

	cases := []testCase{
		{
			name:  "Success: tenant created successfully",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1CreateTenantOK,
			},
			expectedTenant: expectedTenant,
			expectedError:  nil,
		},
		{
			name:  "Fail (step 1): unexpected error from port",
			input: baseInput,
			setupSteps: []mockSetupFunc{
				step1CreateTenantError,
			},
			expectedTenant: tenant.Tenant{},
			expectedError:  errMockStep1,
		},
		{
			name: "Fail: requester is not superadmin",
			input: tenant.CreateTenantCommand{
				Name:           targetTenantName,
				CanImpersonate: targetCanImpersonate,
				Requester:      identity.Requester{},
			},
			setupSteps:     []mockSetupFunc{},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrUnauthorized,
		},
		{
			name: "Fail: canImpersonate is false",
			input: tenant.CreateTenantCommand{
				Name:           targetTenantName,
				CanImpersonate: false,
				Requester:      superAdminRequester,
			},
			setupSteps:     []mockSetupFunc{},
			expectedTenant: tenant.Tenant{},
			expectedError:  tenant.ErrImpersonationFailded,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockCreateTenantPort := mocks.NewMockCreateTenantPort(mockController)

			var expectedCalls []any
			for _, step := range tc.setupSteps {
				call := step(mockCreateTenantPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			createUseCase, _, _, _, _ := tenant.NewCreateTenantService(
				mockCreateTenantPort,
				nil,
				nil,
				nil,
				nil,
			)

			createdTenant, err := createUseCase.CreateTenant(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if createdTenant != tc.expectedTenant {
				t.Errorf("expected tenant %v, got %v", tc.expectedTenant, createdTenant)
			}
		})
	}
}

