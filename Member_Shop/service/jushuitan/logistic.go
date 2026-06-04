package jushuitan

import (
	"Member_shop/config"
	"encoding/json"
	"fmt"
)

// LogisticQueryRequest is the biz payload for /open/logistic/query.
type LogisticQueryRequest struct {
	ShopID        int      `json:"shop_id,omitempty"`
	PageIndex     int      `json:"page_index"`
	PageSize      int      `json:"page_size"`
	ModifiedBegin string   `json:"modified_begin,omitempty"`
	ModifiedEnd   string   `json:"modified_end,omitempty"`
	DateType      int      `json:"date_type,omitempty"`
	SoIDs         []string `json:"so_ids,omitempty"`
}

// LogisticQueryResponse is the parsed response from Jushuitan logistic query.
type LogisticQueryResponse struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data LogisticQueryData `json:"data"`
}

type LogisticQueryData struct {
	PageIndex int             `json:"page_index"`
	PageSize  int             `json:"page_size"`
	DataCount int             `json:"data_count"`
	PageCount int             `json:"page_count"`
	HasNext   bool            `json:"has_next"`
	Orders    []LogisticOrder `json:"orders"`
}

type LogisticOrder struct {
	OID              int            `json:"o_id"`
	ShopID           int            `json:"shop_id"`
	SoID             string         `json:"so_id"`
	ASID             int            `json:"as_id"`
	SendDate         string         `json:"send_date"`
	Freight          float64        `json:"freight"`
	Weight           float64        `json:"weight"`
	WmsCoID          int            `json:"wms_co_id"`
	LCID             string         `json:"lc_id"`
	LID              string         `json:"l_id"`
	LogisticsCompany string         `json:"logistics_company"`
	Items            []LogisticItem `json:"items"`
}

type LogisticItem struct {
	OID          string `json:"o_id"`
	SkuID        string `json:"sku_id"`
	Qty          int    `json:"qty"`
	OuterOiID    string `json:"outer_oi_id"`
	RawSoID      string `json:"raw_so_id"`
	RefundStatus string `json:"refund_status"`
	SkuType      string `json:"sku_type"`
}

// QueryLogistic queries Jushuitan shipment information and returns both parsed and raw responses.
func QueryLogistic(accessToken string, query LogisticQueryRequest) (*LogisticQueryResponse, string, error) {
	if query.PageIndex <= 0 {
		query.PageIndex = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 30
	}
	if query.PageSize > 50 {
		query.PageSize = 50
	}

	cfg := config.LoadConfig()
	body, err := postOpenAPI(accessToken, cfg.JushuitanConfig.LogisticQueryURLTest, query)
	if err != nil {
		return nil, "", err
	}

	var resp LogisticQueryResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, body, fmt.Errorf("解析发货信息查询响应失败: %v", err)
	}
	if resp.Code != 0 {
		return nil, body, fmt.Errorf("查询发货信息失败: %s", resp.Msg)
	}
	return &resp, body, nil
}
