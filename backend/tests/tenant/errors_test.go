package tenant_test

import "fmt"

func newMockError(stepNumber int) error {
	return fmt.Errorf("unexpected error in step %d", stepNumber)
}
