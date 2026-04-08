package historical_data_test

import "fmt"

func newMockError(stepNumber int) error {
	return fmt.Errorf("unexpected error in step %d", stepNumber)
}
