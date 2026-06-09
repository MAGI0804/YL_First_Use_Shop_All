package controllers

import (
	"Member_shop/config"
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	sms "Member_shop/service/sms"
	"Member_shop/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 用户控制器
type UserController struct{}

var (
	wechatHTTPClient = &http.Client{Timeout: 8 * time.Second}
	wechatTokenCache = struct {
		sync.Mutex
		token     string
		expiresAt time.Time
	}{}
)

// 查询用户信息
func (uc *UserController) QueryUserInfo(t *gin.Context) {
	var req requestbody.UserQueryRequest
	//校验参数
	if err := t.ShouldBindJSON(&req); err != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}
	//校验是否存在
	if ex := method.SearchExistence("users_user", "user_id", req.UserId); ex != true {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("用户不存在"))
		return
	}

	t.JSON(http.StatusOK, msg.SuccessResponse("查询成功", method.SelectUserInfo(*t, req.UserId)))
	return
}

// UserModify 修改用户信息
func (uc *UserController) UserModify(c *gin.Context) {
	// 生成请求ID用于日志跟踪
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	log.Printf("[UserModify] 请求开始，requestID: %s, Content-Type: %s", requestID, c.ContentType())

	// 检查是否为multipart/form-data请求
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		// 处理文件上传请求
		log.Printf("[UserModify] 检测到multipart/form-data请求，requestID: %s", requestID)
		userIDStr := c.PostForm("user_id")
		if userIDStr == "" {
			log.Printf("[UserModify] 缺少user_id参数，requestID: %s", requestID)
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("参数错误"))
			return
		}

		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)

		// 查询用户
		var user models.User
		if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
			log.Printf("[UserModify] 用户不存在: %d, requestID: %s", userID, requestID)
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("用户不存在"))
			return
		}
		// 处理普通字段更新
		nickname := c.PostForm("nickname")
		if nickname != "" {
			user.Nickname = nickname
		}

		defaultReceiver := c.PostForm("default_receiver")
		if defaultReceiver != "" {
			user.DefaultReceiver = defaultReceiver
		}

		province := c.PostForm("province")
		if province != "" {
			user.Province = province
		}

		city := c.PostForm("city")
		if city != "" {
			user.City = city
		}

		county := c.PostForm("county")
		if county != "" {
			user.County = county
		}

		detailedAddress := c.PostForm("detailed_address")
		if detailedAddress != "" {
			user.DetailedAddress = detailedAddress
		}

		membershipLevel := c.PostForm("membership_level")
		if membershipLevel != "" {
			var level int
			fmt.Sscanf(membershipLevel, "%d", &level)
			user.MembershipLevel = level
		}

		remarks := c.PostForm("remarks")
		if remarks != "" {
			user.Remarks = remarks
		}

		// 处理头像上传
		log.Printf("[UserModify] 尝试获取上传文件，requestID: %s", requestID)
		file, header, err := c.Request.FormFile("user_img")
		if err != nil {
			log.Printf("[UserModify] 获取文件失败: %v, requestID: %s", err, requestID)
		} else if header != nil && file != nil {
			log.Printf("[UserModify] 成功获取文件: %s, 大小: %d字节, requestID: %s", header.Filename, header.Size, requestID)
			// 生成唯一文件名
			uniqueFilename := utils.GenerateUniqueFilename(header.Filename)

			// 定义保存路径 - 只保存相对路径，不包含media前缀
			savePath := fmt.Sprintf("user_avatars/%s", uniqueFilename)

			// 确保目录存在 - 使用与静态文件服务一致的相对路径
			log.Printf("[UserModify] 尝试创建目录: ./media/user_avatars, requestID: %s", requestID)
			if err := os.MkdirAll(utils.MediaPath("user_avatars"), 0755); err != nil {
				log.Printf("[UserModify] 创建目录失败: %v, requestID: %s", err, requestID)
			}

			// 创建文件 - 使用相对路径
			fullPath := utils.MediaPath(savePath)
			log.Printf("[UserModify] 尝试创建文件: %s, requestID: %s", fullPath, requestID)
			dst, err := os.Create(fullPath)
			if err != nil {
				log.Printf("[UserModify] 创建文件失败: %v, requestID: %s", err, requestID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
				return
			}
			defer dst.Close()

			// 复制文件内容
			log.Printf("[UserModify] 开始复制文件内容, requestID: %s", requestID)
			bytesWritten, err := io.Copy(dst, file)
			if err != nil {
				log.Printf("[UserModify] 文件复制失败: %v, requestID: %s", err, requestID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
				return
			}
			log.Printf("[UserModify] 文件复制成功, 写入字节数: %d, requestID: %s", bytesWritten, requestID)

			// 更新用户头像路径
			user.UserImg = savePath
			log.Printf("[UserModify] 更新用户头像路径: %s, requestID: %s", savePath, requestID)
		}

		// 保存更新
		if err := db.DB.Save(&user).Error; err != nil {
			log.Printf("更新用户信息失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器处理失败"})
			return
		}

		c.JSON(http.StatusOK, msg.SuccessResponseStr("信息修改成功"))
		return
	}

	// 处理普通JSON请求
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON格式"})
		return
	}

	userIDFloat, ok := requestData["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少user_id参数"})
		return
	}

	userID := int(userIDFloat)

	// 查询用户
	var user models.User
	if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查是否尝试修改手机号
	if _, ok := requestData["mobile"]; ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "禁止修改手机号"})
		return
	}

	// 更新字段
	if nickname, ok := requestData["nickname"].(string); ok {
		user.Nickname = nickname
	}

	if defaultReceiver, ok := requestData["default_receiver"].(string); ok {
		user.DefaultReceiver = defaultReceiver
	}

	if province, ok := requestData["province"].(string); ok {
		user.Province = province
	}

	if city, ok := requestData["city"].(string); ok {
		user.City = city
	}

	if county, ok := requestData["county"].(string); ok {
		user.County = county
	}

	if detailedAddress, ok := requestData["detailed_address"].(string); ok {
		user.DetailedAddress = detailedAddress
	}

	if membershipLevel, ok := requestData["membership_level"].(float64); ok {
		user.MembershipLevel = int(membershipLevel)
	}

	if remarks, ok := requestData["remarks"].(string); ok {
		user.Remarks = remarks
	}

	// 保存更新
	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("更新用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器处理失败"})
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("信息修改成功"))
}

