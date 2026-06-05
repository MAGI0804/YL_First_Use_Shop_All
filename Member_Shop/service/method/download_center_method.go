package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	defaultDownloadTaskPageSize = 20
	maxDownloadTaskPageSize     = 100

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
}

var allowedDownloadTaskStatuses = map[string]bool{
	DownloadTaskStatusPending: true,
	DownloadTaskStatusRunning: true,
	DownloadTaskStatusSuccess: true,
	DownloadTaskStatusFailed:  true,
	DownloadTaskStatusExpired: true,
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

type QueryDownloadTasksInput struct {
	Page         int
	PageSize     int
	Status       string
	BusinessType string
	TemplateCode string
	RequestedBy  int
	IsAdmin      bool
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
	go func(taskID string) {
		if err := GenerateDownloadTask(taskID); err != nil {
			log.Printf("generate download task %s failed: %v", taskID, err)
		}
	}(task.TaskID)
	return task, nil
}

func QueryDownloadTasks(input QueryDownloadTasksInput) ([]models.DownloadTask, int64, int, int, error) {
	page, pageSize := NormalizeDownloadTaskPagination(input.Page, input.PageSize)
	query := db.DB.Model(&models.DownloadTask{})

	status := strings.TrimSpace(input.Status)
	if status != "" {
		if !allowedDownloadTaskStatuses[status] {
			return nil, 0, page, pageSize, fmt.Errorf("unsupported task status %q", status)
		}
		query = query.Where("status = ?", status)
	}
	businessType := strings.TrimSpace(input.BusinessType)
	if businessType != "" {
		if !allowedDownloadBusinessTypes[businessType] {
			return nil, 0, page, pageSize, fmt.Errorf("unsupported business_type %q", businessType)
		}
		query = query.Where("business_type = ?", businessType)
	}
	if strings.TrimSpace(input.TemplateCode) != "" {
		query = query.Where("template_code = ?", strings.TrimSpace(input.TemplateCode))
	}
	if !input.IsAdmin {
		query = query.Where("requested_by = ?", input.RequestedBy)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	tasks := []models.DownloadTask{}
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, page, pageSize, err
	}
	return tasks, total, page, pageSize, nil
}

func GetDownloadTask(taskID string, requestedBy int, isAdmin bool) (*models.DownloadTask, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	query := db.DB.Where("task_id = ?", taskID)
	if !isAdmin {
		query = query.Where("requested_by = ?", requestedBy)
	}

	var task models.DownloadTask
	if err := query.First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("download task not found")
		}
		return nil, err
	}
	return &task, nil
}

func RetryDownloadTask(taskID string, requestedBy int, isAdmin bool) (*models.DownloadTask, error) {
	task, err := GetDownloadTask(taskID, requestedBy, isAdmin)
	if err != nil {
		return nil, err
	}
	if task.Status != DownloadTaskStatusFailed {
		return nil, fmt.Errorf("only failed download tasks can be retried")
	}
	updates := map[string]any{
		"status":        DownloadTaskStatusPending,
		"progress":      0,
		"error_message": "",
		"started_at":    nil,
		"finished_at":   nil,
	}
	if err := db.DB.Model(task).Updates(updates).Error; err != nil {
		return nil, err
	}
	return GetDownloadTask(taskID, requestedBy, isAdmin)
}

func ListEnabledDownloadTemplates() ([]models.DownloadTemplate, error) {
	templates := []models.DownloadTemplate{}
	if err := db.DB.Where("status = ?", DownloadTemplateStatusEnabled).
		Order("business_type ASC, template_code ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func ResolveDownloadTaskFile(taskID string, requestedBy int, isAdmin bool) (string, string, error) {
	task, err := GetDownloadTask(taskID, requestedBy, isAdmin)
	if err != nil {
		return "", "", err
	}
	if task.Status != DownloadTaskStatusSuccess {
		return "", "", fmt.Errorf("download task is not ready")
	}
	fullPath, err := safeDownloadFullPath(task.FilePath)
	if err != nil {
		return "", "", err
	}
	if err := db.DB.Model(task).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
		return "", "", err
	}
	fileName := strings.TrimSpace(task.FileName)
	if fileName == "" {
		fileName = task.TaskID + ".xlsx"
	}
	return fullPath, fileName, nil
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

func NormalizeDownloadTaskPagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultDownloadTaskPageSize
	}
	if pageSize > maxDownloadTaskPageSize {
		pageSize = maxDownloadTaskPageSize
	}
	return page, pageSize
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
