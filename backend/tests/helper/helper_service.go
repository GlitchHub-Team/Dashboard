package helper

import (
	"testing"

	"go.uber.org/mock/gomock"
)

type (
	OrderedMockStep                       func() *gomock.Call
	ServiceMockSetupFunc[MockBundleT any] func(MockBundleT) *gomock.Call
)

// SetupOrderedMockSteps applies mock expectations and enforces call order across service test steps.
func SetupOrderedMockSteps(setupSteps []OrderedMockStep) {
	var expectedCalls []any
	for _, step := range setupSteps {
		if call := step(); call != nil {
			expectedCalls = append(expectedCalls, call)
		}
	}
	if len(expectedCalls) > 0 {
		gomock.InOrder(expectedCalls...)
	}
}

// SetupServiceWithOrderedSteps builds typed mock bundles and applies ordered setup steps before creating the service under test.
func SetupServiceWithOrderedSteps[MockBundleT any, ServiceT any](
	t *testing.T,
	newMockBundle func(ctrl *gomock.Controller) MockBundleT,
	setupSteps []ServiceMockSetupFunc[MockBundleT],
	newService func(mockBundle MockBundleT) ServiceT,
) ServiceT {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockBundle := newMockBundle(ctrl)

	boundSetupSteps := make([]OrderedMockStep, 0, len(setupSteps))
	for _, step := range setupSteps {
		currentStep := step
		boundSetupSteps = append(boundSetupSteps, func() *gomock.Call {
			return currentStep(mockBundle)
		})
	}

	SetupOrderedMockSteps(boundSetupSteps)

	return newService(mockBundle)
}
