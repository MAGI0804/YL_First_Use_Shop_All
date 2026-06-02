package requestbody

// AnalyticsFilterRequest 是数据分析接口的统一筛选条件。
// begin_time/end_time 支持 "2006-01-02" 和 "2006-01-02 15:04:05" 两种常用格式。
type AnalyticsFilterRequest struct {
	BeginTime             string `json:"begin_time" form:"begin_time"`                           // 统计开始时间，为空表示不限制开始时间
	EndTime               string `json:"end_time" form:"end_time"`                               // 统计结束时间，为空表示不限制结束时间
	Shopname              string `json:"shopname" form:"shopname"`                               // 店铺或渠道名称，当前按订单来源 order_from 预留筛选
	Category              string `json:"category" form:"category"`                               // 商品分类，用于商品和用户偏好统计
	StyleCode             string `json:"style_code" form:"style_code"`                           // 款号，用于商品和用户偏好统计
	OperatorID            string `json:"operator_id" form:"operator_id"`                         // 后台操作人，当前主要给后续运营报表预留
	LowInventoryThreshold int    `json:"low_inventory_threshold" form:"low_inventory_threshold"` // 低库存阈值，默认 5
	SlowSalesThreshold    int    `json:"slow_sales_threshold" form:"slow_sales_threshold"`       // 滞销销量阈值，默认 0
	Limit                 int    `json:"limit" form:"limit"`                                     // 榜单返回数量，默认 20，最大 100
}
