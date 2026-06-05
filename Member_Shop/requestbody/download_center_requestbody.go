package requestbody

// CreateDownloadTaskRequest creates a download-center generation task.
type CreateDownloadTaskRequest struct {
	TemplateCode string         `json:"template_code" binding:"required"`
	Filters      map[string]any `json:"filters"`
	FileFormat   string         `json:"file_format"`
}

// QueryDownloadTasksRequest lists download-center tasks.
type QueryDownloadTasksRequest struct {
	Page         int    `form:"page" json:"page"`
	PageSize     int    `form:"page_size" json:"page_size"`
	Status       string `form:"status" json:"status"`
	BusinessType string `form:"business_type" json:"business_type"`
	TemplateCode string `form:"template_code" json:"template_code"`
}
