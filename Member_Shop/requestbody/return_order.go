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
	BuyerProvince   string   `json:"buyer_province"`
	BuyerCity       string   `json:"buyer_city"`
	BuyerCounty     string   `json:"buyer_county"`
	BuyerAddress    string   `json:"buyer_address"`
	BuyerPhone      string   `json:"buyer_phone"`
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

type ReturnOrderStatisticsRequest struct {
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
}

type ReturnOrderPushJushuitanRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
}

type JushuitanAfterSalePushRequest struct {
	ReturnOrderID        string                       `json:"outer_as_id"`
	JushuitanAfterSaleID string                       `json:"as_id"`
	OrderID              string                       `json:"so_id"`
	Status               string                       `json:"status"`
	ShopStatus           string                       `json:"shop_status"`
	RefundStatus         string                       `json:"refund_status"`
	Type                 string                       `json:"type"`
	Modified             string                       `json:"modified"`
	Items                []JushuitanAfterSalePushItem `json:"items"`
}

type JushuitanAfterSalePushItem struct {
	OuterOiID string `json:"outer_oi_id"`
	SkuID     string `json:"sku_id"`
	Qty       int    `json:"qty"`
}

type JushuitanAfterSaleReceivedQueryRequest struct {
	PageIndex     int    `json:"page_index"`
	PageSize      int    `json:"page_size"`
	ModifiedBegin string `json:"modified_begin"`
	ModifiedEnd   string `json:"modified_end"`
	OrderID       string `json:"so_id"`
	ReturnOrderID string `json:"outer_as_id"`
	ASID          string `json:"as_id"`
}
