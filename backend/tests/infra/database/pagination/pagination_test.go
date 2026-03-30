package pagination_test

import (
	"testing"

	"backend/internal/infra/database/pagination"
)

func TestPageLimitToOffset(t *testing.T) {
	type testCase struct {
		name           string
		inputPage      int
		inputLimit     int
		expectedOffset int
		expectedError  error
	}

	cases := []testCase{
		{
			name:           "Success (page = 1)",
			inputPage:      1,
			inputLimit:     10,
			expectedOffset: 0,
			expectedError:  nil,
		},
		{
			name:           "Success (page = 2)",
			inputPage:      2,
			inputLimit:     10,
			expectedOffset: 10,
			expectedError:  nil,
		},
		{
			name:           "Error: invalid page (=0)",
			inputPage:      0,
			inputLimit:     10,
			expectedOffset: 0,
			expectedError:  pagination.ErrInvalidPage,
		},
		{
			name:           "Error: invalid page (<0)",
			inputPage:      -1,
			inputLimit:     10,
			expectedOffset: 0,
			expectedError:  pagination.ErrInvalidPage,
		},
		{
			name:           "Error: invalid limit (=0)",
			inputPage:      1,
			inputLimit:     0,
			expectedOffset: 0,
			expectedError:  pagination.ErrInvalidLimit,
		},
		{
			name:           "Error: invalid limit (<0)",
			inputPage:      1,
			inputLimit:     -1,
			expectedOffset: 0,
			expectedError:  pagination.ErrInvalidLimit,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			offset, err := pagination.PageLimitToOffset(tc.inputPage, tc.inputLimit)

			if offset != tc.expectedOffset {
				t.Errorf("want offset %v, got %v", tc.expectedOffset, offset)
			}

			if err != tc.expectedError {
				t.Errorf("want error %v, got %v", tc.expectedError, err)
			}
		})
	}
}
