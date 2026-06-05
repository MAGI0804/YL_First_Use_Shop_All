package method

import "testing"

func TestValidatePaymentAdjustment(t *testing.T) {
	tests := []struct {
		name           string
		orderAmount    float64
		finalPayAmount float64
		discountReason string
		wantDiscount   float64
		wantErr        bool
	}{
		{
			name:           "same amount has no discount",
			orderAmount:    299,
			finalPayAmount: 299,
			wantDiscount:   0,
		},
		{
			name:           "discount requires reason and returns discount amount",
			orderAmount:    299,
			finalPayAmount: 269,
			discountReason: "returning customer",
			wantDiscount:   30,
		},
		{
			name:           "rejects negative final amount",
			orderAmount:    299,
			finalPayAmount: -1,
			wantErr:        true,
		},
		{
			name:           "rejects final amount greater than original amount",
			orderAmount:    299,
			finalPayAmount: 300,
			wantErr:        true,
		},
		{
			name:           "rejects discount without reason",
			orderAmount:    299,
			finalPayAmount: 269,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount, err := ValidatePaymentAdjustment(tt.orderAmount, tt.finalPayAmount, tt.discountReason)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if discount != tt.wantDiscount {
				t.Fatalf("discount = %v, want %v", discount, tt.wantDiscount)
			}
		})
	}
}

func TestValidateOrderReadyToPay(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{
			name:   "delivered order can be paid",
			status: "delivered",
		},
		{
			name:    "pending order cannot be paid",
			status:  "pending",
			wantErr: true,
		},
		{
			name:    "shipped order cannot be paid before delivery",
			status:  "shipped",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderReadyToPay(tt.status)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNormalizeOrderProductListSupportsBackendItems(t *testing.T) {
	got := normalizeOrderProductList(`[{"commodity_id":"SKU001","qty":2},"SKU002",{"sku_id":"SKU003"}]`)
	want := []string{"SKU001", "SKU002", "SKU003"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d, got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("item %d = %q, want %q", i, got[i], want[i])
		}
	}
}
