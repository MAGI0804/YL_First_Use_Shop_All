package method

import (
	"Member_shop/models"
	"testing"
)

func TestBuildBackendOperatorSnapshot(t *testing.T) {
	user := &models.BackendUser{
		ID:         9,
		OperatorNo: "OP202606050001",
		Mobile:     "13800000000",
		Role:       BackendRoleAdmin,
	}

	got := BuildBackendOperatorSnapshot(user)
	if got.ID != user.ID || got.OperatorNo != user.OperatorNo || got.Mobile != user.Mobile || got.Role != user.Role {
		t.Fatalf("snapshot = %+v, want user fields", got)
	}
}

func TestEncodeOperationPayload(t *testing.T) {
	got, err := encodeOperationPayload(map[string]any{"quantity": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `{"quantity":2}` {
		t.Fatalf("payload = %s", got)
	}

	empty, err := encodeOperationPayload(nil)
	if err != nil {
		t.Fatalf("unexpected error for nil: %v", err)
	}
	if empty != "" {
		t.Fatalf("nil payload = %q, want empty", empty)
	}
}

func TestNormalizeBackendPage(t *testing.T) {
	page, pageSize := normalizeBackendPage(0, 500)
	if page != 1 || pageSize != 100 {
		t.Fatalf("normalizePage = (%d, %d), want (1, 100)", page, pageSize)
	}
}
