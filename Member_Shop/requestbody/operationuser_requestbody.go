package requestbody

// AddServiceUserRequest 添加客服用户请求
type AddServiceUserRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Add                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               OperationUserRequest 添加运营用户请求
type AddOperationUserRequest struct {
	Nickname string      `json:"nickname" binding:"required"`
	Mobile   string      `json:"mobile" binding:"required"`
	Password string      `json:"password" binding:"required"`
	Level    interface{} `json:"level" binding:"required"`
}

// VerificationStatusRequest 验证登录状态请求
type VerificationStatusRequest struct {
	Mobile    string `json:"mobile" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ObjectNum string `json:"object_num" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	ObjectNum   int    `json:"object_num" binding:"required"`
	Mobile      string `json:"mobile" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// SendBackendRegisterCaptchaRequest 发送后台账号注册验证码请求
type SendBackendRegisterCaptchaRequest struct {
	Mobile string `json:"mobile" binding:"required"`
}

// BackendRegisterByPhoneRequest 通过手机验证码注册后台运营账号请求
type BackendRegisterByPhoneRequest struct {
	Mobile   string `json:"mobile" binding:"required"`
	Captcha  string `json:"captcha" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
	Level    int    `json:"level"`
	Remarks  string `json:"remarks"`
}