// UserGetID 根据手机号获取用户ID
func (uc *UserController) UserGetID(t *gin.Context) {
	var req requestbody.QueryUserIdByMobileRequest
	//校验参数
	if err := t.ShouldBindJSON(&req); err != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}
	//校验是否存在
	if ex := method.SearchExistence("users_user", "mobile", req.Mobile); ex != true {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("用户不存在"))
		return
	}
	userid, err := method.GetField("users_user", "mobile", req.Mobile, "user_id")
	if err != nil {
		t.JSON(http.StatusUnauthorized, msg.ErrResponse("查询失败", err))
		return
	}
	if userid == "" {
		t.JSON(http.StatusUnauthorized, msg.ErrResponse("查询失败", err))
		return
	}
	info := map[string]any{
		"user_id": userid,
	}
	t.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
	return
}

// WechatLogin 微信小程序登录
func (uc *UserController) WechatLogin(c *gin.Context) {
	var req requestbody.WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON格式"})
		return
	}

	// 获取微信配置
	cfg := config.LoadConfig()

	openid, err := getWechatOpenID(cfg, req.Code)
	if err != nil {
		log.Printf("微信登录失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mobile, err := getWechatPhoneNumber(cfg, req.PhoneCode)
	if err != nil {
		log.Printf("微信手机号授权失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析昵称
	nickname := ""
	if nicknameVal, ok := req.UserInfo["nickName"].(string); ok {
		nickname = nicknameVal
	} else if nicknameVal, ok := req.UserInfo["nickname"].(string); ok {
		nickname = nicknameVal
	}

	// 解析头像URL
	avatarURL := ""
	if avatarVal, ok := req.UserInfo["avatarUrl"].(string); ok {
		avatarURL = avatarVal
	} else if avatarVal, ok := req.UserInfo["avatar_url"].(string); ok {
		avatarURL = avatarVal
	}

	member, err := method.BindWechatPhone(requestbody.BindWechatPhoneRequest{
		OpenID:    openid,
		Mobile:    mobile,
		Nickname:  nickname,
		AvatarURL: avatarURL,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "mobile is not a member" || err.Error() == "member disabled" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"code": status, "message": memberLoginErrorMessage(err)})
		return
	}

	var user models.User
	if err := db.DB.Where("user_id = ?", member.UserID).First(&user).Error; err != nil {
		log.Printf("查询微信用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	// 生成令牌
	accessToken, refreshToken, err := utils.GenerateTokens(user.UserID, cfg)
	if err != nil {
		log.Printf("生成令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	responseData := gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": gin.H{
				"access":  accessToken,
				"refresh": refreshToken,
			},
			"user_id":     user.UserID,
			"member_no":   member.MemberNo,
			"mobile":      member.Mobile,
			"phone_bound": true,
			"nickname":    user.Nickname,
		},
	}

	// 如果有头像，返回完整的头像URL
	if user.UserImg != "" {
		// 获取请求的协议，考虑反向代理环境
		proto := utils.GetRequestProto(c)
		baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
		// 检查头像URL是否已经是完整URL，如果不是则构建完整URL
		if !strings.HasPrefix(user.UserImg, "http://") && !strings.HasPrefix(user.UserImg, "https://") {
			responseData["data"].(gin.H)["avatar_url"] = utils.BuildFullImageURL(baseURL, user.UserImg, "media")
		} else {
			responseData["data"].(gin.H)["avatar_url"] = user.UserImg
		}
	}

	c.JSON(http.StatusOK, responseData)
}

func getWechatOpenID(cfg config.Config, code string) (string, error) {
	wxURL := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.WechatConfig.LoginURL,
		cfg.WechatConfig.AppID,
		cfg.WechatConfig.AppSecret,
		code,
	)

	resp, err := wechatHTTPClient.Get(wxURL)
	if err != nil {
		return "", fmt.Errorf("微信登录失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("微信登录失败")
	}

	var wxResult map[string]interface{}
	if err := json.Unmarshal(body, &wxResult); err != nil {
		return "", fmt.Errorf("微信登录失败")
	}

	if errcode, ok := wxResult["errcode"]; ok {
		errMsg, _ := wxResult["errmsg"].(string)
		return "", fmt.Errorf("微信登录失败: %v %s", errcode, errMsg)
	}

	openid, ok := wxResult["openid"].(string)
	if !ok || openid == "" {
		return "", fmt.Errorf("微信登录失败，未获取到openid")
	}
	return openid, nil
}

func getWechatPhoneNumber(cfg config.Config, phoneCode string) (string, error) {
	accessToken, err := getWechatStableAccessToken(cfg)
	if err != nil {
		return "", err
	}

	payload, _ := json.Marshal(map[string]string{"code": phoneCode})
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s?access_token=%s", cfg.WechatConfig.PhoneNumberURL, accessToken),
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("微信手机号授权失败")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := wechatHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("微信手机号授权失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("微信手机号授权失败")
	}

	var wxResult struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
		} `json:"phone_info"`
	}
	if err := json.Unmarshal(body, &wxResult); err != nil {
		return "", fmt.Errorf("微信手机号授权失败")
	}
	if wxResult.ErrCode != 0 {
		return "", fmt.Errorf("微信手机号授权失败: %s", wxResult.ErrMsg)
	}
	mobile := wxResult.PhoneInfo.PurePhoneNumber
	if mobile == "" {
		mobile = wxResult.PhoneInfo.PhoneNumber
	}
	if !method.IsValidMobile(mobile) {
		return "", fmt.Errorf("微信手机号格式无效")
	}
	return mobile, nil
}

func getWechatStableAccessToken(cfg config.Config) (string, error) {
	wechatTokenCache.Lock()
	if wechatTokenCache.token != "" && time.Now().Before(wechatTokenCache.expiresAt) {
		token := wechatTokenCache.token
		wechatTokenCache.Unlock()
		return token, nil
	}
	wechatTokenCache.Unlock()

	payload, _ := json.Marshal(map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         cfg.WechatConfig.AppID,
		"secret":        cfg.WechatConfig.AppSecret,
		"force_refresh": false,
	})
	req, err := http.NewRequest(http.MethodPost, cfg.WechatConfig.AccessTokenURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("微信access_token获取失败")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := wechatHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("微信access_token获取失败")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("微信access_token获取失败")
	}

	var wxResult struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &wxResult); err != nil {
		return "", fmt.Errorf("微信access_token获取失败")
	}
	if wxResult.ErrCode != 0 || wxResult.AccessToken == "" {
		return "", fmt.Errorf("微信access_token获取失败: %s", wxResult.ErrMsg)
	}
	wechatTokenCache.Lock()
	wechatTokenCache.token = wxResult.AccessToken
	wechatTokenCache.expiresAt = time.Now().Add(110 * time.Minute)
	wechatTokenCache.Unlock()
	return wxResult.AccessToken, nil
}

