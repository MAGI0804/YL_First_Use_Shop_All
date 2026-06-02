package requestbody

// ReturnOrderCreateRequest 创建售后订单请求体
// 兼容旧字段 return_type/return_reason，同时支持新售后入口需要的 type/reason/specific_reasons/sub_order_id。
type ReturnOrderCreateRequest struct {
	OrderID         string   `json:"order_id" binding:"required"`
	UserID          int      `json:"user_id" binding:"required"`
	SubOrderID      string   `json:"sub_order_id"`
	ReturnReason    string   `json:"return_reason"`
	Reason          string   `json:"reason"`
	ReturnType      string   `json:"return_type"` // 兼容旧字段：refund、exchange、return_refund
	Type            string   `json:"type"`        // 新字段：return、exchange、refund
	SpecificReasons string   `json:"specific_reasons"`
	ReturnAmount    float64  `json:"return_amount"`
	ProductIDs      []string `json:"product_ids"`
	BuyerProvince   string   `json:"buyer_province" binding:"required"`
	BuyerCity       string   `json:"buyer_city" binding:"required"`
	BuyerCounty     string   `json:"buyer_county" binding:"required"`
	BuyerAddress    string   `json:"buyer_address" binding:"required"`
	BuyerPhone      string   `json:"buyer_phone" binding:"required"`
	Remark          string   `json:"remark"`
}

// ReturnOrderQueryRequest 退货订单查询请求体
type ReturnOrderQueryRequest struct {
	ReturnOrderID string `json:"return_order_id"`
	OrderID       string `json:"order_id"`
	UserID        int    `json:"user_id"`
	Status        string `json:"status"`
	Page          int    `json:"page" binding:"min=1"`
	PageSize      int    `json:"page_size" binding:"min=1,max=100"`
}

// ReturnOrderApproveRequest 退货订单审核请求体
type ReturnOrderApproveRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
	ApproveStatus string `json:"approve_status" binding:"required"` // approved, rejected
	UserID        int    `json:"user_id" binding:"required"`
	Remark        string `json:"remark"`
}

// ReturnOrderDetailRequest 退货订单详情查询请求体
type ReturnOrderDetailRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
}
