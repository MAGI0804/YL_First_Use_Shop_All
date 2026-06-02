package requestbody

// StyleCodeSearchRequest is the request for searching by style code
type StyleCodeSearchRequest struct {
	Shopname  string `json:"shop_name" binding:"required"` // 店铺名称
	StyleCode string `json:"style_code"`                   // 商品款式编码
	Page      int    `json:"page"`                         // 页码
	PageSize  int    `json:"page_size"`                    // 每页数量
}
