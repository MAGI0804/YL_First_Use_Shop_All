package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const (
	defaultDownloadMaxRows = 5000
	downloadFileTTL        = 7 * 24 * time.Hour
)

var safeDBColumnPattern = regexp.MustCompile(`^[A-Za-z0-9_.]+$`)

type DownloadExportHeader struct {
	Field  string  `json:"field"`
	Header string  `json:"header"`
	Width  float64 `json:"width"`
	Format string  `json:"format"`
}

func GenerateDownloadTask(taskID string) error {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("task_id is required")
	}

	var task models.DownloadTask
	if err := db.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("download task not found")
		}
		return err
	}
	if task.Status != DownloadTaskStatusPending && task.Status != DownloadTaskStatusRunning {
		return fmt.Errorf("download task status %s cannot be generated", task.Status)
	}

	startedAt := time.Now()
	if err := db.DB.Model(&task).Updates(map[string]any{
		"status":     DownloadTaskStatusRunning,
		"progress":   10,
		"started_at": &startedAt,
	}).Error; err != nil {
		return err
	}

	if err := generateDownloadTaskFile(&task); err != nil {
		finishedAt := time.Now()
		_ = db.DB.Model(&task).Updates(map[string]any{
			"status":        DownloadTaskStatusFailed,
			"progress":      100,
			"error_message": sanitizeDownloadError(err),
			"finished_at":   &finishedAt,
		}).Error
		return err
	}
	return nil
}

func BuildDownloadSQL(template *models.DownloadTemplate, filters map[string]any) (string, []any, error) {
	if err := ValidateDownloadTaskRequest(template, filters, ""); err != nil {
		return "", nil, err
	}

	rules, err := ParseDownloadFilterRules(template.AllowedFilters)
	if err != nil {
		return "", nil, err
	}

	conditions := make([]string, 0)
	args := make([]any, 0)
	limit := defaultDownloadMaxRows
	for key, value := range normalizeDownloadFilters(filters) {
		if isEmptyDownloadFilterValue(value) {
			continue
		}
		rule, ok := rules[key]
		if !ok {
			return "", nil, fmt.Errorf("filter %q is not allowed", key)
		}
		operator := strings.ToUpper(strings.TrimSpace(rule.Operator))
		if operator == "LIMIT" {
			limit = normalizeDownloadLimit(value)
			continue
		}
		if !isSafeDownloadDBColumn(rule.DBColumn) {
			return "", nil, fmt.Errorf("unsafe filter column %q", rule.DBColumn)
		}
		if !isAllowedDownloadOperator(operator) {
			return "", nil, fmt.Errorf("unsupported filter operator %q", rule.Operator)
		}
		conditions = append(conditions, fmt.Sprintf("%s %s ?", rule.DBColumn, operator))
		if operator == "LIKE" {
			args = append(args, "%"+fmt.Sprintf("%v", value)+"%")
		} else {
			args = append(args, value)
		}
	}

	query := strings.TrimSpace(template.SQLContent)
	if len(conditions) > 0 {
		if strings.Contains(strings.ToLower(query), " where ") {
			query += " AND " + strings.Join(conditions, " AND ")
		} else {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
	}
	orderBy := strings.TrimSpace(template.DefaultOrderBy)
	if orderBy != "" {
		if !isSafeDownloadOrderBy(orderBy) {
			return "", nil, fmt.Errorf("unsafe default_order_by")
		}
		query += " ORDER BY " + orderBy
	}
	query += " LIMIT ?"
	args = append(args, limit)
	return query, args, nil
}

func ParseDownloadExportHeaders(raw string) ([]DownloadExportHeader, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("export_headers is required")
	}
	var headers []DownloadExportHeader
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		return nil, fmt.Errorf("export_headers invalid: %w", err)
	}
	for index := range headers {
		headers[index].Field = strings.TrimSpace(headers[index].Field)
		headers[index].Header = strings.TrimSpace(headers[index].Header)
		if headers[index].Field == "" || headers[index].Header == "" {
			return nil, fmt.Errorf("export_headers contains empty field or header")
		}
	}
	return headers, nil
}

