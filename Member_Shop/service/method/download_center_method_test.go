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

func TestNormalizeDownloadFileFormatUsesTemplateDefault(t *testing.T) {
	format, err := NormalizeDownloadFileFormat("csv", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if format != "csv" {
		t.Fatalf("format = %s, want csv", format)
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
