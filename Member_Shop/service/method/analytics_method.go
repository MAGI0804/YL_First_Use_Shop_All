package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	defaultAnalyticsLimit        = 20
	maxAnalyticsLimit            = 100
	defaultLowInventoryThreshold = 5
	defaultSlowSalesThreshold    = 0
)

// AnalyticsFilter 是数据分析模块内部统一使用的筛选条件。
// Controller 保持轻量，只把请求体转换成这个结构，所有统计口径都集中在 service 层。
type AnalyticsFilter struct {
	BeginTime             string // 统计开始时间
	EndTime               string // 统计结束时间
	Shopname              string // 店铺或渠道，当前映射到订单来源 order_from
	Category              string // 商品分类
	StyleCode             string // 款号
	OperatorID            string // 后台操作人，当前预留给后续运营报表
	LowInventoryThreshold int    // 低库存阈值
	SlowSalesThreshold    int    // 滞销销量阈值
	Limit                 int    // 榜单数量限制
}

// SalesSummaryResult 是销售统计接口的返回结构。
type SalesSummaryResult struct {
	OrderCount          int64             `json:"order_count"`           // 订单总数
	PaidOrderCount      int64             `json:"paid_order_count"`      // 已支付订单数
	CanceledOrderCount  int64             `json:"canceled_order_count"`  // 已取消订单数
	SalesAmount         float64           `json:"sales_amount"`          // 实际销售额，按已支付订单 final_pay_amount 汇总
	PaidAmount          float64           `json:"paid_amount"`           // 已支付订单最终实付金额汇总
	OriginalOrderAmount float64           `json:"original_order_amount"` // 原订单金额汇总
	DiscountAmount      float64           `json:"discount_amount"`       // 优惠金额汇总
	RefundAmount        float64           `json:"refund_amount"`         // 已完成退款售后金额
	AverageOrderValue   float64           `json:"average_order_value"`   // 客单价
	Daily               []SalesDailyPoint `json:"daily"`                 // 按天统计明细
}

// SalesDailyPoint 表示某一天的销售统计。
type SalesDailyPoint struct {
	Date                string  `json:"date" gorm:"column:date"`
	OrderCount          int64   `json:"order_count" gorm:"column:order_count"`
	PaidOrderCount      int64   `json:"paid_order_count" gorm:"column:paid_order_count"`
	CanceledOrderCount  int64   `json:"canceled_order_count" gorm:"column:canceled_order_count"`
	SalesAmount         float64 `json:"sales_amount" gorm:"column:sales_amount"`
	PaidAmount          float64 `json:"paid_amount" gorm:"column:paid_amount"`
	OriginalOrderAmount float64 `json:"original_order_amount" gorm:"column:original_order_amount"`
	DiscountAmount      float64 `json:"discount_amount" gorm:"column:discount_amount"`
	RefundAmount        float64 `json:"refund_amount" gorm:"column:refund_amount"`
	AverageOrderValue   float64 `json:"average_order_value" gorm:"column:average_order_value"`
}

// UserSummaryResult 是用户分析接口的返回结构。
type UserSummaryResult struct {
	NewUserCount        int64                 `json:"new_user_count"`        // 新增用户数
	NewMemberCount      int64                 `json:"new_member_count"`      // 新增会员数
	OrderUserCount      int64                 `json:"order_user_count"`      // 下单用户数
	PaidUserCount       int64                 `json:"paid_user_count"`       // 支付用户数
	RepurchaseUserCount int64                 `json:"repurchase_user_count"` // 复购用户数
	CategoryPreferences []UserPreferencePoint `json:"category_preferences"`  // 用户购买分类偏好
	StylePreferences    []UserPreferencePoint `json:"style_preferences"`     // 用户购买款号偏好
}

