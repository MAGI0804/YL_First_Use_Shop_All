package jushuitan

import (
	"Member_shop/config"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	if cfg.JushuitanConfig.AccessTokenTest == "" {
		return "", fmt.Errorf("JST_ACCESS_TOKEN_TEST未配置")
	}
	return cfg.JushuitanConfig.AccessTokenTest, nil
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
	if cfg.JushuitanConfig.AppKeyProd == "" || cfg.JushuitanConfig.AppSecretProd == "" || cfg.JushuitanConfig.AuthCodeProd == "" {
		return "", fmt.Errorf("聚水潭生产应用配置未完整设置")
	}

	timestamp := int(time.Now().Unix())
	charset := "UTF-8"
	grantType := "authorization_code"
	code := cfg.JushuitanConfig.AuthCodeProd

	convertedStr := fmt.Sprintf("%sapp_key%scharset%scode%sgrant_type%stimestamp%d",
		cfg.JushuitanConfig.AppSecretProd, cfg.JushuitanConfig.AppKeyProd, charset, code, grantType, timestamp)
	sign := md5Encrypt(convertedStr)

	url := cfg.JushuitanConfig.GetTokenURLProd
	if strings.TrimSpace(url) == "" {
		return "", fmt.Errorf("JST_GET_TOKEN_URL_PROD未配置")
	}
	data := fmt.Sprintf("app_key=%s&grant_type=%s&timestamp=%d&code=%s&charset=%s&sign=%s",
		cfg.JushuitanConfig.AppKeyProd, grantType, timestamp, code, charset, sign)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("GetTokenProd response: %s\n", string(body))

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
