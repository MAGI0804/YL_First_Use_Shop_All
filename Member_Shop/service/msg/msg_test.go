package msg

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSuccessResponseRemovesJSONNulls(t *testing.T) {
	data := map[string]any{
		"items": nil,
		"detail": map[string]any{
			"images": nil,
			"remark": nil,
		},
		"list": []any{
			nil,
			map[string]any{"info": nil},
		},
	}

	response := SuccessResponse("ok", &data)
	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(body), "null") {
		t.Fatalf("response still contains null: %s", string(body))
	}
	if !strings.Contains(string(body), `"items":[]`) {
		t.Fatalf("items should be an empty array, got: %s", string(body))
	}
	if !strings.Contains(string(body), `"remark":""`) {
		t.Fatalf("unknown nil field should be an empty string, got: %s", string(body))
	}
}