func memberLoginErrorMessage(err error) string {
	switch err.Error() {
	case "mobile is not a member":
		return "该手机号不是会员，请联系商家开通会员"
	case "member disabled":
		return "该会员已停用，请联系商家处理"
	case "mobile already linked to another wechat user", "mobile already bound to another openid":
		return "该手机号已绑定其他微信账号"
	case "openid already bound to another mobile":
		return "当前微信账号已绑定其他手机号"
	case "member already linked to another user":
		return "该会员已绑定其他用户ID"
	default:
		return err.Error()
	}
}

// AddData 添加用户数据
func (uc *UserController) AddData(t *gin.Context) {
	var req requestbody.AddUserData
	//校验参数
	if err := t.ShouldBindJSON(&req); err != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 创建用户数据
	userData := models.UserData{
		UserID:     req.UserId,
		DataType:   req.DataType,
		DataValue:  req.DataValue,
		CreateTime: time.Now(),
	}
	ex := method.CreateUserData(&userData)
	if ex != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponse("创建失败", ex))
		return
	}
	info, err := method.SelectUserData(req.UserId)
	if err != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponse("创建失败", err))
		return
	}
	t.JSON(http.StatusOK, msg.SuccessResponse("创建成功", &info))
	return
}

func (uc *UserController) FindData(t *gin.Context) {
	var req requestbody.FindDataRequest
	//校验参数
	if err := t.ShouldBindJSON(&req); err != nil {
		t.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体格式"})
		return
	}
	//校验是否存在
	if ex := method.SearchExistence("users_data", "user_id", req.UserId); ex != true {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("用户不存在"))
		return
	}
	info, err := method.SelectUserData(req.UserId)
	if err != nil {
		t.JSON(http.StatusBadRequest, msg.ErrResponse("查询失败", err))
		return
	}
	t.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
	return
}

