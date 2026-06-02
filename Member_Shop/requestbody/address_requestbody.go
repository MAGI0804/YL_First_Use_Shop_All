package requestbody

// AddAddressRequest 新增地址请求
type AddAddressRequest struct {
	UserID          int    `json:"user_id" binding:"required"`
	Province        string `json:"province" binding:"required"`
	City            string `json:"city" binding:"required"`
	County          string `json:"county" binding:"required"`
	DetailedAddress string `json:"detailed_address" binding:"required"`
	ReceiverName    string `json:"receiver_name" binding:"required"`
	PhoneNumber     string `json:"phone_number" binding:"required"`
	IsDefault       bool   `json:"is_default"`
	Remark          string `json:"remark"`
}

// DeleteAddressRequest 删除地址请求
type DeleteAddressRequest struct {
	AddressID int `json:"address_id" binding:"required"`
	UserID    int `json:"user_id" binding:"required"`
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	AddressID       int    `json:"address_id" binding:"required"`
	UserID          int    `json:"user_id" binding:"required"`
	Province        string `json:"province"`
	City            string `json:"city"`
	County          string `json:"county"`
	DetailedAddress string `json:"detailed_address"`
	ReceiverName    string `json:"receiver_name"`
	PhoneNumber     string `json:"phone_number"`
	IsDefault       bool   `json:"is_default"`
	Remark          string `json:"remark"`
}

// SetDefaultAddressRequest 设置默认地址请求
type SetDefaultAddressRequest struct {
	AddressID int `json:"address_id" binding:"required"`
	UserID    int `json:"user_id" binding:"required"`
}

// GetAddressesRequest 获取地址列表请求
type GetAddressesRequest struct {
	UserID int `json:"user_id" binding:"required"`
}

// GetAddressByIDRequest 获取地址详情请求
type GetAddressByIDRequest struct {
	AddressID int `json:"address_id" binding:"required"`
	UserID    int `json:"user_id" binding:"required"`
}
