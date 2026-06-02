package requestbody

// UserQueryRequest 用户查询入参
type UserQueryRequest struct {
	UserId int `json:"user_id" required:"true"`
}

// QueryUserIdByMobileRequest 根据手机号查询用户ID入参
type QueryUserIdByMobileRequest struct {
	Mobile string `json:"mobile" required:"true"`
}

// WechatLoginRequest 微信登录/注册入参
type WechatLoginRequest struct {
	Code string `json:"code" required:"true"`
}

// AddUserData 添加用户数据入参
type AddUserData struct {
	UserId    int    `json:"user_id" required:"true"`
	DataType  string `json:"data_type" required:"true"`
	DataValue string `json:"data_value" required:"true"`
}

// FindDataRequest 查询用户数据入参
type FindDataRequest struct {
	UserId int `json:"user_id" required:"true"`
}

// SendRegisterCaptchaRequest 发送注册验证码请求
type SendRegisterCaptchaRequest struct {
	Mobile string `json:"mobile" binding:"required"`
}

// BindWechatPhoneRequest 绑定微信手机号请求
type BindWechatPhoneRequest struct {
	OpenID  string `json:"openid" binding:"required"`
	Mobile  string `json:"mobile" binding:"required"`
	Captcha string `json:"captcha"`
}

// UpdatePlatformInfoRequest 更新平台信息请求
type UpdatePlatformInfoRequest struct {
	UserID       int     `json:"user_id" binding:"required"`
	TmallID      string  `json:"tmall_id"`
	TmallAmount  float64 `json:"tmall_amount"`
	YouzanID     string  `json:"youzan_id"`
	YouzanAmount float64 `json:"youzan_amount"`
}

// MemberAmountSummaryRequest 会员金额汇总请求
type MemberAmountSummaryRequest struct {
	UserID   int    `json:"user_id"`
	MemberNo string `json:"member_no"`
	Mobile   string `json:"mobile"`
}
