package method

import (
	"Member_shop/models"
	"strings"
	"testing"
)

func TestValidateDownloadTaskRequestRejectsUnknownFilter(t *testing.T) {
	template := testDownloadTemplate()
	err := ValidateDownloadTaskRequest(template, map[string]any{"status": "paid", "unsafe": "x"}, "xlsx")
	if err == nil || !strings.Contains(err.Error(), `filter "unsafe" is not allowed`) {
		t.Fatalf("expected unknown filter error, got %v", err)
	}
}

func TestValidateDownloadTaskRequestRejectsDisabledTemplate(t *testing.T) {
	template := testDownloadTemplate()
	template.Status = DownloadTemplateStatusDisabled
	err := ValidateDownloadTaskRequest(template, map[string]any{"status": "paid"}, "xlsx")
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled template error, got %v", err)
	}
}

func TestValidateDownloadTaskRequestRejectsNonSelectSQL(t *testing.T) {
	template := testDownloadTemplate()
	template.SQLContent = "DELETE FROM order_data"
	err := ValidateDownloadTaskRequest(template, map[string]any{"status": "paid"}, "xlsx")
	if err == nil || !strings.Contains(err.Error(), "SELECT") {
		t.Fatalf("expected select-only error, got %v", err)
	}
}

func TestNormalizeDownloadFileFormatRejectsCSVUntilImplemented(t *testing.T) {
	_, err := NormalizeDownloadFileFormat("csv", "")
	if err == nil || !strings.Contains(err.Error(), "unsupported file_format") {
		t.Fatalf("expected unsupported csv error, got %v", err)
	}
}

func TestNormalizeDownloadTaskPagination(t *testing.T) {
	page, pageSize := NormalizeDownloadTaskPagination(0, 0)
	if page != 1 || pageSize != defaultDownloadTaskPageSize {
		t.Fatalf("page/pageSize = %d/%d, want 1/%d", page, pageSize, defaultDownloadTaskPageSize)
	}

	page, pageSize = NormalizeDownloadTaskPagination(3, 1000)
	if page != 3 || pageSize != maxDownloadTaskPageSize {
		t.Fatalf("page/pageSize = %d/%d, want 3/%d", page, pageSize, maxDownloadTaskPageSize)
	}
}

func TestParseDownloadFilterRulesRejectsDuplicateField(t *testing.T) {
	_, err := ParseDownloadFilterRules(`[
		{"field":"status","operator":"=","db_column":"status"},
		{"field":"status","operator":"=","db_column":"pay_status"}
	]`)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate field error, got %v", err)
	}
}

func testDownloadTemplate() *models.DownloadTemplate {
	return &models.DownloadTemplate{
		TemplateCode:  "order_export",
		TemplateName:  "订单导出",
		BusinessType:  DownloadBusinessOrder,
		SQLContent:    "SELECT order_id, status FROM order_data",
		ModelFields:   `[{"field":"order_id","model":"Order","model_field":"OrderID"}]`,
		ExportHeaders: `[{"field":"order_id","header":"订单号"}]`,
		AllowedFilters: `[
			{"field":"begin_time","operator":">=","db_column":"order_time","type":"datetime"},
			{"field":"end_time","operator":"<=","db_column":"order_time","type":"datetime"},
			{"field":"status","operator":"=","db_column":"status","type":"string"}
		]`,
		FileFormat: "xlsx",
		Status:     DownloadTemplateStatusEnabled,
	}
}
