package historical_data_test

import (
	"testing"

	"backend/internal/historical_data"
)

func TestHistoricalDataFilter_Normalize(t *testing.T) {
	t.Run("uses default limit when missing", func(t *testing.T) {
		filter := historical_data.HistoricalDataFilter{}
		got := filter.Normalize()
		if got.Limit != historical_data.DefaultHistoricalDataLimit {
			t.Fatalf("expected default limit %d, got %d", historical_data.DefaultHistoricalDataLimit, got.Limit)
		}
	})

	t.Run("preserves explicit limit", func(t *testing.T) {
		filter := historical_data.HistoricalDataFilter{Limit: 42}
		got := filter.Normalize()
		if got.Limit != 42 {
			t.Fatalf("expected limit 42, got %d", got.Limit)
		}
	})
}
