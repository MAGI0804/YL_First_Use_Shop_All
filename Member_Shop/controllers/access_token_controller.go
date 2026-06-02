package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/msg"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"Member_shop/config"
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm/clause"
)

// AccessTokenController 访问令牌控制器
type AccessTokenController struct{}

// TokenRefresh 刷新JWT令牌 - 对应Django的TokenRefreshView
func (ac *AccessTokenController) TokenRefresh(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON格式"})
		return
	}

	refreshToken, ok := requestData["refresh"].(string)
	if !ok || refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少refresh参数"})
		return
	}

	// 获取配置
	cfg := config.LoadConfig()

	// 解析并验证刷新令牌
	newAccessToken, err := utils.RefreshAccessToken(refreshToken, cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的刷新令牌"})
		return
	}

	// 返回与Django相同格式的响应
	c.JSON(http.StatusOK, gin.H{
		"access": newAccessToken,
	})
}

// GetToken 获取访问令牌
func (ac *AccessTokenController) GetToken(c *gin.Context) {
	// 获取客户端IP地址，优先使用X-Forwarded-For
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	var ipAddress string

	if xForwardedFor != "" {
		// 提取第一个IP地址
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ipAddress = strings.TrimSpace(ips[0])
		} else {
			ipAddress = c.ClientIP()
		}
	} else {
		ipAddress = c.ClientIP()
	}

	// 确保IP不为空
	if ipAddress == "" {
		ipAddress = "unknown"
	}

	// 检查该IP是否已存在token
	info := map[string]any{}
	RegisterTime := time.Now()
	
	// 从access_token hash中获取该IP对应的token
	token, err := db.GetTokenRedis(ipAddress)
	if err == nil && token != "" {
		// 如果token存在，直接返回
		info["access_token"] = token
		info["RegisterTime"] = RegisterTime
		info["ip_addresses"] = ipAddress
		c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
		return
	} else if err != nil && err != redis.Nil {
		// 如果发生非空错误，返回错误
		c.JSON(http.StatusBadRequest, msg.ErrResponse("申请失败", err))
		return
	}

	// 生成32位随机不重复token
	var accessToken string
	for {
		// 生成16字节随机数据，转为32位十六进制字符串
		randomBytes := make([]byte, 16)
		rand.Read(randomBytes)
		accessToken = hex.EncodeToString(randomBytes)

		// 检查token是否已存在
		var tokenObj models.AccessToken
		if err := db.DB.Where("access_token = ?", accessToken).First(&tokenObj).Error; err != nil {
			// 不存在则使用此token
			break
		}
	}

	if err := db.SaveToken(accessToken, ipAddress); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("申请失败", err))
		return
	}

	// 创建新token记录
	tokenRecord := models.AccessToken{
		IPAddress:    ipAddress, // 唯一键，用于判断是否冲突
		AccessToken:  accessToken,
		RegisterTime: RegisterTime,
	}

	eerr := db.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ip_address"}}, // 冲突的唯一键字段
		DoUpdates: clause.Assignments(map[string]interface{}{
			"access_token":  accessToken,  // 冲突时更新token
			"register_time": RegisterTime, // 冲突时更新注册时间
		}),
	}).Create(&tokenRecord).Error

	if eerr != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("保存Token记录失败", eerr))
		return
	}
	info["access_token"] = accessToken
	info["RegisterTime"] = RegisterTime
	info["ip_addresses"] = ipAddress
	c.JSON(http.StatusCreated, msg.SuccessResponse("申请成功", &info))
	return
}

// GetIPs 获取所有唯一IP地址
func (ac *AccessTokenController) GetIPs(c *gin.Context) {
	var requestData requestbody.RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid JSON format",
		})
		return
	}

	// 验证shopname参数
	if requestData.ShopName != "youlan_kids" {
		c.JSON(http.StatusForbidden, msg.ErrResponseStr("参数不正确"))
		return
	}

	// 获取所有唯一IP地址
	var ipAddresses []string
	if err := db.DB.Model(&models.AccessToken{}).Distinct().Pluck("ip_address", &ipAddresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("查询错误", err))
		return
	}
	info := map[string]any{
		"ip_addresses": ipAddresses,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
}
