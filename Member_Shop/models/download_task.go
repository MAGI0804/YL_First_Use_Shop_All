package models

import "time"

// DownloadTask records one requested export generation.
type DownloadTask struct {
	TaskID        string     `gorm:"column:task_id;primaryKey;size:40;comment:任务ID" json:"task_id"`
	TemplateCode  string     `gorm:"column:template_code;size:80;index;not null;comment:模板编码" json:"template_code"`
	BusinessType  string     `gorm:"column:business_type;size:40;index;not null;comment:业务类型" json:"business_type"`
	TaskName      string     `gorm:"column:task_name;size:160;not null;comment:任务名称" json:"task_name"`
	Filters       string     `gorm:"column:filters;type:json;comment:筛选条件快照JSON" json:"filters"`
	Status        string     `gorm:"column:status;size:20;index;not null;comment:pending/running/success/failed/expired" json:"status"`
	Progress      int        `gorm:"column:progress;default:0;comment:生成进度0-100" json:"progress"`
	RowCount      int64      `gorm:"column:row_count;default:0;comment:导出行数" json:"row_count"`
	FilePath      string     `gorm:"column:file_path;size:255;comment:生成文件相对路径" json:"file_path"`
	FileName      string     `gorm:"column:file_name;size:180;comment:下载文件名" json:"file_name"`
	FileSize      int64      `gorm:"column:file_size;default:0;comment:文件大小字节" json:"file_size"`
	ErrorMessage  string     `gorm:"column:error_message;type:text;comment:失败原因" json:"error_message"`
	DownloadCount int        `gorm:"column:download_count;default:0;comment:下载次数" json:"download_count"`
	RequestedBy   int        `gorm:"column:requested_by;index;not null;comment:申请人后台用户ID" json:"requested_by"`
	StartedAt     *time.Time `gorm:"column:started_at;comment:开始生成时间" json:"started_at"`
	FinishedAt    *time.Time `gorm:"column:finished_at;comment:生成完成时间" json:"finished_at"`
	ExpiresAt     *time.Time `gorm:"column:expires_at;index;comment:过期时间" json:"expires_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (DownloadTask) TableName() string {
	return "download_task"
}
