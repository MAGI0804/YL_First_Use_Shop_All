package requestbody

type QueryOrdersRequest struct {
	Shopname string `json:"shopname" binding:"required"`
	UserID   int    `json:"user_id" binding:"required"`
	Status   string `json:"status" binding:"omitempty"`
	Page     int    `json:"page" binding:"required,min=1"`
	PageSize int    `json:"page_size" binding:"required,min=1,max=50"`
}

type OrderListRequest struct {
	Shopname  string `json:"shopname" binding:"required"`
	UserID    int    `json:"user_id"`
	Status    string `json:"status"`
	Page      int    `json:"page" binding:"required,min=1"`
	PageSize  int    `json:"page_size" binding:"required,min=1,max=50"`
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
	Tid       string `json:"tid"`
}

type OrderDetailRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	UserID  int    `json:"user_id"`
}

type ChangeStatusRequest struct {
	OrderID          string      `json:"order_id" binding:"required"`
	Status           string      `json:"status" binding:"required"`
	ExpressCompany   string      `json:"express_company"`
	ExpressNumber    string      `json:"express_number"`
	LogisticsProcess interface{} `json:"logistics_process"`
}

type OrderCreateRequest struct {
	ReceiverName    string        `json:"receiver_name" binding:"required"`
	Province        string        `json:"province" binding:"required"`
	City            string        `json:"city" binding:"required"`
	County          string        `json:"county" binding:"required"`
	DetailedAddress string        `json:"detailed_address" binding:"required"`
	OrderAmount     float64       `json:"order_amount" binding:"required"`
	ProductList     []interface{} `json:"product_list" binding:"required,dive"`
	UserID          int           `json:"user_id" binding:"required"`
	ReceiverPhone   interface{}   `json:"receiver_phone"`
	ExpressCompany  string        `json:"express_company"`
	ExpressNumber   string        `json:"express_number"`
	Remark          string        `json:"remark"`
}

type OrderCancelRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	UserID  int    `json:"user_id" binding:"required"`
}

type OrderDeliverRequest struct {
	OrderID        string `json:"order_id" binding:"required"`
	UserID         int    `json:"user_id" binding:"required"`
	ExpressCompany string `json:"express_company" binding:"required"`
	ExpressNumber  string `json:"express_number" binding:"required"`
}

type OrderReceiveRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	UserID  int    `json:"user_id" binding:"required"`
}

type OrderRequestReturnRequest struct {
	OrderID         string   `json:"order_id" binding:"required"`
	UserID          int      `json:"user_id" binding:"required"`
	OrderStatus     string   `json:"order_status" binding:"required"`
	Type            string   `json:"type" binding:"omitempty,oneof=return exchange refund"`
	Reason          string   `json:"reason" binding:"required"`
	SpecificReasons string   `json:"specific_reasons" binding:"required"`
	BuyerProvince   string   `json:"buyer_province" `
	BuyerCity       string   `json:"buyer_city" `
	BuyerCounty     string   `json:"buyer_county" `
	BuyerAddress    string   `json:"buyer_address"`
	BuyerPhone      string   `json:"buyer_phone" `
	ProductIDs      []string `json:"product_ids" binding:"omitempty"`
}

type SyncLogisticsInfoRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

type ChangeReceivingDataRequest struct {
	OrderID         string `json:"order_id" binding:"required"`
	ReceiverName    string `json:"receiver_name" binding:"required"`
	ReceiverPhone   string `json:"receiver_phone"`
	Province        string `json:"province" binding:"required"`
	City            string `json:"city" binding:"required"`
	County          string `json:"county" binding:"required"`
	DetailedAddress string `json:"detailed_address" binding:"required"`
}

type ReturnOrderDeliverRequest struct {
	ReturnOrderID  string `json:"return_order_id" binding:"required"`
	UserID         int    `json:"user_id" binding:"required"`
	ExpressCompany string `json:"express_company" binding:"required"`
	ExpressNumber  string `json:"express_number" binding:"required"`
}

type ReturnOrderReceiveRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
	UserID        int    `json:"user_id" binding:"required"`
}

type ReturnOrderCancelRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
	UserID        int    `json:"user_id" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

type ReturnOrderUpdateBuyerInfoRequest struct {
	ReturnOrderID string `json:"return_order_id" binding:"required"`
	UserID        int    `json:"user_id" binding:"required"`
	BuyerProvince string `json:"buyer_province" binding:"required"`
	BuyerCity     string `json:"buyer_city" binding:"required"`
	BuyerCounty   string `json:"buyer_county" binding:"required"`
	BuyerAddress  string `json:"buyer_address" binding:"required"`
	BuyerPhone    string `json:"buyer_phone" binding:"required"`
}

// OrderSearchByProductNameRequest 按商品名称搜索订单请求体
type OrderSearchByProductNameRequest struct {
	Shopname    string `json:"shopname" binding:"required"`
	UserID      int    `json:"user_id"`
	ProductName string `json:"product_name" binding:"required"`
	Status      string `json:"status"`
	Page        int    `json:"page" binding:"required,min=1"`
	PageSize    int    `json:"page_size" binding:"required,min=1,max=50"`
	BeginTime   string `json:"begin_time"`
	EndTime     string `json:"end_time"`
	Tid         string `json:"tid"`
}

type SubOrderDetailRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

type ChangeSubOrderStatusRequest struct {
	SubOrderID string `json:"sub_order_id" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

type SubOrderCancelRequest struct {
	SubOrderID string `json:"sub_order_id" binding:"required"`
	UserID     int    `json:"user_id"`
	Reason     string `json:"reason"`
}

type SubOrderReturnRequest struct {
	SubOrderID      string `json:"sub_order_id" binding:"required"`
	UserID          int    `json:"user_id"`
	Reason          string `json:"reason" binding:"required"`
	SpecificReasons string `json:"specific_reasons" binding:"required"`
	BuyerProvince   string `json:"buyer_province"`
	BuyerCity       string `json:"buyer_city"`
	BuyerCounty     string `json:"buyer_county"`
	BuyerAddress    string `json:"buyer_address"`
	BuyerPhone      string `json:"buyer_phone"`
}

type JushuitanShipInfoRequest struct {
	SoID             string                  `json:"so_id" binding:"required"`
	OID              int                     `json:"o_id"`
	LID              string                  `json:"l_id"`
	LCID             string                  `json:"lc_id"`
	OrderFrom        string                  `json:"order_from"`
	WmsCoID          int                     `json:"wms_co_id"`
	LogisticsCompany string                  `json:"logistics_company"`
	SendDate         string                  `json:"send_date"`
	IsSendAll        bool                    `json:"is_send_all"`
	Items            []JushuitanShipItemInfo `json:"items"`
}

type JushuitanShipItemInfo struct {
	OID       int    `json:"oi_id"`
	SkuID     string `json:"sku_id"`
	Qty       int    `json:"qty"`
	Name      string `json:"name"`
	OuterOiID string `json:"outer_oi_id"`
	SoID      string `json:"so_id"`
}