func generateDownloadTaskFile(task *models.DownloadTask) error {
	template, err := FindEnabledDownloadTemplate(task.TemplateCode)
	if err != nil {
		return err
	}

	var filters map[string]any
	if strings.TrimSpace(task.Filters) == "" {
		filters = map[string]any{}
	} else if err := json.Unmarshal([]byte(task.Filters), &filters); err != nil {
		return fmt.Errorf("task filters invalid: %w", err)
	}

	query, args, err := BuildDownloadSQL(template, filters)
	if err != nil {
		return err
	}
	headers, err := ParseDownloadExportHeaders(template.ExportHeaders)
	if err != nil {
		return err
	}

	rows, err := db.DB.Raw(query, args...).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	relativePath := filepath.Join("download_center", time.Now().Format("2006-01-02"), task.TaskID+".xlsx")
	fullPath := utils.MediaPath(relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	rowCount, err := writeDownloadRowsToXLSX(headers, rows, fullPath)
	if err != nil {
		return err
	}
	stat, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	finishedAt := time.Now()
	expiresAt := finishedAt.Add(downloadFileTTL)
	return db.DB.Model(task).Updates(map[string]any{
		"status":        DownloadTaskStatusSuccess,
		"progress":      100,
		"row_count":     rowCount,
		"file_path":     filepath.ToSlash(relativePath),
		"file_size":     stat.Size(),
		"error_message": "",
		"finished_at":   &finishedAt,
		"expires_at":    &expiresAt,
	}).Error
}

func writeDownloadRowsToXLSX(headers []DownloadExportHeader, rows *sql.Rows, fullPath string) (int64, error) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "Sheet1"
	stream, err := file.NewStreamWriter(sheet)
	if err != nil {
		return 0, err
	}

	headerRow := make([]interface{}, 0, len(headers))
	for index, header := range headers {
		headerRow = append(headerRow, header.Header)
		if header.Width > 0 {
			columnName := excelColumnName(index + 1)
			if err := file.SetColWidth(sheet, columnName, columnName, header.Width); err != nil {
				return 0, err
			}
		}
	}
	if err := stream.SetRow("A1", headerRow); err != nil {
		return 0, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	rowNumber := 2
	var rowCount int64
	for rows.Next() {
		values := make([]any, len(columns))
		scanTargets := make([]any, len(columns))
		for index := range values {
			scanTargets[index] = &values[index]
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return rowCount, err
		}
		valueByColumn := make(map[string]any, len(columns))
		for index, column := range columns {
			valueByColumn[column] = normalizeDownloadCellValue(values[index])
		}

		dataRow := make([]interface{}, 0, len(headers))
		for _, header := range headers {
			dataRow = append(dataRow, valueByColumn[header.Field])
		}
		cell, err := excelize.CoordinatesToCellName(1, rowNumber)
		if err != nil {
			return rowCount, err
		}
		if err := stream.SetRow(cell, dataRow); err != nil {
			return rowCount, err
		}
		rowNumber++
		rowCount++
	}
	if err := rows.Err(); err != nil {
		return rowCount, err
	}
	if err := stream.Flush(); err != nil {
		return rowCount, err
	}
	return rowCount, file.SaveAs(fullPath)
}

func safeDownloadFullPath(relativePath string) (string, error) {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return "", fmt.Errorf("download file path is empty")
	}
	cleanRelative := filepath.Clean(relativePath)
	if filepath.IsAbs(cleanRelative) || strings.HasPrefix(cleanRelative, "..") {
		return "", fmt.Errorf("download file path invalid")
	}
	fullPath := utils.MediaPath(cleanRelative)
	mediaRoot := utils.MediaRoot()
	rel, err := filepath.Rel(mediaRoot, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("download file path escapes media root")
	}
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("download file not found")
	}
	return fullPath, nil
}

func normalizeDownloadCellValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format("2006-01-02 15:04:05")
	default:
		return typed
	}
}

func isEmptyDownloadFilterValue(value any) bool {
	if value == nil {
		return true
	}
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == ""
	}
	return false
}

func isSafeDownloadDBColumn(column string) bool {
	return safeDBColumnPattern.MatchString(strings.TrimSpace(column))
}

func isAllowedDownloadOperator(operator string) bool {
	switch operator {
	case "=", ">=", "<=", "LIKE":
		return true
	default:
		return false
	}
}

func isSafeDownloadOrderBy(orderBy string) bool {
	orderBy = strings.TrimSpace(orderBy)
	if orderBy == "" {
		return true
	}
	blocked := []string{";", "--", "/*", "*/"}
	for _, token := range blocked {
		if strings.Contains(orderBy, token) {
			return false
		}
	}
	parts := strings.Split(orderBy, ",")
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 || len(fields) > 2 {
			return false
		}
		if !isSafeDownloadDBColumn(fields[0]) {
			return false
		}
		if len(fields) == 2 {
			direction := strings.ToUpper(fields[1])
			if direction != "ASC" && direction != "DESC" {
				return false
			}
		}
	}
	return true
}

func normalizeDownloadLimit(value any) int {
	limit := defaultDownloadMaxRows
	switch typed := value.(type) {
	case int:
		limit = typed
	case int64:
		limit = int(typed)
	case float64:
		limit = int(typed)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			limit = parsed
		}
	}
	if limit <= 0 || limit > defaultDownloadMaxRows {
		return defaultDownloadMaxRows
	}
	return limit
}

func excelColumnName(index int) string {
	name := ""
	for index > 0 {
		index--
		name = string(rune('A'+index%26)) + name
		index /= 26
	}
	return name
}

func sanitizeDownloadError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "download task failed"
	}
	if len(message) > 500 {
		return message[:500]
	}
	return message
}