// UserPreferencePoint 表示用户购买偏好的一个统计项。
type UserPreferencePoint struct {
	Name      string `json:"name" gorm:"column:name"`             // 分类名或款号
	UserCount int64  `json:"user_count" gorm:"column:user_count"` // 购买用户数
	SalesQty  int64  `json:"sales_qty" gorm:"column:sales_qty"`   // 购买件数
}

// ProductSummaryResult 是商品分析接口的返回结构。
type ProductSummaryResult struct {
	HotSKUs               []ProductSalesPoint `json:"hot_skus"`                // 热销 SKU
	HotStyleCodes         []StyleSalesPoint   `json:"hot_style_codes"`         // 热销款号
	SlowMovingProducts    []ProductSalesPoint `json:"slow_moving_products"`    // 滞销商品
	InventoryTurnoverRate float64             `json:"inventory_turnover_rate"` // 库存周转率：销量 / 当前库存
	LowInventoryCount     int64               `json:"low_inventory_count"`     // 低库存商品数
	AverageRating         float64             `json:"average_rating"`          // 评价平均分
	GoodRate              float64             `json:"good_rate"`               // 好评率
}

// ProductSalesPoint 表示 SKU 维度销量统计。
type ProductSalesPoint struct {
	CommodityID string  `json:"commodity_id" gorm:"column:commodity_id"`
	Name        string  `json:"name" gorm:"column:name"`
	StyleCode   string  `json:"style_code" gorm:"column:style_code"`
	Category    string  `json:"category" gorm:"column:category"`
	SalesQty    int64   `json:"sales_qty" gorm:"column:sales_qty"`
	SalesAmount float64 `json:"sales_amount" gorm:"column:sales_amount"`
	Inventory   int     `json:"inventory" gorm:"column:inventory"`
}

// StyleSalesPoint 表示款号维度销量统计。
type StyleSalesPoint struct {
	StyleCode   string  `json:"style_code" gorm:"column:style_code"`
	SalesQty    int64   `json:"sales_qty" gorm:"column:sales_qty"`
	SalesAmount float64 `json:"sales_amount" gorm:"column:sales_amount"`
	Inventory   int64   `json:"inventory" gorm:"column:inventory"`
}

// TrafficSummaryResult 是流量分析预留接口的返回结构。
type TrafficSummaryResult struct {
	Status  string           `json:"status"`  // not_implemented 表示当前只是预留接口
	Message string           `json:"message"` // 给调用方的明确说明
	Data    []map[string]any `json:"data"`    // 预留数据列表，当前固定为空数组
}

type analyticsScope struct {
	BeginTime             *time.Time
	EndTime               *time.Time
	Shopname              string
	Category              string
	StyleCode             string
	OperatorID            string
	LowInventoryThreshold int
	SlowSalesThreshold    int
	Limit                 int
}

