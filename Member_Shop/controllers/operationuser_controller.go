package controllers

import (
	"Member_shop/config"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	mrand "math/rand"
	"net/http"
	"regexp"
	"time"

	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	sms "Member_shop/service/sms"
	"Member_shop/utils"

	"github.com/gin-gonic/gin"
)

// OperationUserController 运营用户管理控制器
// 负责处理运营用户和客服用户相关的HTTP请求
type OperationUserController struct{}

// AddServiceUser 处理添加客服用户请求
// 验证必填字段（昵称、手机号、密码），调用service层创建客服账号
func (ouc *OperationUserController) AddServiceUser(c *gin.Context) {
	requestID := generateRequestID()
	log.Printf("add_service_user request received, request_id=%s", requestID)

	if c.Request.Method != "POST" {
		log.Printf("add_service_user received non-POST request, request_id=%s", requestID)
		c.JSON(http.StatusMethodNotAllowed, msg.ErrResponseStr("Method not allowed"))
		return
	}

	var req requestbody.AddServiceUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid JSON format in add_service_user request, request_id=%s, error=%v", requestID, err)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid JSON format"))
		return
	}

	log.Printf("add_service_user parameters: nickname=%s, mobile=%s, request_id=%s",
		req.Nickname, req.Mobile, requestID)

	if req.Nickname == "" || req.Mobile == "" || req.Password == "" {
		log.Printf("Missing required fields in add_service_user, request_id=%s", requestID)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Missing required fields"))
		return
	}

	userID, err := method.AddServiceUser(req)
	if err != nil {
		log.Printf("AddServiceUser failed, request_id=%s, error=%v", requestID, err)
		if err.Error() == "手机号已存在" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Mobile number already exists"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("Server error occurred"))
		return
	}

	data := map[string]any{"user_id": userID}
	log.Printf("Customer service user created successfully: %s, request_id=%s", userID, requestID)
	c.JSON(http.StatusCreated, msg.SuccessResponse("Customer service user created successfully", &data))
}

// AddOperationUser 处理添加运营用户请求
// 验证必填字段（昵称、手机号、密码），支持level字段设置运营级别
func (ouc *OperationUserController) AddOperationUser(c *gin.Context) {
	requestID := generateRequestID()
	log.Printf("add_operation_user request received, request_id=%s", requestID)

	if c.Request.Method != "POST" {
		log.Printf("add_operation_user received non-POST request, request_id=%s", requestID)
		c.JSON(http.StatusMethodNotAllowed, msg.ErrResponseStr("Method not allowed"))
		return
	}

	var req requestbody.AddOperationUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid JSON format in add_operation_user request, request_id=%s, error=%v", requestID, err)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid JSON format"))
		return
	}

	log.Printf("add_operation_user parameters: nickname=%s, mobile=%s, request_id=%s",
		req.Nickname, req.Mobile, requestID)

	if req.Nickname == "" || req.Mobile == "" || req.Password == "" {
		log.Printf("Missing required fields in add_operation_user, request_id=%s", requestID)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Missing required fields"))
		return
	}

	userID, err := method.AddOperationUser(req)
	if err != nil {
		log.Printf("AddOperationUser failed, request_id=%s, error=%v", requestID, err)
		errMsg := err.Error()
		if errMsg == "手机号已存在" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Mobile number already exists"))
			return
		}
		if errMsg == "level格式无效，必须是数字" || errMsg == "level类型无效" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid level format"))
			return
		}
		if errMsg == "level不能为0" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Level cannot be zero"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("Server error occurred"))
		return
	}

	data := map[string]any{"user_id": userID}
	log.Printf("Operation user created successfully: %s, request_id=%s", userID, requestID)
	c.JSON(http.StatusCreated, msg.SuccessResponse("Operation user created successfully", &data))
}

// VerificationStatus 处理登录状态验证请求
// 根据手机号和密码验证用户登录状态，支持运营用户和客服用户两种类型
func (ouc *OperationUserController) VerificationStatus(c *gin.Context) {
	var req requestbody.VerificationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的JSON格式"))
		return
	}

	if req.Mobile == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("缺少必要参数"))
		return
	}

	mobileRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !mobileRegex.MatchString(req.Mobile) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("手机号格式错误"))
		return
	}

	result, err := method.VerificationStatus(req)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "object_num参数错误" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("object_num参数错误"))
			return
		}
		if errMsg == "手机号未注册" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("手机号未注册"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("用户信息查询失败，请稍后重试"))
		return
	}

	data := map[string]any{
		"user_id":  result.UserID,
		"nickname": result.Nickname,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("登录状态验证成功", &data))
}

// ChangePassword 处理修改密码请求
// 验证必填字段，验证旧密码是否正确，支持运营用户和客服用户两种类型
func (ouc *OperationUserController) ChangePassword(c *gin.Context) {
	requestID := generateRequestID()
	log.Printf("change_password request received, request_id=%s", requestID)

	if c.Request.Method != "POST" {
		log.Printf("change_password received non-POST request, request_id=%s", requestID)
		c.JSON(http.StatusMethodNotAllowed, msg.ErrResponseStr("Method not allowed"))
		return
	}

	var req requestbody.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid JSON format, request_id=%s, error=%v", requestID, err)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid JSON format"))
		return
	}

	if req.ObjectNum == 0 || req.Mobile == "" || req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Missing required fields"))
		return
	}

	if req.ObjectNum != 1 && req.ObjectNum != 2 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("object_num must be 1 or 2"))
		return
	}

	mobileRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !mobileRegex.MatchString(req.Mobile) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid mobile format"))
		return
	}

	err := method.ChangePassword(req)
	if err != nil {
		log.Printf("ChangePassword failed, request_id=%s, error=%v", requestID, err)
		errMsg := err.Error()
		if errMsg == "用户不存在" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("User not found"))
			return
		}
		if errMsg == "旧密码错误" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Old password is incorrect"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("Server error after multiple attempts"))
		return
	}

	log.Printf("Password updated successfully, request_id=%s", requestID)
	c.JSON(http.StatusOK, msg.SuccessResponseStr("Password updated successfully"))
}

