package models

import "time"

// DownloadTemplate defines one backend-approved export template.
type DownloadTemplate struct {
	ID             uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TemplateCode   string    `gorm:"column:template_code;size:80;uniqueIndex;not null;comment:下载模板编码" json:"template_code"`
	TemplateName   string    `gorm:"column:template_name;size:120;not null;comment:下载模板名称" json:"template_name"`
	BusinessType   string    `gorm:"column:business_type;size:40;index;not null;comment:业务类型 order/product/report/inventory/after_sale" json:"business_type"`
	SQLContent     string    `gorm:"column:sql_content;type:text;not null;comment:SQL内容，仅后端预置，不接收前端SQL" json:"sql_content"`
	ModelFields    string    `gorm:"column:model_fields;type:json;not null;comment:相关模型字段JSON" json:"model_fields"`
	ExportHeaders  string    `gorm:"column:export_headers;type:json;not null;comment:导出列头JSON" json:"export_headers"`
	AllowedFilters string    `gorm:"column:allowed_filters;type:json;comment:允许筛选字段JSON" json:"allowed_filters"`
	DefaultOrderBy string    `gorm:"column:default_order_by;size:120;comment:默认排序白名单值" json:"default_order_by"`
	FileFormat     string    `gorm:"column:file_format;size:20;default:'xlsx';comment:文件格式 xlsx/csv" json:"file_format"`
	Status         string    `gorm:"column:status;size:20;default:'enabled';index;comment:enabled/disabled" json:"status"`
	CreatedBy      int       `gorm:"column:created_by;default:0;comment:创建人后台用户ID" json:"created_by"`
	UpdatedBy      int       `gorm:"column:updated_by;default:0;comment:更新人后台用户ID" json:"updated_by"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (DownloadTemplate) TableName() string {
	return "download_template"
}