// SalesSummary 汇总订单金额、支付金额、退款金额和日维度数据。
func SalesSummary(filter AnalyticsFilter) (*SalesSummaryResult, error) {
	scope, err := normalizeAnalyticsFilter(filter)
	if err != nil {
		return nil, err
	}

	result := &SalesSummaryResult{Daily: []SalesDailyPoint{}}
	if err := orderQuery(scope).Count(&result.OrderCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(scope).Where("pay_status = ?", "paid").Count(&result.PaidOrderCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(scope).Where("status IN ?", []string{"canceled", "cancelled"}).
		Count(&result.CanceledOrderCount).Error; err != nil {
		return nil, err
	}

	result.OriginalOrderAmount, err = scanFloat(orderQuery(scope).Select("COALESCE(SUM(order_amount), 0)"))
	if err != nil {
		return nil, err
	}
	result.DiscountAmount, err = scanFloat(orderQuery(scope).
		Where("pay_status = ?", "paid").
		Select("COALESCE(SUM(discount_amount), 0)"))
	if err != nil {
		return nil, err
	}
	result.PaidAmount, err = scanFloat(orderQuery(scope).
		Where("pay_status = ?", "paid").
		Select("COALESCE(SUM(final_pay_amount), 0)"))
	if err != nil {
		return nil, err
	}
	result.SalesAmount = result.PaidAmount
	result.RefundAmount, err = refundAmount(scope)
	if err != nil {
		return nil, err
	}
	if result.PaidOrderCount > 0 {
		result.AverageOrderValue = roundFloat(result.PaidAmount/float64(result.PaidOrderCount), 2)
	}

	if err := loadSalesDaily(scope, result); err != nil {
		return nil, err
	}
	roundSalesSummary(result)
	return result, nil
}

// UserSummary 汇总用户增长、下单用户、支付用户、复购用户和购买偏好。
func UserSummary(filter AnalyticsFilter) (*UserSummaryResult, error) {
	scope, err := normalizeAnalyticsFilter(filter)
	if err != nil {
		return nil, err
	}

	result := &UserSummaryResult{
		CategoryPreferences: []UserPreferencePoint{},
		StylePreferences:    []UserPreferencePoint{},
	}
	if err := applyTimeRange(db.DB.Model(&models.User{}), "registration_date", scope).
		Count(&result.NewUserCount).Error; err != nil {
		return nil, err
	}
	if err := applyTimeRange(db.DB.Model(&models.Member{}), "created_at", scope).
		Count(&result.NewMemberCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(scope).Where("user_id > 0").Distinct("user_id").Count(&result.OrderUserCount).Error; err != nil {
		return nil, err
	}
	if err := orderQuery(scope).Where("pay_status = ?", "paid").Where("user_id > 0").Distinct("user_id").
		Count(&result.PaidUserCount).Error; err != nil {
		return nil, err
	}
	if err := repurchaseUserQuery(scope).Count(&result.RepurchaseUserCount).Error; err != nil {
		return nil, err
	}

	if err := userPreferenceQuery(scope, "c.category").
		Select("c.category AS name, COUNT(DISTINCT o.user_id) AS user_count, COALESCE(SUM(s.qty), 0) AS sales_qty").
		Group("c.category").
		Order("sales_qty DESC, user_count DESC").
		Limit(scope.Limit).
		Scan(&result.CategoryPreferences).Error; err != nil {
		return nil, err
	}
	if err := userPreferenceQuery(scope, "c.style_code").
		Select("c.style_code AS name, COUNT(DISTINCT o.user_id) AS user_count, COALESCE(SUM(s.qty), 0) AS sales_qty").
		Group("c.style_code").
		Order("sales_qty DESC, user_count DESC").
		Limit(scope.Limit).
		Scan(&result.StylePreferences).Error; err != nil {
		return nil, err
	}
	return result, nil
}

// ProductSummary 汇总商品销量、库存周转、低库存和评价数据。
func ProductSummary(filter AnalyticsFilter) (*ProductSummaryResult, error) {
	scope, err := normalizeAnalyticsFilter(filter)
	if err != nil {
		return nil, err
	}

	result := &ProductSummaryResult{
		HotSKUs:            []ProductSalesPoint{},
		HotStyleCodes:      []StyleSalesPoint{},
		SlowMovingProducts: []ProductSalesPoint{},
	}
	if err := productSalesQuery(scope).
		Select(`
			s.commodity_id,
			c.name,
			c.style_code,
			c.category,
			COALESCE(SUM(s.qty), 0) AS sales_qty,
			COALESCE(SUM(s.sub_amount), 0) AS sales_amount,
			c.inventory
		`).
		Group("s.commodity_id, c.name, c.style_code, c.category, c.inventory").
		Order("sales_qty DESC, sales_amount DESC").
		Limit(scope.Limit).
		Scan(&result.HotSKUs).Error; err != nil {
		return nil, err
	}
	if err := productSalesQuery(scope).
		Select(`
			c.style_code,
			COALESCE(SUM(s.qty), 0) AS sales_qty,
			COALESCE(SUM(s.sub_amount), 0) AS sales_amount,
			0 AS inventory
		`).
		Group("c.style_code").
		Order("sales_qty DESC, sales_amount DESC").
		Limit(scope.Limit).
		Scan(&result.HotStyleCodes).Error; err != nil {
		return nil, err
	}

	result.SlowMovingProducts, err = slowMovingProducts(scope)
	if err != nil {
		return nil, err
	}
	if err := fillStyleInventories(scope, result.HotStyleCodes); err != nil {
		return nil, err
	}
	result.LowInventoryCount, err = lowInventoryCount(scope)
	if err != nil {
		return nil, err
	}
	if err := loadInventoryTurnover(scope, result); err != nil {
		return nil, err
	}
	if err := loadReviewSummary(scope, result); err != nil {
		return nil, err
	}
	return result, nil
}

// TrafficSummary 明确返回预留状态，不从请求日志推导访问数据。
func TrafficSummary(filter AnalyticsFilter) (*TrafficSummaryResult, error) {
	if _, err := normalizeAnalyticsFilter(filter); err != nil {
		return nil, err
	}
	return &TrafficSummaryResult{
		Status:  "not_implemented",
		Message: "当前项目尚未接入页面访问埋点，流量分析接口先预留，后续接入 PageViewLog 后再统计访问量和转化率。",
		Data:    []map[string]any{},
	}, nil
}

// ExportAnalytics 返回一次性结构化导出数据。
// 第一版只做 JSON 聚合，文件导出可在接口稳定后基于这个结构生成 Excel 或 CSV。
func ExportAnalytics(filter AnalyticsFilter) (map[string]any, error) {
	sales, err := SalesSummary(filter)
	if err != nil {
		return nil, err
	}
	users, err := UserSummary(filter)
	if err != nil {
		return nil, err
	}
	products, err := ProductSummary(filter)
	if err != nil {
		return nil, err
	}
	traffic, err := TrafficSummary(filter)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"filter":   filter,
		"sales":    sales,
		"users":    users,
		"products": products,
		"traffic":  traffic,
	}, nil
}

func normalizeAnalyticsFilter(filter AnalyticsFilter) (analyticsScope, error) {
	beginTime, err := parseAnalyticsTime(filter.BeginTime, false)
	if err != nil {
		return analyticsScope{}, err
	}
	endTime, err := parseAnalyticsTime(filter.EndTime, true)
	if err != nil {
		return analyticsScope{}, err
	}
	if beginTime != nil && endTime != nil && beginTime.After(*endTime) {
		return analyticsScope{}, fmt.Errorf("begin_time cannot be after end_time")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultAnalyticsLimit
	}
	if limit > maxAnalyticsLimit {
		limit = maxAnalyticsLimit
	}

	lowInventoryThreshold := filter.LowInventoryThreshold
	if lowInventoryThreshold <= 0 {
		lowInventoryThreshold = defaultLowInventoryThreshold
	}

	slowSalesThreshold := filter.SlowSalesThreshold
	if slowSalesThreshold < 0 {
		slowSalesThreshold = defaultSlowSalesThreshold
	}

	return analyticsScope{
		BeginTime:             beginTime,
		EndTime:               endTime,
		Shopname:              strings.TrimSpace(filter.Shopname),
		Category:              strings.TrimSpace(filter.Category),
		StyleCode:             strings.TrimSpace(filter.StyleCode),
		OperatorID:            strings.TrimSpace(filter.OperatorID),
		LowInventoryThreshold: lowInventoryThreshold,
		SlowSalesThreshold:    slowSalesThreshold,
		Limit:                 limit,
	}, nil
}

func parseAnalyticsTime(value string, isEnd bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err != nil {
			continue
		}
		if isEnd && layout == "2006-01-02" {
			parsed = parsed.AddDate(0, 0, 1).Add(-time.Nanosecond)
		}
		return &parsed, nil
	}
	return nil, fmt.Errorf("invalid analytics time %q", value)
}

func orderQuery(scope analyticsScope) *gorm.DB {
	query := db.DB.Model(&models.Order{})
	query = applyTimeRange(query, "order_time", scope)
	if scope.Shopname != "" {
		query = query.Where("order_from = ?", scope.Shopname)
	}
	return query
}

func productSalesQuery(scope analyticsScope) *gorm.DB {
	query := db.DB.Table("sub_order_data AS s").
		Joins("JOIN Commodity_data AS c ON c.commodity_id = s.commodity_id")
	query = applyTimeRange(query, "s.create_time", scope)
	query = applyProductFilter(query, scope)
	if scope.Shopname != "" {
		query = query.Joins("JOIN order_data AS o ON o.order_id = s.order_id").
			Where("o.order_from = ?", scope.Shopname)
	}
	return query
}

func userPreferenceQuery(scope analyticsScope, nonEmptyColumn string) *gorm.DB {
	query := db.DB.Table("order_data AS o").
		Joins("JOIN sub_order_data AS s ON s.order_id = o.order_id").
		Joins("JOIN Commodity_data AS c ON c.commodity_id = s.commodity_id").
		Where("o.user_id > 0").
		Where(nonEmptyColumn + " <> ''")
	query = applyTimeRange(query, "o.order_time", scope)
	query = applyProductFilter(query, scope)
	if scope.Shopname != "" {
		query = query.Where("o.order_from = ?", scope.Shopname)
	}
	return query
}

func repurchaseUserQuery(scope analyticsScope) *gorm.DB {
	subQuery := orderQuery(scope).
		Select("user_id").
		Where("user_id > 0").
		Group("user_id").
		Having("COUNT(*) > 1")
	return db.DB.Table("(?) AS repurchase_users", subQuery)
}

func applyTimeRange(query *gorm.DB, column string, scope analyticsScope) *gorm.DB {
	if scope.BeginTime != nil {
		query = query.Where(column+" >= ?", *scope.BeginTime)
	}
	if scope.EndTime != nil {
		query = query.Where(column+" <= ?", *scope.EndTime)
	}
	return query
}

func applyProductFilter(query *gorm.DB, scope analyticsScope) *gorm.DB {
	if scope.Category != "" {
		query = query.Where("c.category = ?", scope.Category)
	}
	if scope.StyleCode != "" {
		query = query.Where("c.style_code = ?", scope.StyleCode)
	}
	return query
}

func scanFloat(query *gorm.DB) (float64, error) {
	var value float64
	if err := query.Scan(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}

func refundAmount(scope analyticsScope) (float64, error) {
	query := refundQuery(scope).
		Select("COALESCE(SUM(COALESCE(s.sub_amount, o.order_amount, 0)), 0)")
	return scanFloat(query)
}

func refundDaily(scope analyticsScope) (map[string]float64, error) {
	type refundDailyPoint struct {
		Date         string  `gorm:"column:date"`
		RefundAmount float64 `gorm:"column:refund_amount"`
	}
	rows := []refundDailyPoint{}
	query := refundQuery(scope).
		Select("DATE(r.completed_time) AS date, COALESCE(SUM(COALESCE(s.sub_amount, o.order_amount, 0)), 0) AS refund_amount").
		Group("DATE(r.completed_time)")
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.Date] = row.RefundAmount
	}
	return result, nil
}

