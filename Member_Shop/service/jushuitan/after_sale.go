package jushuitan

import (
	"Member_shop/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AfterSaleItem 售后商品明细。
// 字段按聚水潭售后上传接口的外部单号、商品编码、数量等核心字段组织。
type AfterSaleItem struct {
	OuterOiID string  `json:"outer_oi_id,omitempty"`
	SkuID     string  `json:"sku_id,omitempty"`
	ShopSkuID string  `json:"shop_sku_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	Qty       int     `json:"qty"`
	Amount    float64 `json:"amount,omitempty"`
	Type      string  `json:"type,omitempty"`
}

// AfterSaleData 售后上传/修改数据。
type AfterSaleData struct {
	ShopID           int             `json:"shop_id"`
	OuterASID        string          `json:"outer_as_id"`
	SoID             string          `json:"so_id"`
	Type             string          `json:"type"`
	ShopStatus       string          `json:"shop_status,omitempty"`
	QuestionType     string          `json:"question_type,omitempty"`
	Reason           string          `json:"reason,omitempty"`
	Remark           string          `json:"remark,omitempty"`
	Created          string          `json:"created,omitempty"`
	Modified         string          `json:"modified,omitempty"`
	BuyerAccount     string          `json:"buyer_account,omitempty"`
	ReceiverState    string          `json:"receiver_state,omitempty"`
	ReceiverCity     string          `json:"receiver_city,omitempty"`
	ReceiverDistrict string          `json:"receiver_district,omitempty"`
	ReceiverAddress  string          `json:"receiver_address,omitempty"`
	ReceiverPhone    string          `json:"receiver_phone,omitempty"`
	Items            []AfterSaleItem `json:"items"`
}

type AfterSaleReceivedQuery struct {
	PageIndex     int    `json:"page_index"`
	PageSize      int    `json:"page_size"`
	ModifiedBegin string `json:"modified_begin,omitempty"`
	ModifiedEnd   string `json:"modified_end,omitempty"`
	SoID          string `json:"so_id,omitempty"`
	OuterASID     string `json:"outer_as_id,omitempty"`
	ASID          string `json:"as_id,omitempty"`
}

type apiResponse struct {
	Code    interface{}     `json:"code"`
	Message string          `json:"message"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

func SendAfterSale(accessToken string, data AfterSaleData) (string, error) {
	cfg := config.LoadConfig()
	apiURL, err := activeURL(
		cfg,
		cfg.JushuitanConfig.AfterSaleUploadURLTest,
		cfg.JushuitanConfig.AfterSaleUploadURLProd,
		"JST_AFTERSALE_UPLOAD_URL_TEST",
		"JST_AFTERSALE_UPLOAD_URL_PROD",
	)
	if err != nil {
		return "", err
	}

	body, err := postOpenAPI(accessToken, apiURL, data)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if !apiCodeOK(resp.Code) {
		return "", fmt.Errorf("上传售后失败: %s", apiResponseMessage(resp))
	}
	return body, nil
}

func QueryAfterSaleReceived(accessToken string, query AfterSaleReceivedQuery) (string, error) {
	if query.PageIndex <= 0 {
		query.PageIndex = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	cfg := config.LoadConfig()
	apiURL, err := activeURL(
		cfg,
		cfg.JushuitanConfig.AfterSaleReceivedQueryURLTest,
		cfg.JushuitanConfig.AfterSaleReceivedQueryURLProd,
		"JST_AFTERSALE_RECEIVED_QUERY_URL_TEST",
		"JST_AFTERSALE_RECEIVED_QUERY_URL_PROD",
	)
	if err != nil {
		return "", err
	}

	body, err := postOpenAPI(accessToken, apiURL, query)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if !apiCodeOK(resp.Code) {
		return "", fmt.Errorf("查询实际收货失败: %s", apiResponseMessage(resp))
	}
	return body, nil
}

func postOpenAPI(accessToken, apiURL string, bizValue interface{}) (string, error) {
	cfg := config.LoadConfig()
	apiEnv, err := activeOpenAPIEnvironment(cfg)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(accessToken) == "" {
		return "", fmt.Errorf("聚水潭access_token不能为空")
	}
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		return "", fmt.Errorf("聚水潭接口URL未配置")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	charset := "UTF-8"
	biz, err := json.Marshal(bizValue)
	if err != nil {
		return "", fmt.Errorf("序列化biz参数失败: %v", err)
	}

	convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
		apiEnv.AppSecret, accessToken, apiEnv.AppKey, string(biz), charset, timestamp, Version)
	sign := md5Encrypt(convertedStr)

	form := url.Values{}
	form.Set("app_key", apiEnv.AppKey)
	form.Set("access_token", accessToken)
	form.Set("timestamp", timestamp)
	form.Set("charset", charset)
	form.Set("version", strconv.Itoa(Version))
	form.Set("sign", sign)
	form.Set("biz", string(biz))

	log.Printf("聚水潭请求已生成: api_url=%s biz=%s", apiURL, string(biz))

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func apiCodeOK(code interface{}) bool {
	switch value := code.(type) {
	case float64:
		return value == 0
	case string:
		return value == "0" || strings.EqualFold(value, "success")
	case nil:
		return false
	default:
		return false
	}
}

func apiResponseMessage(resp apiResponse) string {
	if resp.Message != "" {
		return resp.Message
	}
	if resp.Msg != "" {
		return resp.Msg
	}
	return "未知错误"
}
