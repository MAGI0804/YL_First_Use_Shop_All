package method

import "testing"

func TestCalculateAfterSaleRate(t *testing.T) {
	tests := []struct {
		name            string
		afterSaleOrders int64
		completedOrders int64
		want            float64
	}{
		{name: "no completed orders", afterSaleOrders: 3, completedOrders: 0, want: 0},
		{name: "zero after sales", afterSaleOrders: 0, completedOrders: 10, want: 0},
		{name: "rounds to four decimals", afterSaleOrders: 1, completedOrders: 3, want: 0.3333},
		{name: "whole rate", afterSaleOrders: 5, completedOrders: 5, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateAfterSaleRate(tt.afterSaleOrders, tt.completedOrders)
			if got != tt.want {
				t.Fatalf("rate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReturnOrderTimeRange(t *testing.T) {
	begin, end, err := parseReturnOrderTimeRange("2026-06-01", "2026-06-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if begin == nil || end == nil {
		t.Fatalf("expected both begin and end")
	}
	if !end.After(*begin) {
		t.Fatalf("end should be after begin")
	}

	if _, _, err := parseReturnOrderTimeRange("2026-06-03", "2026-06-01"); err == nil {
		t.Fatalf("expected begin-after-end error")
	}

	if _, _, err := parseReturnOrderTimeRange("bad-time", ""); err == nil {
		t.Fatalf("expected invalid time error")
	}
}