func refundQuery(scope analyticsScope) *gorm.DB {
	query := db.DB.Table("return_order_data AS r").
		Joins("LEFT JOIN sub_order_data AS s ON s.sub_order_id = r.sub_order_id").
		Joins("LEFT JOIN order_data AS o ON o.order_id = r.order_id").
		Where("r.status IN ?", []string{ReturnOrderStatusCompleted, "returned"}).
		Where("r.type IN ?", []string{"refund", "return", "return_refund"})
	query = applyTimeRange(query, "r.completed_time", scope)
	if scope.Shopname != "" {
		query = query.Where("o.order_from = ?", scope.Shopname)
	}
	return query
}

func loadSalesDaily(scope analyticsScope, result *SalesSummaryResult) error {
	daily := []SalesDailyPoint{}
	if err := orderQuery(scope).
		Select(`
			DATE(order_time) AS date,
			COUNT(*) AS order_count,
			SUM(CASE WHEN pay_status = 'paid' THEN 1 ELSE 0 END) AS paid_order_count,
			SUM(CASE WHEN status IN ('canceled', 'cancelled') THEN 1 ELSE 0 END) AS canceled_order_count,
			COALESCE(SUM(CASE WHEN pay_status = 'paid' THEN final_pay_amount ELSE 0 END), 0) AS sales_amount,
			COALESCE(SUM(CASE WHEN pay_status = 'paid' THEN final_pay_amount ELSE 0 END), 0) AS paid_amount,
			COALESCE(SUM(order_amount), 0) AS original_order_amount,
			COALESCE(SUM(CASE WHEN pay_status = 'paid' THEN discount_amount ELSE 0 END), 0) AS discount_amount
		`).
		Group("DATE(order_time)").
		Order("date ASC").
		Scan(&daily).Error; err != nil {
		return err
	}

	refunds, err := refundDaily(scope)
	if err != nil {
		return err
	}
	for index := range daily {
		daily[index].RefundAmount = refunds[daily[index].Date]
		if daily[index].PaidOrderCount > 0 {
			daily[index].AverageOrderValue = daily[index].PaidAmount / float64(daily[index].PaidOrderCount)
		}
		daily[index].SalesAmount = roundFloat(daily[index].SalesAmount, 2)
		daily[index].PaidAmount = roundFloat(daily[index].PaidAmount, 2)
		daily[index].OriginalOrderAmount = roundFloat(daily[index].OriginalOrderAmount, 2)
		daily[index].DiscountAmount = roundFloat(daily[index].DiscountAmount, 2)
		daily[index].RefundAmount = roundFloat(daily[index].RefundAmount, 2)
		daily[index].AverageOrderValue = roundFloat(daily[index].AverageOrderValue, 2)
	}
	result.Daily = daily
	return nil
}

