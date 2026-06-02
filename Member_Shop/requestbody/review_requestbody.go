package requestbody

// ReviewCreateRequest 创建评价请求结构体
// 用户针对已收货的商品进行评价
type ReviewCreateRequest struct {
	UserID      int      `json:"user_id" binding:"required"`       // 用户ID，必填
	OrderID     string   `json:"order_id" binding:"required"`      // 订单ID，必填
	SubOrderID  string   `json:"sub_order_id" binding:"required"`  // 子订单ID，必填
	CommodityID string   `json:"commodity_id" binding:"required"` // 商品ID，必填
	StyleCode   string   `json:"style_code"`                       // 款式编码
	Rating      int      `json:"rating" binding:"required,min=1,max=5"` // 评分，1-5分，必填
	Content     string   `json:"content"`                          // 评价内容
	Images      []string `json:"images"`                           // 评价图片列表
	Tags        []string `json:"tags"`                             // 评价标签列表
}

// ReviewProductQueryRequest 商品评价查询请求结构体
// 用于前台展示商品评价列表
type ReviewProductQueryRequest struct {
	CommodityID string `json:"commodity_id"` // 商品ID
	StyleCode   string `json:"style_code"`   // 款式编码
	Page        int    `json:"page"`         // 页码
	PageSize    int    `json:"page_size"`   // 每页数量
}

// ReviewBackendQueryRequest 后台评价查询请求结构体
// 用于后台管理查看和筛选评价
type ReviewBackendQueryRequest struct {
	UserID      int    `json:"user_id"`      // 用户ID
	OrderID     string `json:"order_id"`     // 订单ID
	SubOrderID  string `json:"sub_order_id"` // 子订单ID
	CommodityID string `json:"commodity_id"` // 商品ID
	StyleCode   string `json:"style_code"`   // 款式编码
	Status      string `json:"status"`       // 评价状态：pending/approved/rejected/hidden
	Page        int    `json:"page"`         // 页码
	PageSize    int    `json:"page_size"`   // 每页数量
}

// ReviewAuditRequest 评价审核请求结构体
// 后台管理员审核评价
type ReviewAuditRequest struct {
	ReviewID    uint   `json:"review_id" binding:"required"` // 评价ID，必填
	Status      string `json:"status" binding:"required"`    // 审核状态，必填：approved/rejected/hidden
	AuditRemark string `json:"audit_remark"`                // 审核备注
}

// ReviewReplyRequest 评价回复请求结构体
// 运营或客服回复用户评价
type ReviewReplyRequest struct {
	ReviewID   uint   `json:"review_id" binding:"required"`   // 评价ID，必填
	OperatorID string `json:"operator_id" binding:"required"` // 操作员ID，必填
	Content    string `json:"content" binding:"required"`    // 回复内容，必填
}

// ReviewStatisticsRequest 评价统计请求结构体
// 获取商品或款式的评价统计数据
type ReviewStatisticsRequest struct {
	CommodityID string `json:"commodity_id"` // 商品ID
	StyleCode   string `json:"style_code"`  // 款式编码
}
