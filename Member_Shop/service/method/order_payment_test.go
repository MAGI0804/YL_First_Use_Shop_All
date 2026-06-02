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