func slowMovingProducts(scope analyticsScope) ([]ProductSalesPoint, error) {
	type soldRow struct {
		CommodityID string `gorm:"column:commodity_id"`
		SalesQty    int64  `gorm:"column:sales_qty"`
	}
	soldRows := []soldRow{}
	if err := productSalesQuery(scope).
		Select("s.commodity_id, COALESCE(SUM(s.qty), 0) AS sales_qty").
		Group("s.commodity_id").
		Scan(&soldRows).Error; err != nil {
		return nil, err
	}

	soldQtyByCommodity := make(map[string]int64, len(soldRows))
	for _, row := range soldRows {
		soldQtyByCommodity[row.CommodityID] = row.SalesQty
	}

	commodities := []models.Commodity{}
	query := db.DB.Model(&models.Commodity{}).Where("inventory > 0")
	if scope.Category != "" {
		query = query.Where("category = ?", scope.Category)
	}
	if scope.StyleCode != "" {
		query = query.Where("style_code = ?", scope.StyleCode)
	}
	if err := query.Order("inventory DESC, commodity_id ASC").Find(&commodities).Error; err != nil {
		return nil, err
	}

	result := make([]ProductSalesPoint, 0)
	for _, commodity := range commodities {
		salesQty := soldQtyByCommodity[commodity.CommodityID]
		if salesQty > int64(scope.SlowSalesThreshold) {
			continue
		}
		result = append(result, ProductSalesPoint{
			CommodityID: commodity.CommodityID,
			Name:        commodity.Name,
			StyleCode:   commodity.StyleCode,
			Category:    commodity.Category,
			SalesQty:    salesQty,
			Inventory:   commodity.Inventory,
		})
		if len(result) >= scope.Limit {
			break
		}
	}
	return result, nil
}

