package helper

import "go.uber.org/mock/gomock"

type OrderedMockStep func() *gomock.Call

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
