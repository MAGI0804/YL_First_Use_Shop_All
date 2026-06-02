package requestbody

type AddToCartRequest struct {
	UserID        int    `json:"user_id" binding:"required"`
	CommodityCode string `json:"commodity_code" binding:"required"`
	Quantity      int    `json:"quantity"`
}

type BatchDeleteFromCartRequest struct {
	UserID         int      `json:"user_id" binding:"required"`
	CommodityCodes []string `json:"commodity_codes"`
}

type QueryCartItemsRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

type UpdateCartItemQuantityRequest struct {
	UserID        int    `json:"user_id" binding:"required"`
	CommodityCode string `json:"commodity_code" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required"`
}

type IncreaseCartItemQuantityRequest struct {
	UserID        int    `json:"user_id" binding:"required"`
	CommodityCode string `json:"commodity_code" binding:"required"`
}

type DecreaseCartItemQuantityRequest struct {
	UserID        int    `json:"user_id" binding:"required"`
	CommodityCode string `json:"commodity_code" binding:"required"`
}

type ClearCartRequest struct {
	UserID int `json:"user_id" binding:"required"`
}