func lowInventoryCount(scope analyticsScope) (int64, error) {
	query := db.DB.Model(&models.Commodity{}).
		Where("inventory <= ?", scope.LowInventoryThreshold)
	if scope.Category != "" {
		query = query.Where("category = ?", scope.Category)
	}
	if scope.StyleCode != "" {
		query = query.Where("style_code = ?", scope.StyleCode)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func fillStyleInventories(scope analyticsScope, rows []StyleSalesPoint) error {
	inventoryByStyleCode, err := styleInventoryByCode(scope)
	if err != nil {
		return err
	}
	for index := range rows {
		rows[index].Inventory = inventoryByStyleCode[rows[index].StyleCode]
		rows[index].SalesAmount = roundFloat(rows[index].SalesAmount, 2)
	}
	return nil
}

func styleInventoryByCode(scope analyticsScope) (map[string]int64, error) {
	type inventoryRow struct {
		StyleCode string `gorm:"column:style_code"`
		Inventory int64  `gorm:"column:inventory"`
	}
	rows := []inventoryRow{}
	query := db.DB.Model(&models.Commodity{}).
		Select("style_code, COALESCE(SUM(inventory), 0) AS inventory").
		Where("style_code <> ''").
		Group("style_code")
	if scope.Category != "" {
		query = query.Where("category = ?", scope.Category)
	}
	if scope.StyleCode != "" {
		query = query.Where("style_code = ?", scope.StyleCode)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(rows))
	for _, row := range rows {
		result[row.StyleCode] = row.Inventory
	}
	return result, nil
}

func loadInventoryTurnover(scope analyticsScope, result *ProductSummaryResult) error {
	totalSalesQty, err := scanFloat(productSalesQuery(scope).Select("COALESCE(SUM(s.qty), 0)"))
	if err != nil {
		return err
	}

	query := db.DB.Model(&models.Commodity{})
	if scope.Category != "" {
		query = query.Where("category = ?", scope.Category)
	}
	if scope.StyleCode != "" {
		query = query.Where("style_code = ?", scope.StyleCode)
	}
	totalInventory, err := scanFloat(query.Select("COALESCE(SUM(inventory), 0)"))
	if err != nil {
		return err
	}
	if totalInventory > 0 {
		result.InventoryTurnoverRate = roundFloat(totalSalesQty/totalInventory, 4)
	}
	return nil
}

func loadReviewSummary(scope analyticsScope, result *ProductSummaryResult) error {
	var total int64
	if err := reviewSummaryQuery(scope).Count(&total).Error; err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	var averageRating float64
	if err := reviewSummaryQuery(scope).Select("COALESCE(AVG(r.rating), 0)").Scan(&averageRating).Error; err != nil {
		return err
	}

	var goodCount int64
	if err := reviewSummaryQuery(scope).Where("r.rating >= ?", 4).Count(&goodCount).Error; err != nil {
		return err
	}
	result.AverageRating = roundFloat(averageRating, 2)
	result.GoodRate = roundFloat(float64(goodCount)/float64(total), 4)
	return nil
}

func reviewSummaryQuery(scope analyticsScope) *gorm.DB {
	query := db.DB.Table("product_reviews AS r").
		Joins("JOIN Commodity_data AS c ON c.commodity_id = r.commodity_id").
		Where("r.status = ?", models.ReviewStatusApproved)
	query = applyTimeRange(query, "r.created_at", scope)
	query = applyProductFilter(query, scope)
	return query
}

func roundSalesSummary(result *SalesSummaryResult) {
	result.SalesAmount = roundFloat(result.SalesAmount, 2)
	result.PaidAmount = roundFloat(result.PaidAmount, 2)
	result.OriginalOrderAmount = roundFloat(result.OriginalOrderAmount, 2)
	result.DiscountAmount = roundFloat(result.DiscountAmount, 2)
	result.RefundAmount = roundFloat(result.RefundAmount, 2)
	result.AverageOrderValue = roundFloat(result.AverageOrderValue, 2)
}

func roundFloat(value float64, precision int) float64 {
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}
