package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProcessTimeFieldsRemovesJSONNulls(t *testing.T) {
	input := map[string]interface{}{
		"data":       nil,
		"items":      nil,
		"remark":     nil,
		"created_at": "2026-05-18T12:34:56+08:00",
		"nested": map[string]interface{}{
			"images": nil,
			"title":  nil,
		},
		"list": []interface{}{
			nil,
			map[string]interface{}{"detail": nil},
		},
	}

	processed := processTimeFields(input)
	body, err := json.Marshal(processed)
	if err != nil {
		t.Fatalf("marshal processed response: %v", err)
	}
	if strings.Contains(string(body), "null") {
		t.Fatalf("processed response still contains null: %s", string(body))
	}

	result := processed.(map[string]interface{})
	if _, ok := result["data"].(map[string]interface{}); !ok {
		t.Fatalf("data should become an empty object, got %T", result["data"])
	}
	if items, ok := result["items"].([]interface{}); !ok || len(items) != 0 {
		t.Fatalf("items should become an empty array, got %#v", result["items"])
	}
	if result["remark"] != "" {
		t.Fatalf("unknown null field should become an empty string, got %#v", result["remark"])
	}
	if result["created_at"] != "2026-05-18 12:34:56" {
		t.Fatalf("time formatting changed unexpectedly, got %#v", result["created_at"])
	}
}

func TestProcessTimeFieldsHandlesTypedNilValues(t *testing.T) {
	type response struct {
		Data   *struct{}         `json:"data"`
		Items  []string          `json:"items"`
		Meta   map[string]string `json:"meta"`
		Remark interface{}       `json:"remark"`
	}

	processed := processTimeFields(response{})
	body, err := json.Marshal(processed)
	if err != nil {
		t.Fatalf("marshal processed response: %v", err)
	}
	if strings.Contains(string(body), "null") {
		t.Fatalf("processed typed response still contains null: %s", string(body))
	}
}

func TestFormatTimeMiddlewareRemovesJSONNulls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(FormatTimeMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"items":    nil,
				"page":     1,
				"pageSize": 10,
				"total":    0,
			},
			"msg": "查询成功",
		})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(recorder, request)

	body := recorder.Body.String()
	if strings.Contains(body, "null") {
		t.Fatalf("middleware response still contains null: %s", body)
	}
	if !strings.Contains(body, `"items":[]`) {
		t.Fatalf("middleware response should contain empty items array, got: %s", body)
	}
}
