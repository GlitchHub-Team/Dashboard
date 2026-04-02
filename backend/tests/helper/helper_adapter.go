package helper

import (
	"testing"

	"go.uber.org/mock/gomock"
)

type AdapterMockSetupFunc[MockBundleT any] func(MockBundleT) *gomock.Call

// SetupAdapterWithOrderedSteps builds typed mock bundles and applies ordered setup steps before creating the adapter under test.
func SetupAdapterWithOrderedSteps[MockBundleT any, AdapterT any](
	t *testing.T,
	newMockBundle func(ctrl *gomock.Controller) MockBundleT,
	setupSteps []AdapterMockSetupFunc[MockBundleT],
	newAdapter func(mockBundle MockBundleT) AdapterT,
) AdapterT {
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

	return newAdapter(mockBundle)
}