func (uc *UserController) SendRegisterCaptcha(c *gin.Context) {
	var req requestbody.SendRegisterCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}
	if !method.IsValidMobile(req.Mobile) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid mobile"))
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

func (uc *UserController) BindWechatPhone(c *gin.Context) {
	var req requestbody.BindWechatPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}

	member, err := method.BindWechatPhone(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"user_id":   member.UserID,
		"member_no": member.MemberNo,
		"openid":    member.OpenID,
		"mobile":    member.Mobile,
		"nickname":  member.Nickname,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("Wechat phone bound successfully", &data))
}

func (uc *UserController) UpdatePlatformInfo(c *gin.Context) {
	var req requestbody.UpdatePlatformInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}

	member, err := method.UpdatePlatformInfo(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"user_id":       member.UserID,
		"member_no":     member.MemberNo,
		"tmall_id":      member.TmallID,
		"tmall_amount":  member.TmallAmount,
		"youzan_id":     member.YouzanID,
		"youzan_amount": member.YouzanAmount,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("Platform info updated successfully", &data))
}

func (uc *UserController) MemberAmountSummary(c *gin.Context) {
	var req requestbody.MemberAmountSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("Invalid request"))
		return
	}

	member, err := method.MemberAmountSummary(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"user_id":            member.UserID,
		"member_no":          member.MemberNo,
		"mobile":             member.Mobile,
		"total_order_amount": member.TotalOrderAmount,
		"total_paid_amount":  member.TotalPaidAmount,
		"tmall_id":           member.TmallID,
		"tmall_amount":       member.TmallAmount,
		"youzan_id":          member.YouzanID,
		"youzan_amount":      member.YouzanAmount,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}
