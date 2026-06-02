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

// InventoryQueryRequest 库存查询请求结构
// 用于向聚水潭系统发送库存查询请求的参数封装
type InventoryQueryRequest struct {
	PageIndex int    `json:"page_index"` // 页码，从1开始
	PageSize  int    `json:"page_size"`  // 每页数量，最大支持100
	SkuIDs    string `json:"sku_ids"`    // 商品SKU编码列表，多个用逗号分隔
}

// InventoryQueryResponse 库存查询响应结构
// 聚水潭系统返回的库存查询结果封装
type InventoryQueryResponse struct {
	Code    int           `json:"code"`    // 响应码，0表示成功
	Message string        `json:"message"` // 响应消息
	Data    InventoryData `json:"data"`    // 响应数据
}

// InventoryData 库存查询响应数据
// 包含分页信息和库存明细列表
type InventoryData struct {
	PageIndex  int             `json:"page_index"` // 当前页码
	PageSize   int             `json:"page_size"`  // 每页数量
	DataCount  int             `json:"data_count"` // 总数据条数
	PageCount  int             `json:"page_count"` // 总页数
	HasNext    bool            `json:"has_next"`   // 是否有下一页
	Inventorys []InventoryItem `json:"inventorys"` // 库存明细列表
}

// InventoryItem 库存明细项
// 每件商品的具体库存信息
type InventoryItem struct {
	IID          string `json:"i_id"`          // 款式编码
	SkuID        string `json:"sku_id"`        // 商品SKU编码
	Name         string `json:"name"`          // 商品名称
	Qty          int    `json:"qty"`           // 主仓实际库存
	VirtualQty   int    `json:"virtual_qty"`   // 虚拟库存数量
	OrderLock    int    `json:"order_lock"`    // 订单占用数量
	PickLock     int    `json:"pick_lock"`     // 订单待发数量
	DefectiveQty int    `json:"defective_qty"` // 次品数量
	ReturnQty    int    `json:"return_qty"`    // 退货数量
	Modified     string `json:"modified"`      // 最后修改时间
}

// QueryInventory 查询商品库存
// accessToken: 聚水潭API访问令牌
// skuIDs: 要查询的商品SKU编码列表
// pageIndex: 查询的页码，从1开始
// pageSize: 每页返回的数量，最大100
// 返回库存查询响应和错误信息
func QueryInventory(accessToken string, skuIDs []string, pageIndex, pageSize int) (*InventoryQueryResponse, error) {
	cfg := config.LoadConfig()
	if cfg.JushuitanConfig.AppKeyProd == "" || cfg.JushuitanConfig.AppSecretProd == "" {
		return nil, fmt.Errorf("聚水潭生产应用配置未完整设置")
	}

	// 构建请求时间戳
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	charset := "UTF-8"

	// 序列化业务参数
	biz, err := json.Marshal(InventoryQueryRequest{
		PageIndex: pageIndex,
		PageSize:  pageSize,
		SkuIDs:    strings.Join(skuIDs, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("序列化biz参数失败: %v", err)
	}

	// 构建签名参数并计算签名
	convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
		cfg.JushuitanConfig.AppSecretProd, accessToken, cfg.JushuitanConfig.AppKeyProd, string(biz), charset, timestamp, Version)
	sign := md5Encrypt(convertedStr)

	// 使用生产环境接口
	url := strings.TrimRight(cfg.JushuitanConfig.OpenAPIURLProd, "/") + "/open/inventory/query"

	// 构建POST表单数据
	data := fmt.Sprintf("app_key=%s&access_token=%s&timestamp=%s&charset=%s&version=%d&sign=%s&biz=%s",
		cfg.JushuitanConfig.AppKeyProd, accessToken, timestamp, charset, Version, sign, string(biz))

	log.Printf("库存查询请求已生成: sku_count=%d page=%d page_size=%d", len(skuIDs), pageIndex, pageSize)

	// 发送HTTP请求
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("库存查询响应: %s", string(body))

	// 解析响应JSON
	var invResp InventoryQueryResponse
	if err := json.Unmarshal(body, &invResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应码
	if invResp.Code != 0 {
		return nil, fmt.Errorf("查询失败: %s", invResp.Message)
	}

	return &invResp, nil
}