// SendBackendRegisterCaptcha 发送后台注册验证码
// 生成6位数字验证码，保存到数据库并发送短信到用户手机
func (ouc *OperationUserController) SendBackendRegisterCaptcha(c *gin.Context) {
	var req requestbody.SendBackendRegisterCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	if !method.IsValidMobile(req.Mobile) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid mobile"))
		return
	}
	if err := method.CanSendBackendRegisterCaptcha(req.Mobile); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	captcha := db.GenerateCaptcha()
	if err := db.SaveCaptcha(req.Mobile, captcha); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	if _, err := sms.SendSms(req.Mobile, captcha); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("Send sms failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("Captcha sent successfully"))
}

// BackendRegisterByPhone 手机号注册后台运营用户
// 验证码验证成功后，在backend_operation_user表中创建新的运营账号
func (ouc *OperationUserController) BackendRegisterByPhone(c *gin.Context) {
	var req requestbody.BackendRegisterByPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}

	backendUser, err := method.RegisterBackendUserByPhone(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	accessToken, refreshToken, err := utils.GenerateTokens(int(backendUser.ID), config.LoadConfig())
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(backendUser, accessToken, refreshToken))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend user registered successfully", &data))
}

// BackendLogin logs an active backend account in with mobile and password.
func (ouc *OperationUserController) BackendLogin(c *gin.Context) {
	var req requestbody.BackendLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	backendUser, accessToken, refreshToken, err := method.BackendLogin(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr(err.Error()))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(backendUser, accessToken, refreshToken))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend login successful", &data))
}

// BackendMe returns the current backend user and permissions from a valid token.
func (ouc *OperationUserController) BackendMe(c *gin.Context) {
	backendUser := currentBackendUser(c)
	if backendUser == nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr("backend user missing"))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(backendUser, "", ""))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend token valid", &data))
}

// AddBackendUserInvite creates a pending backend account by mobile and nickname.
func (ouc *OperationUserController) AddBackendUserInvite(c *gin.Context) {
	if !method.IsBackendAdmin(currentBackendUser(c)) {
		c.JSON(http.StatusForbidden, msg.ErrResponseStr("admin permission required"))
		return
	}
	var req requestbody.AddBackendUserInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	user, err := method.AddBackendUserInvite(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(user, "", ""))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend user invited", &data))
}

// QueryBackendUsers lists backend accounts for administrators.
func (ouc *OperationUserController) QueryBackendUsers(c *gin.Context) {
	if !method.IsBackendAdmin(currentBackendUser(c)) {
		c.JSON(http.StatusForbidden, msg.ErrResponseStr("admin permission required"))
		return
	}
	var req requestbody.QueryBackendUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	users, total, err := method.QueryBackendUsers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	items := make([]method.BackendUserSession, 0, len(users))
	for i := range users {
		items = append(items, method.BuildBackendUserSession(&users[i], "", ""))
	}
	data := map[string]any{
		"items":     items,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend users queried", &data))
}

// UpdateBackendUserStatus changes a backend account status.
func (ouc *OperationUserController) UpdateBackendUserStatus(c *gin.Context) {
	if !method.IsBackendAdmin(currentBackendUser(c)) {
		c.JSON(http.StatusForbidden, msg.ErrResponseStr("admin permission required"))
		return
	}
	var req requestbody.UpdateBackendUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	user, err := method.UpdateBackendUserStatus(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(user, "", ""))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend user status updated", &data))
}

// UpdateBackendUser updates role, status, and page permissions for an account.
func (ouc *OperationUserController) UpdateBackendUser(c *gin.Context) {
	if !method.IsBackendAdmin(currentBackendUser(c)) {
		c.JSON(http.StatusForbidden, msg.ErrResponseStr("admin permission required"))
		return
	}
	var req requestbody.UpdateBackendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	user, err := method.UpdateBackendUser(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := backendSessionData(method.BuildBackendUserSession(user, "", ""))
	c.JSON(http.StatusOK, msg.SuccessResponse("Backend user updated", &data))
}

func backendSessionData(session method.BackendUserSession) map[string]any {
	return map[string]any{"user": session}
}

func currentBackendUser(c *gin.Context) *models.BackendUser {
	userValue, ok := c.Get("backendUser")
	if !ok {
		return nil
	}
	user, ok := userValue.(*models.BackendUser)
	if !ok {
		return nil
	}
	return user
}

// generateRequestID 生成唯一请求ID
// 用于日志追踪和请求唯一标识，使用加密随机数生成
func generateRequestID() string {
	bytes := make([]byte, 16)
	_, err := crand.Read(bytes)
	if err != nil {
		mrand.Seed(time.Now().UnixNano())
		randomNum := mrand.Intn(0x100000000)
		return fmt.Sprintf("%d%08x", time.Now().UnixNano(), randomNum)
	}
	return hex.EncodeToString(bytes)
}
