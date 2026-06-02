package requestbody

// ActivityImageRelationRequest 更新活动图片关系请求
type ActivityImageRelationRequest struct {
	ActivityID  float64   `json:"activity_id"`
	StyleCodes  []string  `json:"style_codes"`
	Category    string    `json:"category"`
}

// ActivityImageStatusRequest 活动图片状态变更请求
type ActivityImageStatusRequest struct {
	ActivityID float64 `json:"activity_id"`
}

// ImageOrderRequest 图片顺序请求
type ImageOrderRequest struct {
	ID    int `json:"id"`
	Order int `json:"order"`
}

// BatchUpdateActivityImageOrderRequest 批量更新活动图片顺序请求
type BatchUpdateActivityImageOrderRequest struct {
	Images []ImageOrderRequest `json:"images"`
}

// BatchQueryActivityImagesRequest 批量查询活动图片请求
type BatchQueryActivityImagesRequest struct {
	Page             float64 `json:"page"`
	PageSize         float64 `json:"pageSize"`
	Status           string  `json:"status"`
	StartTime        string  `json:"start_time"`        // 开始时间（格式：2006-01-02 15:04:05）
	EndTime          string  `json:"end_time"`          // 结束时间（格式：2006-01-02 15:04:05）
	HasActivityDetail *bool  `json:"has_activity_detail"` // 是否有活动详情（可选）
}

// AddActivityImgRequest 添加活动图片请求（表单）
type AddActivityImgRequest struct {
	Category    string `form:"category"`
	Notes       string `form:"notes"`
	Commodities string `form:"commodities"`
	// Image 文件字段在控制器中单独处理
}

// AddPromotionalPicRequest 新增宣传图请求
type AddPromotionalPicRequest struct {
	ActivityID float64 `json:"activity_id"` // 活动图id
}

// UpdatePromotionalPicOrderRequest 调整宣传图位置请求
type UpdatePromotionalPicOrderRequest struct {
	ActivityID    float64 `json:"activity_id"`    // 活动图id
	OldOrder      int     `json:"old_order"`      // 原顺序
	NewOrder      int     `json:"new_order"`      // 新顺序
}

// DeletePromotionalPicRequest 删除宣传图请求
type DeletePromotionalPicRequest struct {
	ActivityID float64 `json:"activity_id"` // 活动图id
	Order      int     `json:"order"`       // 要删除的宣传图顺序
}

// GetActivityImageDetailRequest 根据活动图id查询详情请求
type GetActivityImageDetailRequest struct {
	ActivityID float64 `json:"activity_id"` // 活动图id
}

// SetHasActivityDetailRequest 设置活动详情请求
type SetHasActivityDetailRequest struct {
	ActivityID       float64 `json:"activity_id"`       // 活动图id
	HasActivityDetail bool  `json:"has_activity_detail"` // 是否有活动详情
}
