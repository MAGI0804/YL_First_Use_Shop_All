package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	DownloadTemplateStatusEnabled  = "enabled"
	DownloadTemplateStatusDisabled = "disabled"

	DownloadTaskStatusPending = "pending"
	DownloadTaskStatusRunning = "running"
	DownloadTaskStatusSuccess = "success"
	DownloadTaskStatusFailed  = "failed"
	DownloadTaskStatusExpired = "expired"

	DownloadBusinessOrder     = "order"
	DownloadBusinessProduct   = "product"
	DownloadBusinessReport    = "report"
	DownloadBusinessInventory = "inventory"
	DownloadBusinessAfterSale = "after_sale"
)

var allowedDownloadBusinessTypes = map[string]bool{
	DownloadBusinessOrder:     true,
	DownloadBusinessProduct:   true,
	DownloadBusinessReport:    true,
	DownloadBusinessInventory: true,
	DownloadBusinessAfterSale: true,
}

var allowedDownloadFileFormats = map[string]bool{
	"xlsx": true,
	"csv":  true,
}

type DownloadFilterRule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	DBColumn string `json:"db_column"`
	Type     string `json:"type"`
}

type CreateDownloadTaskInput struct {
	TemplateCode string
	Filters      map[string]any
	FileFormat   string
	RequestedBy  int
}

func CreateDownloadTask(input CreateDownloadTaskInput) (*models.DownloadTask, error) {
	template, err := FindEnabledDownloadTemplate(input.TemplateCode)
	if err != nil {
		return nil, err
	}
	if err := ValidateDownloadTaskRequest(template, input.Filters, input.FileFormat); err != nil {
		return nil, err
	}

	filtersJSON, err := json.Marshal(normalizeDownloadFilters(input.Filters))
	if err != nil {
		return nil, fmt.Errorf("marshal filters: %w", err)
	}

	task := &models.DownloadTask{
		TaskID:       GenerateDownloadTaskID(time.Now()),
		TemplateCode: template.TemplateCode,
		BusinessType: template.BusinessType,
		TaskName:     template.TemplateName,
		Filters:      string(filtersJSON),
		Status:       DownloadTaskStatusPending,
		Progress:     0,
		RequestedBy:  input.RequestedBy,
		FileName:     buildInitialDownloadFileName(template, input.FileFormat),
	}
	if err := db.DB.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func FindEnabledDownloadTemplate(templateCode string) (*models.DownloadTemplate, error) {
	templateCode = strings.TrimSpace(templateCode)
	if templateCode == "" {
		return nil, fmt.Errorf("template_code is required")
	}

	var template models.DownloadTemplate
	if err := db.DB.Where("template_code = ?", templateCode).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("download template not found")
		}
		return nil, err
	}
	if strings.TrimSpace(template.Status) != DownloadTemplateStatusEnabled {
		return nil, fmt.Errorf("download template disabled")
	}
	return &template, nil
}

func ValidateDownloadTaskRequest(template *models.DownloadTemplate, filters map[string]any, fileFormat string) error {
	if template == nil {
		return fmt.Errorf("download template is required")
	}
	if strings.TrimSpace(template.TemplateCode) == "" {
		return fmt.Errorf("template_code is required")
	}
	if !allowedDownloadBusinessTypes[strings.TrimSpace(template.BusinessType)] {
		return fmt.Errorf("unsupported business_type %q", template.BusinessType)
	}
	if strings.TrimSpace(template.Status) != DownloadTemplateStatusEnabled {
		return fmt.Errorf("download template disabled")
	}
	if strings.TrimSpace(template.SQLContent) == "" {
		return fmt.Errorf("sql_content is required")
	}
	if !isSelectOnlySQL(template.SQLContent) {
		return fmt.Errorf("sql_content must be a SELECT query")
	}
	if _, err := ParseDownloadFilterRules(template.AllowedFilters); err != nil {
		return err
	}
	if err := validateDownloadFilters(template.AllowedFilters, filters); err != nil {
		return err
	}
	if _, err := NormalizeDownloadFileFormat(template.FileFormat, fileFormat); err != nil {
		return err
	}
	return nil
}

func ParseDownloadFilterRules(raw string) (map[string]DownloadFilterRule, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]DownloadFilterRule{}, nil
	}

	var rules []DownloadFilterRule
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return nil, fmt.Errorf("allowed_filters invalid: %w", err)
	}

	result := make(map[string]DownloadFilterRule, len(rules))
	for _, rule := range rules {
		field := strings.TrimSpace(rule.Field)
		if field == "" {
			return nil, fmt.Errorf("allowed_filters contains empty field")
		}
		if result[field].Field != "" {
			return nil, fmt.Errorf("allowed_filters contains duplicate field %q", field)
		}
		rule.Field = field
		result[field] = rule
	}
	return result, nil
}

func NormalizeDownloadFileFormat(templateDefault, requested string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(requested))
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(templateDefault))
	}
	if format == "" {
		format = "xlsx"
	}
	if !allowedDownloadFileFormats[format] {
		return "", fmt.Errorf("unsupported file_format %q", format)
	}
	return format, nil
}

func GenerateDownloadTaskID(now time.Time) string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("DL%s00000000", now.Format("20060102150405"))
	}
	return fmt.Sprintf("DL%s%s", now.Format("20060102150405"), strings.ToUpper(hex.EncodeToString(buf)))
}

func validateDownloadFilters(allowedFiltersRaw string, filters map[string]any) error {
	allowed, err := ParseDownloadFilterRules(allowedFiltersRaw)
	if err != nil {
		return err
	}
	for key := range filters {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("filter field cannot be empty")
		}
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("filter %q is not allowed", key)
		}
	}
	return nil
}

func normalizeDownloadFilters(filters map[string]any) map[string]any {
	if filters == nil {
		return map[string]any{}
	}
	return filters
}

func isSelectOnlySQL(sqlContent string) bool {
	normalized := strings.TrimSpace(strings.ToLower(sqlContent))
	if !strings.HasPrefix(normalized, "select ") {
		return false
	}
	blocked := []string{";", " insert ", " update ", " delete ", " drop ", " alter ", " truncate "}
	padded := " " + normalized + " "
	for _, token := range blocked {
		if strings.Contains(padded, token) {
			return false
		}
	}
	return true
}

func buildInitialDownloadFileName(template *models.DownloadTemplate, requestedFormat string) string {
	format, err := NormalizeDownloadFileFormat(template.FileFormat, requestedFormat)
	if err != nil {
		format = "xlsx"
	}
	name := strings.TrimSpace(template.TemplateName)
	if name == "" {
		name = strings.TrimSpace(template.TemplateCode)
	}
	if name == "" {
		name = "下载任务"
	}
	return fmt.Sprintf("%s-%s.%s", name, time.Now().Format("2006-01-02"), format)
}
