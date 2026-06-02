package requestbody

type TokenRedis struct {
	IP          string `redis:"ip"`
	AccessToken string `gorm:"column:access_token"`
}

type RequestData struct {
	ShopName string `json:"shopname"`
}
