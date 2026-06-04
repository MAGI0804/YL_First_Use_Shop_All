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

func TestNormalizeJushuitanAfterSaleStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
	}{
		{name: "approved english", status: "APPROVED", want: ReturnOrderStatusApproved},
		{name: "approved chinese", status: "审核通过", want: ReturnOrderStatusApproved},
		{name: "received chinese", status: "退货入库", want: ReturnOrderStatusReceived},
		{name: "completed chinese", status: "退款成功", want: ReturnOrderStatusCompleted},
		{name: "rejected chinese", status: "审核拒绝", want: ReturnOrderStatusRejected},
		{name: "unknown falls back pending", status: "WAIT_SELLER_CONFIRM", want: ReturnOrderStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeJushuitanAfterSaleStatus(tt.status)
			if got != tt.want {
				t.Fatalf("status = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseAfterSaleProductItems(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		items := parseAfterSaleProductItems(`[{"commodity_id":"SKU001","qty":2}]`)
		if len(items) != 1 || items[0]["commodity_id"] != "SKU001" {
			t.Fatalf("unexpected items: %#v", items)
		}
	})

	t.Run("single object", func(t *testing.T) {
		items := parseAfterSaleProductItems(`{"commodity_id":"SKU002","qty":1}`)
		if len(items) != 1 || items[0]["commodity_id"] != "SKU002" {
			t.Fatalf("unexpected items: %#v", items)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		items := parseAfterSaleProductItems(`bad-json`)
		if len(items) != 0 {
			t.Fatalf("expected no items, got %#v", items)
		}
	})
}

func TestExtractJushuitanAfterSaleUpdates(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"datas": []interface{}{
				map[string]interface{}{
					"outer_as_id": "T001",
					"as_id":       "1001",
					"so_id":       "Y001",
					"status":      "退货入库",
				},
				map[string]interface{}{
					"outer_as_id":    "T002",
					"so_id":          "Y002",
					"received_date":  "2026-06-04 10:00:00",
					"unrelated_text": "kept",
				},
			},
		},
	}

	updates := extractJushuitanAfterSaleUpdates(payload)
	if len(updates) != 2 {
		t.Fatalf("len(updates) = %d, want 2", len(updates))
	}
	if updates[0].ReturnID != "T001" || updates[0].JushuitanAfterSaleID != "1001" || updates[0].Status != "退货入库" {
		t.Fatalf("unexpected first update: %#v", updates[0])
	}
	if updates[1].ReturnID != "T002" || updates[1].Status != ReturnOrderStatusReceived {
		t.Fatalf("unexpected second update: %#v", updates[1])
	}
}
