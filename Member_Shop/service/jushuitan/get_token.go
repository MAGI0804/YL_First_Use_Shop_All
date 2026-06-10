package jushuitan

import (
	"Member_shop/config"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    TokenData `json:"data"`
}

type TokenData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func md5Encrypt(paymentStr string) string {
	h := md5.New()
	h.Write([]byte(paymentStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetTokenTest() (string, error) {
	cfg := config.LoadConfig()
	return requestToken(
		cfg.JushuitanConfig.AppKeyTest,
		cfg.JushuitanConfig.AppSecretTest,
		cfg.JushuitanConfig.AuthCodeTest,
		cfg.JushuitanConfig.GetTokenURLTest,
		"测试",
	)
}

func GetToken() (string, error) {
	cfg := config.LoadConfig()
	if useJushuitanTestEnvironment(cfg) {
		return GetTokenTest()
	}
	return GetTokenProd()
}

func GetTokenProd() (string, error) {
	cfg := config.LoadConfig()
	return requestToken(
		cfg.JushuitanConfig.AppKeyProd,
		cfg.JushuitanConfig.AppSecretProd,
		cfg.JushuitanConfig.AuthCodeProd,
		cfg.JushuitanConfig.GetTokenURLProd,
		"生产",
	)
}

func requestToken(appKey, appSecret, authCode, tokenURL, envName string) (string, error) {
	if appKey == "" || appSecret == "" || authCode == "" {
		return "", fmt.Errorf("聚水潭%s应用配置未完整设置", envName)
	}
	tokenURL = strings.TrimSpace(tokenURL)
	if tokenURL == "" {
		return "", fmt.Errorf("聚水潭%s应用token刷新URL未配置", envName)
	}
	timestamp := int(time.Now().Unix())
	charset := "UTF-8"
	grantType := "authorization_code"

	convertedStr := fmt.Sprintf("%sapp_key%scharset%scode%sgrant_type%stimestamp%d",
		appSecret, appKey, charset, authCode, grantType, timestamp)
	sign := md5Encrypt(convertedStr)

	data := url.Values{}
	data.Set("app_key", appKey)
	data.Set("grant_type", grantType)
	data.Set("timestamp", fmt.Sprintf("%d", timestamp))
	data.Set("code", authCode)
	data.Set("charset", charset)
	data.Set("sign", sign)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("GetToken%s response: %s\n", envName, string(body))

	return parseTokenFromResponse(body)
}

func parseTokenFromResponse(body []byte) (string, error) {
	var resp TokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("获取token失败: %s", resp.Message)
	}

	return resp.Data.AccessToken, nil
}
