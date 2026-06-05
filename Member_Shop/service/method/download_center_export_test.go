package method

import (
	"Member_shop/models"
	"strings"
	"testing"
)

func TestBuildDownloadSQLUsesWhitelistedFiltersAndParameters(t *testing.T) {
	template := testDownloadTemplate()
	template.DefaultOrderBy = "order_time DESC"

	query, args, err := BuildDownloadSQL(template, map[string]any{
		"status":     "paid",
		"begin_time": "2026-06-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(query, "WHERE") || !strings.Contains(query, "status = ?") || !strings.Contains(query, "order_time >= ?") {
		t.Fatalf("query does not contain expected parameterized filters: %s", query)
	}
	if strings.Contains(query, "paid") || strings.Contains(query, "2026-06-01") {
		t.Fatalf("query should not inline user values: %s", query)
	}
	if len(args) != 3 {
		t.Fatalf("args len = %d, want 3 including limit", len(args))
	}
}

func TestBuildDownloadSQLRejectsUnsafeFilterColumn(t *testing.T) {
	template := testDownloadTemplate()
	template.AllowedFilters = `[{"field":"status","operator":"=","db_column":"status;DROP","type":"string"}]`
	_, _, err := BuildDownloadSQL(template, map[string]any{"status": "paid"})
	if err == nil || !strings.Contains(err.Error(), "unsafe filter column") {
		t.Fatalf("expected unsafe column error, got %v", err)
	}
}

func TestBuildDownloadSQLRejectsUnsafeOrderBy(t *testing.T) {
	template := testDownloadTemplate()
	template.DefaultOrderBy = "order_time DESC; DROP TABLE order_data"
	_, _, err := BuildDownloadSQL(template, map[string]any{"status": "paid"})
	if err == nil || !strings.Contains(err.Error(), "unsafe default_order_by") {
		t.Fatalf("expected unsafe order by error, got %v", err)
	}
}

func TestParseDownloadExportHeadersRequiresFields(t *testing.T) {
	_, err := ParseDownloadExportHeaders(`[{"field":"","header":"订单号"}]`)
	if err == nil || !strings.Contains(err.Error(), "empty field") {
		t.Fatalf("expected empty field error, got %v", err)
	}
}

func TestNormalizeDownloadLimitClampsToDefaultMax(t *testing.T) {
	if got := normalizeDownloadLimit(float64(999999)); got != defaultDownloadMaxRows {
		t.Fatalf("limit = %d, want %d", got, defaultDownloadMaxRows)
	}
	if got := normalizeDownloadLimit("25"); got != 25 {
		t.Fatalf("limit = %d, want 25", got)
	}
}

func TestSafeDownloadOrderByAllowsMultipleColumns(t *testing.T) {
	if !isSafeDownloadOrderBy("order_time DESC, order_id ASC") {
		t.Fatalf("expected multi-column order by to be safe")
	}
	if isSafeDownloadOrderBy("order_time DESC -- comment") {
		t.Fatalf("expected comment order by to be unsafe")
	}
}

func TestBuildDownloadSQLRejectsCSVUntilImplemented(t *testing.T) {
	template := &models.DownloadTemplate{
		TemplateCode:   "csv_export",
		TemplateName:   "CSV",
		BusinessType:   DownloadBusinessOrder,
		SQLContent:     "SELECT order_id FROM order_data",
		AllowedFilters: `[]`,
		ExportHeaders:  `[{"field":"order_id","header":"订单号"}]`,
		FileFormat:     "csv",
		Status:         DownloadTemplateStatusEnabled,
		DefaultOrderBy: "order_id ASC",
		ModelFields:    `[]`,
	}
	_, _, err := BuildDownloadSQL(template, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "unsupported file_format") {
		t.Fatalf("expected unsupported csv error, got %v", err)
	}
}
