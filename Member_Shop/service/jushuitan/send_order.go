package jushuitan

import (
	"Member_shop/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ShopQueryRequest 店铺查询请求
// 用于向聚水潭系统发送店铺查询请求的参数封装
type ShopQueryRequest struct {
	PageIndex int   `json:"page_index"` // 页码
	PageSize  int   `json:"page_size"`  // 每页数量
	ShopIDs   []int `json:"shop_ids"`   // 店铺ID列表
}

// ShopQueryResponse 店铺查询响应
// 聚水潭系统返回的店铺查询结果封装
type ShopQueryResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Data    ShopQueryData `json:"data"`
}

// ShopQueryData 店铺查询响应数据
// 包含分页信息和店铺列表
type ShopQueryData struct {
	PageIndex  int        `json:"page_index"`  // 页码
	PageSize   int        `json:"page_size"`   // 每页数量
	TotalCount int        `json:"total_count"` // 总数
	TotalPages int        `json:"total_pages"` // 总页数
	Shops      []ShopInfo `json:"shops"`       // 店铺列表
}

// ShopInfo 店铺信息
// 包含店铺的基本信息
type ShopInfo struct {
	ShopID   int    `json:"shop_id"`   // 店铺ID
	ShopName string `json:"shop_name"` // 店铺名称
	Status   string `json:"status"`    // 状态
}

const Version = 2 // API版本号

// QueryShops 查询店铺信息
// accessToken: 访问令牌
// shopIDs: 店铺ID列表
// pageIndex: 页码
// pageSize: 每页数量
// 返回JSON字符串和错误信息
func QueryShops(accessToken string, shopIDs []int, pageIndex, pageSize int) (string, error) {
	cfg := config.LoadConfig()
	if cfg.JushuitanConfig.AppKeyTest == "" || cfg.JushuitanConfig.AppSecretTest == "" {
		return "", fmt.Errorf("聚水潭测试应用配置未完整设置")
	}

	appKey := cfg.JushuitanConfig.AppKeyTest
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	charset := "UTF-8"

	biz, err := json.Marshal(ShopQueryRequest{
		PageIndex: pageIndex,
		PageSize:  pageSize,
		ShopIDs:   shopIDs,
	})
	if err != nil {
		return "", fmt.Errorf("序列化biz参数失败: %v", err)
	}

	convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
		cfg.JushuitanConfig.AppSecretTest, accessToken, appKey, string(biz), charset, timestamp, Version)
	sign := md5Encrypt(convertedStr)

	apiURL := strings.TrimSpace(cfg.JushuitanConfig.ShopQueryURLTest)
	if apiURL == "" {
		return "", fmt.Errorf("JST_SHOP_QUERY_URL_TEST未配置")
	}
	data := fmt.Sprintf("app_key=%s&access_token=%s&timestamp=%s&charset=%s&version=%d&sign=%s&biz=%s",
		appKey, accessToken, timestamp, charset, Version, sign, string(biz))

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("QueryShops response: %s\n", string(body))

	return string(body), nil
}

// OrderItem 订单商品明细
// 包含订单中每件商品的详细信息
type OrderItem struct {
	SkuID        string  `json:"sku_id"`        // 商品SKU编码
	ShopSkuID    string  `json:"shop_sku_id"`   // 店铺商品编码
	Amount       float64 `json:"amount"`        // 商品成交总金额
	BasePrice    float64 `json:"base_price"`    // 商品原价
	Qty          int     `json:"qty"`           // 商品购买数量
	Name         string  `json:"name"`          // 商品名称
	OuterOiID    string  `json:"outer_oi_id"`   // 订单商品明细主键
	BatchID      string  `json:"batch_id"`      // 商品批次号
	ProducedDate string  `json:"produced_date"` // 商品生产日期
}

// PayInfo 支付信息
// 包含订单的支付相关详细信息
type PayInfo struct {
	OuterPayID    string  `json:"outer_pay_id"`   // 外部支付流水号
	PayDate       string  `json:"pay_date"`       // 支付时间
	Payment       string  `json:"payment"`        // 支付方式
	SellerAccount string  `json:"seller_account"` // 卖家收款账号
	BuyerAccount  string  `json:"buyer_account"`  // 买家付款账号
	Amount        float64 `json:"amount"`         // 支付金额
}

// OrderData 订单完整数据
// 包含订单的所有信息，用于上传到聚水潭系统
type OrderData struct {
	ShopID           int         `json:"shop_id"`           // 店铺ID
	SoID             string      `json:"so_id"`             // 外部订单号
	OrderDate        string      `json:"order_date"`        // 订单创建时间
	ShopStatus       string      `json:"shop_status"`       // 订单状态
	ShopBuyerID      string      `json:"shop_buyer_id"`     // 买家账号
	ReceiverState    string      `json:"receiver_state"`    // 收货省份
	ReceiverCity     string      `json:"receiver_city"`     // 收货城市
	ReceiverDistrict string      `json:"receiver_district"` // 收货区县
	ReceiverAddress  string      `json:"receiver_address"`  // 详细收货地址
	ReceiverName     string      `json:"receiver_name"`     // 收货人姓名
	ReceiverPhone    string      `json:"receiver_phone"`    // 收货人手机号
	ReceiverZip      string      `json:"receiver_zip"`      // 收货邮编
	PayAmount        float64     `json:"pay_amount"`        // 订单应付金额
	Freight          float64     `json:"freight"`           // 运费金额
	Remark           string      `json:"remark"`            // 订单备注
	BuyerMessage     string      `json:"buyer_message"`     // 买家留言
	ShopModified     string      `json:"shop_modified"`     // 订单修改时间
	Items            []OrderItem `json:"items"`             // 商品列表
	Pay              PayInfo     `json:"pay"`               // 支付信息
}

// OrderUploadRequest 订单上传请求
// 订单数据列表，用于批量上传订单到聚水潭
type OrderUploadRequest []OrderData

// OrderUploadResponse 订单上传响应
// 聚水潭系统返回的上传结果
type OrderUploadResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
}

// SendOrder 上传订单到聚水潭
// accessToken: 访问令牌
// orderData: 订单数据
// 返回JSON字符串和错误信息
func SendOrder(accessToken string, orderData OrderData) (string, error) {
	cfg := config.LoadConfig()
	if cfg.JushuitanConfig.AppKeyTest == "" || cfg.JushuitanConfig.AppSecretTest == "" {
		return "", fmt.Errorf("聚水潭测试应用配置未完整设置")
	}

	appKey := cfg.JushuitanConfig.AppKeyTest
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	charset := "UTF-8"

	biz, err := json.Marshal([]OrderData{orderData})
	if err != nil {
		return "", fmt.Errorf("序列化biz参数失败: %v", err)
	}

	convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
		cfg.JushuitanConfig.AppSecretTest, accessToken, appKey, string(biz), charset, timestamp, Version)
	sign := md5Encrypt(convertedStr)

	apiURL := strings.TrimSpace(cfg.JushuitanConfig.OrderUploadURLTest)
	if apiURL == "" {
		return "", fmt.Errorf("JST_ORDER_UPLOAD_URL_TEST未配置")
	}
	data := fmt.Sprintf("app_key=%s&access_token=%s&timestamp=%s&charset=%s&version=%d&sign=%s&biz=%s",
		appKey, accessToken, timestamp, charset, Version, sign, string(biz))

	log.Printf("聚水潭订单上传请求已生成: so_id=%s item_count=%d", orderData.SoID, len(orderData.Items))

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("SendOrder response: %s\n", string(body))

	var uploadResp OrderUploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if uploadResp.Code != 0 {
		return "", fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	return string(body), nil
}
