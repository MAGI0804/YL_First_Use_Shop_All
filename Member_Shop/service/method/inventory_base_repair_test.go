package method

import "testing"

func TestInventoryChangeTypeNamesMatchPlan(t *testing.T) {
	tests := map[string]string{
		"order deduct":   InventoryChangeOrderDeduct,
		"order cancel":   InventoryChangeOrderCancelRestore,
		"return restore": InventoryChangeReturnRestore,
		"manual adjust":  InventoryChangeManualAdjust,
		"jst sync":       InventoryChangeSyncJushuitan,
		"transfer":       InventoryChangeStockTransfer,
		"stock check":    InventoryChangeStockCheck,
	}

	want := map[string]string{
		"order deduct":   "order_create_deduct",
		"order cancel":   "order_cancel_restore",
		"return restore": "return_completed_restore",
		"manual adjust":  "manual_adjust",
		"jst sync":       "jushuitan_sync",
		"transfer":       "stock_transfer",
		"stock check":    "stock_check",
	}

	for name, got := range tests {
		if got != want[name] {
			t.Fatalf("%s change type = %q, want %q", name, got, want[name])
		}
	}
}

func TestCalculateJushuitanAvailableQty(t *testing.T) {
	tests := []struct {
		name       string
		qty        int
		orderLock  int
		virtualQty int
		want       int
	}{
		{name: "main stock minus order lock plus virtual", qty: 10, orderLock: 3, virtualQty: 2, want: 9},
		{name: "floors at zero", qty: 1, orderLock: 5, virtualQty: 0, want: 0},
		{name: "virtual can increase available", qty: 0, orderLock: 0, virtualQty: 4, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateJushuitanAvailableQty(tt.qty, tt.orderLock, tt.virtualQty)
			if got != tt.want {
				t.Fatalf("available = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateInventoryTransferInput(t *testing.T) {
	tests := []struct {
		name    string
		input   InventoryTransferInput
		wantErr bool
	}{
		{
			name: "valid transfer trims fields",
			input: InventoryTransferInput{
				CommodityID:         " SKU001 ",
				Qty:                 3,
				SourceWarehouseCode: " WH-A ",
				TargetWarehouseCode: " WH-B ",
			},
		},
		{
			name: "rejects empty commodity",
			input: InventoryTransferInput{
				Qty:                 3,
				SourceWarehouseCode: "WH-A",
				TargetWarehouseCode: "WH-B",
			},
			wantErr: true,
		},
		{
			name: "rejects non-positive qty",
			input: InventoryTransferInput{
				CommodityID:         "SKU001",
				Qty:                 0,
				SourceWarehouseCode: "WH-A",
				TargetWarehouseCode: "WH-B",
			},
			wantErr: true,
		},
		{
			name: "rejects same warehouse",
			input: InventoryTransferInput{
				CommodityID:         "SKU001",
				Qty:                 3,
				SourceWarehouseCode: "WH-A",
				TargetWarehouseCode: "WH-A",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInventoryTransferInput(&tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.input.CommodityID != "SKU001" {
				t.Fatalf("commodity_id was not normalized: %q", tt.input.CommodityID)
			}
		})
	}
}

func TestValidateInventoryStockCheckInput(t *testing.T) {
	tests := []struct {
		name    string
		input   InventoryStockCheckInput
		wantErr bool
	}{
		{
			name: "valid stock check trims fields",
			input: InventoryStockCheckInput{
				CommodityID:   " SKU001 ",
				ActualQty:     10,
				WarehouseCode: " WH-A ",
			},
		},
		{
			name: "rejects empty commodity",
			input: InventoryStockCheckInput{
				ActualQty: 10,
			},
			wantErr: true,
		},
		{
			name: "rejects negative actual qty",
			input: InventoryStockCheckInput{
				CommodityID: "SKU001",
				ActualQty:   -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInventoryStockCheckInput(&tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.input.CommodityID != "SKU001" || tt.input.WarehouseCode != "WH-A" {
				t.Fatalf("fields were not normalized: commodity=%q warehouse=%q", tt.input.CommodityID, tt.input.WarehouseCode)
			}
		})
	}
}

func TestOpenInventoryIdempotencyKey(t *testing.T) {
	input := ChangeInventoryInput{
		CommodityID:       "SKU001",
		ChangeType:        InventoryChangeOrderDeduct,
		RelatedSubOrderID: "SUB001",
	}

	firstKey, firstStable := openInventoryIdempotencyKey(input)
	secondKey, secondStable := openInventoryIdempotencyKey(input)
	if !firstStable || !secondStable {
		t.Fatalf("sub order inventory changes should use stable idempotency keys")
	}
	if firstKey != secondKey {
		t.Fatalf("stable idempotency keys differ: %q != %q", firstKey, secondKey)
	}
	if openInventoryMovementNo(firstKey) != openInventoryMovementNo(secondKey) {
		t.Fatalf("movement no should be deterministic for stable idempotency keys")
	}
}

func TestOpenInventoryManualIdempotencyKeyIsNotStable(t *testing.T) {
	key, stable := openInventoryIdempotencyKey(ChangeInventoryInput{
		CommodityID: "SKU001",
		ChangeType:  InventoryChangeManualAdjust,
	})
	if stable {
		t.Fatalf("manual inventory change without business id should not be stable")
	}
	if key == "" {
		t.Fatalf("manual inventory change still needs an idempotency key")
	}
}

func TestNonNegativeOpenInventoryQty(t *testing.T) {
	if got := nonNegativeOpenInventoryQty(-3); got != 0 {
		t.Fatalf("negative qty normalized to %d, want 0", got)
	}
	if got := nonNegativeOpenInventoryQty(5); got != 5 {
		t.Fatalf("positive qty normalized to %d, want 5", got)
	}
}
