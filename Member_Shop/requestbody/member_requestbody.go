package requestbody

type MemberCreateRequest struct {
	MemberNo         string  `json:"member_no"`
	UserID           int     `json:"user_id"`
	OpenID           string  `json:"openid"`
	Mobile           string  `json:"mobile" binding:"required"`
	ManualUniqueCode string  `json:"manual_unique_code"`
	Nickname         string  `json:"nickname"`
	Status           string  `json:"status"`
	Source           string  `json:"source"`
	TmallID          string  `json:"tmall_id"`
	TmallAmount      float64 `json:"tmall_amount"`
	YouzanID         string  `json:"youzan_id"`
	YouzanAmount     float64 `json:"youzan_amount"`
	Remarks          string  `json:"remarks"`
}

type MemberUpdateRequest struct {
	ID               uint    `json:"id" binding:"required"`
	MemberNo         string  `json:"member_no"`
	UserID           int     `json:"user_id"`
	OpenID           string  `json:"openid"`
	Mobile           string  `json:"mobile"`
	ManualUniqueCode string  `json:"manual_unique_code"`
	Nickname         string  `json:"nickname"`
	Status           string  `json:"status"`
	Source           string  `json:"source"`
	TotalOrderAmount float64 `json:"total_order_amount"`
	TotalPaidAmount  float64 `json:"total_paid_amount"`
	TmallID          string  `json:"tmall_id"`
	TmallAmount      float64 `json:"tmall_amount"`
	YouzanID         string  `json:"youzan_id"`
	YouzanAmount     float64 `json:"youzan_amount"`
	Remarks          string  `json:"remarks"`
}

type MemberListRequest struct {
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
	Mobile           string `json:"mobile"`
	MemberNo         string `json:"member_no"`
	ManualUniqueCode string `json:"manual_unique_code"`
	Nickname         string `json:"nickname"`
	Status           string `json:"status"`
	TagID            uint   `json:"tag_id"`
	TagName          string `json:"tag_name"`
}

type MemberDetailRequest struct {
	ID       uint   `json:"id"`
	MemberNo string `json:"member_no"`
	Mobile   string `json:"mobile"`
	UserID   int    `json:"user_id"`
}

type MemberTagListRequest struct {
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type MemberTagCreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Color   string `json:"color"`
	Remarks string `json:"remarks"`
}

type MemberTagSetRequest struct {
	MemberID uint   `json:"member_id" binding:"required"`
	TagIDs   []uint `json:"tag_ids"`
}

type MemberCartQueryRequest struct {
	MemberID uint   `json:"member_id"`
	MemberNo string `json:"member_no"`
	Mobile   string `json:"mobile"`
	UserID   int    `json:"user_id"`
}

type MemberCartAddRequest struct {
	MemberID      uint   `json:"member_id"`
	MemberNo      string `json:"member_no"`
	Mobile        string `json:"mobile"`
	UserID        int    `json:"user_id"`
	CommodityCode string `json:"commodity_code" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=1"`
}

type MemberCartUpdateQuantityRequest struct {
	MemberID      uint   `json:"member_id"`
	MemberNo      string `json:"member_no"`
	Mobile        string `json:"mobile"`
	UserID        int    `json:"user_id"`
	CommodityCode string `json:"commodity_code" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=0"`
}

type MemberCartDeleteRequest struct {
	MemberID       uint     `json:"member_id"`
	MemberNo       string   `json:"member_no"`
	Mobile         string   `json:"mobile"`
	UserID         int      `json:"user_id"`
	CommodityCodes []string `json:"commodity_codes"`
}

type BackendOrderItemRequest struct {
	CommodityCode string  `json:"commodity_code" binding:"required"`
	Quantity      int     `json:"quantity" binding:"required,min=1"`
	Price         float64 `json:"price"`
}

type BackendCreateOrderRequest struct {
	MemberID        uint                      `json:"member_id"`
	MemberNo        string                    `json:"member_no"`
	Mobile          string                    `json:"mobile"`
	UserID          int                       `json:"user_id"`
	ReceiverName    string                    `json:"receiver_name" binding:"required"`
	ReceiverPhone   string                    `json:"receiver_phone" binding:"required"`
	Province        string                    `json:"province" binding:"required"`
	City            string                    `json:"city" binding:"required"`
	County          string                    `json:"county" binding:"required"`
	DetailedAddress string                    `json:"detailed_address" binding:"required"`
	Items           []BackendOrderItemRequest `json:"items" binding:"required,dive"`
	OrderAmount     float64                   `json:"order_amount"`
	ExpressCompany  string                    `json:"express_company"`
	ExpressNumber   string                    `json:"express_number"`
	Remark          string                    `json:"remark"`
	BackendRemark   string                    `json:"backend_remark"`
}

type OperationLogQueryRequest struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	OperatorID uint   `json:"operator_id"`
	Action     string `json:"action"`
	Module     string `json:"module"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	MemberID   uint   `json:"member_id"`
	UserID     int    `json:"user_id"`
	OrderID    string `json:"order_id"`
	BeginTime  string `json:"begin_time"`
	EndTime    string `json:"end_time"`
}
