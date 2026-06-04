package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/service/jushuitan"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ReturnOrderStatusPending      = "pending"
	ReturnOrderStatusApproved     = "approved"
	ReturnOrderStatusRejected     = "rejected"
	ReturnOrderStatusBuyerShipped = "buyer_shipped"
	ReturnOrderStatusReceived     = "received"
	ReturnOrderStatusCompleted    = "completed"
	ReturnOrderStatusCanceled     = "canceled"

	JushuitanPushStatusPending = "pending"
	JushuitanPushStatusSuccess = "success"
	JushuitanPushStatusFailed  = "failed"
)

var returnOrderTerminalStatuses = map[string]bool{
	ReturnOrderStatusRejected:  true,
	ReturnOrderStatusCompleted: true,
	ReturnOrderStatusCanceled:  true,
	"returned":                 true,
}

// ReturnOrderCreateInput 是售后统一创建入口的内部参数。
// /return_order/create 和 /order/request_return 都先组装这个结构，再走同一套校验与落库逻辑。
type ReturnOrderCreateInput struct {
	UserID          int
	OrderID         string
	SubOrderID      string
	OrderStatus     string
	Type            string
	Reason          string
	SpecificReasons string
	ProductIDs      []string
	ProductList     string
	BuyerProvince   string
	BuyerCity       string
	BuyerCounty     string
	BuyerAddress    string
	BuyerPhone      string
	Remark          string
}

type ReturnReasonRank struct {
	Reason string `json:"reason" gorm:"column:reason"`
	Count  int64  `json:"count" gorm:"column:count"`
}

type ReturnOrderStatisticsResult struct {
	TotalCount      int64              `json:"total_count"`
	PendingCount    int64              `json:"pending_count"`
	CompletedCount  int64              `json:"completed_count"`
	AfterSaleRate   float64            `json:"after_sale_rate"`
	AfterSaleAmount float64            `json:"after_sale_amount"`
	ReasonRank      []ReturnReasonRank `json:"reason_rank"`
	CompletedOrders int64              `json:"completed_orders"`
	AfterSaleOrders int64              `json:"after_sale_orders"`
}

func ConvertReturnOrderToMap(returnOrder models.ReturnOrder) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = returnOrder.ID
	result["user_id"] = returnOrder.UserID
	result["return_id"] = returnOrder.ReturnID
	result["order_id"] = returnOrder.OrderID
	result["sub_order_id"] = returnOrder.SubOrderID
	result["sub_order_product_info"] = returnOrder.SubOrderProductInfo
	result["order_status"] = returnOrder.OrderStatus
	result["product_list"] = returnOrder.ProductList
	result["type"] = returnOrder.Type
	result["status"] = returnOrder.Status
	result["after_sale_status"] = returnOrder.Status
	result["is_after_sale_completed"] = isAfterSaleCompleted(returnOrder.Status)
	result["display_gray"] = shouldDisplayGrayForAfterSale(returnOrder.Status)
	result["express_company"] = returnOrder.ExpressCompany
	result["express_number"] = returnOrder.ExpressNumber
	result["reason"] = returnOrder.Reason
	result["specific_reasons"] = returnOrder.SpecificReasons
	result["buyer_province"] = returnOrder.BuyerProvince
	result["buyer_city"] = returnOrder.BuyerCity
	result["buyer_county"] = returnOrder.BuyerCounty
	result["buyer_address"] = returnOrder.BuyerAddress
	result["buyer_phone"] = returnOrder.BuyerPhone
	result["remarks"] = returnOrder.Remarks
	result["jushuitan_after_sale_id"] = returnOrder.JushuitanAfterSaleID
	result["jushuitan_push_status"] = returnOrder.JushuitanPushStatus
	result["jushuitan_push_response"] = returnOrder.JushuitanPushResponse

	if returnOrder.RequestTime != nil {
		result["request_time"] = returnOrder.RequestTime.Format("2006-01-02 15:04:05")
	} else {
		result["request_time"] = ""
	}

	if returnOrder.ShippedTime != nil {
		result["shipped_time"] = returnOrder.ShippedTime.Format("2006-01-02 15:04:05")
	} else {
		result["shipped_time"] = ""
	}

	if returnOrder.CanceledTime != nil {
		result["canceled_time"] = returnOrder.CanceledTime.Format("2006-01-02 15:04:05")
	} else {
		result["canceled_time"] = ""
	}

	if returnOrder.CompletedTime != nil {
		result["completed_time"] = returnOrder.CompletedTime.Format("2006-01-02 15:04:05")
	} else {
		result["completed_time"] = ""
	}

	if returnOrder.JushuitanUpdatedTime != nil {
		result["jushuitan_updated_time"] = returnOrder.JushuitanUpdatedTime.Format("2006-01-02 15:04:05")
	} else {
		result["jushuitan_updated_time"] = ""
	}

	return result
}

func ConvertReturnOrdersToMap(returnOrders []models.ReturnOrder) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(returnOrders))
	for _, returnOrder := range returnOrders {
		result = append(result, ConvertReturnOrderToMap(returnOrder))
	}
	return result
}

// CreateReturnOrder 创建退货订单
func CreateReturnOrder(returnOrder *models.ReturnOrder) error {
	// 生成退货订单号
	returnOrder.ReturnID = fmt.Sprintf("RET%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	returnOrder.Status = "pending"
	timeNow := time.Now()
	returnOrder.RequestTime = &timeNow

	return db.DB.Create(returnOrder).Error
}

// CreateReturnOrderFromInput 创建售后单。
// 这里只负责申请售后，不扣减或回滚库存；库存回滚统一放在仓库收货并完成售后时处理。
func CreateReturnOrderFromInput(input ReturnOrderCreateInput) (*ReturnOrderResult, error) {
	var result *ReturnOrderResult
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		order, returnOrder, err := buildReturnOrderForCreate(tx, input)
		if err != nil {
			return err
		}

		if err := tx.Create(returnOrder).Error; err != nil {
			return err
		}

		result = &ReturnOrderResult{
			ReturnOrder: returnOrder,
			Order:       order,
			ReturnID:    returnOrder.ReturnID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if pushErr := PushReturnOrderToJushuitan(result.ReturnID); pushErr != nil {
		result.ReturnOrder.JushuitanPushStatus = JushuitanPushStatusFailed
		result.ReturnOrder.JushuitanPushResponse = pushErr.Error()
	}
	_ = db.DB.Where("return_id = ?", result.ReturnID).First(result.ReturnOrder).Error
	return result, nil
}

func ReturnOrderStatistics(beginTime, endTime string) (*ReturnOrderStatisticsResult, error) {
	begin, end, err := parseReturnOrderTimeRange(beginTime, endTime)
	if err != nil {
		return nil, err
	}

	result := &ReturnOrderStatisticsResult{ReasonRank: []ReturnReasonRank{}}
	query := returnOrderStatisticsQuery(begin, end)
	if err := query.Count(&result.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := returnOrderStatisticsQuery(begin, end).Where("status = ?", ReturnOrderStatusPending).Count(&result.PendingCount).Error; err != nil {
		return nil, err
	}
	if err := returnOrderStatisticsQuery(begin, end).Where("status IN ?", []string{ReturnOrderStatusCompleted, "returned"}).Count(&result.CompletedCount).Error; err != nil {
		return nil, err
	}
	if err := returnOrderStatisticsQuery(begin, end).
		Select("COUNT(DISTINCT order_id)").
		Scan(&result.AfterSaleOrders).Error; err != nil {
		return nil, err
	}
	if err := completedOrderStatisticsQuery(begin, end).Count(&result.CompletedOrders).Error; err != nil {
		return nil, err
	}
	result.AfterSaleRate = calculateAfterSaleRate(result.AfterSaleOrders, result.CompletedOrders)

	if err := returnOrderAmountQuery(begin, end).
		Scan(&result.AfterSaleAmount).Error; err != nil {
		return nil, err
	}
	result.AfterSaleAmount = roundFloat(result.AfterSaleAmount, 2)

	if err := returnOrderStatisticsQuery(begin, end).
		Select("reason, COUNT(*) AS count").
		Where("reason <> ''").
		Group("reason").
		Order("count DESC").
		Limit(10).
		Scan(&result.ReasonRank).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func returnOrderStatisticsQuery(begin, end *time.Time) *gorm.DB {
	query := db.DB.Model(&models.ReturnOrder{})
	if begin != nil {
		query = query.Where("request_time >= ?", *begin)
	}
	if end != nil {
		query = query.Where("request_time <= ?", *end)
	}
	return query
}

func completedOrderStatisticsQuery(begin, end *time.Time) *gorm.DB {
	query := db.DB.Model(&models.Order{}).Where("status = ?", "delivered")
	if begin != nil {
		query = query.Where("delivered_time >= ?", *begin)
	}
	if end != nil {
		query = query.Where("delivered_time <= ?", *end)
	}
	return query
}

func returnOrderAmountQuery(begin, end *time.Time) *gorm.DB {
	query := db.DB.Table("return_order_data AS r").
		Joins("LEFT JOIN sub_order_data AS s ON s.sub_order_id = r.sub_order_id").
		Joins("LEFT JOIN order_data AS o ON o.order_id = r.order_id").
		Where("r.status IN ?", []string{ReturnOrderStatusCompleted, "returned"}).
		Where("r.type IN ?", []string{"refund", "return", "return_refund"})
	if begin != nil {
		query = query.Where("r.completed_time >= ?", *begin)
	}
	if end != nil {
		query = query.Where("r.completed_time <= ?", *end)
	}
	return query.Select("COALESCE(SUM(COALESCE(s.sub_amount, o.final_pay_amount, o.order_amount, 0)), 0)")
}

func parseReturnOrderTimeRange(beginTime, endTime string) (*time.Time, *time.Time, error) {
	begin, err := parseReturnOrderTime(beginTime, false)
	if err != nil {
		return nil, nil, err
	}
	end, err := parseReturnOrderTime(endTime, true)
	if err != nil {
		return nil, nil, err
	}
	if begin != nil && end != nil && begin.After(*end) {
		return nil, nil, fmt.Errorf("begin_time cannot be after end_time")
	}
	return begin, end, nil
}

func parseReturnOrderTime(value string, isEnd bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
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
	return nil, fmt.Errorf("invalid return order time %q", value)
}

func calculateAfterSaleRate(afterSaleOrders, completedOrders int64) float64 {
	if completedOrders <= 0 {
		return 0
	}
	return roundFloat(float64(afterSaleOrders)/float64(completedOrders), 4)
}

func buildReturnOrderForCreate(tx *gorm.DB, input ReturnOrderCreateInput) (*models.Order, *models.ReturnOrder, error) {
	input.OrderID = strings.TrimSpace(input.OrderID)
	input.SubOrderID = strings.TrimSpace(input.SubOrderID)
	input.Type = normalizeReturnType(input.Type)
	input.Reason = strings.TrimSpace(input.Reason)
	input.SpecificReasons = strings.TrimSpace(input.SpecificReasons)

	if input.Type == "" {
		input.Type = "return"
	}
	if input.OrderID == "" {
		return nil, nil, fmt.Errorf("order_id不能为空")
	}
	if input.UserID <= 0 {
		return nil, nil, fmt.Errorf("user_id不能为空")
	}
	if input.Reason == "" {
		return nil, nil, fmt.Errorf("售后原因不能为空")
	}
	if input.SpecificReasons == "" {
		input.SpecificReasons = input.Reason
	}

	var order models.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", input.OrderID).First(&order).Error; err != nil {
		return nil, nil, err
	}
	if order.UserID != 0 && order.UserID != input.UserID {
		return nil, nil, fmt.Errorf("售后申请用户与订单用户不一致")
	}
	if order.Status == "canceled" || order.Status == "cancelled" {
		return nil, nil, fmt.Errorf("订单已取消，不能发起售后")
	}

	orderStatus := input.OrderStatus
	if orderStatus == "" {
		orderStatus = order.Status
	}

	productList := strings.TrimSpace(input.ProductList)
	subOrderProductInfo := ""
	if input.SubOrderID != "" {
		// 子订单售后只允许关联自己的商品信息，避免前端传入其它订单的商品造成串单。
		subOrder, err := loadReturnSubOrder(tx, input.OrderID, input.SubOrderID)
		if err != nil {
			return nil, nil, err
		}
		subOrderProductInfo = subOrder.ProductInfo
		if productList == "" {
			productList = subOrder.ProductInfo
		}
	} else {
		// 主订单维度售后先拦截未结束的整单售后，避免同一批商品被重复申请。
		if err := ensureNoActiveReturnOrder(tx, input.OrderID, ""); err != nil {
			return nil, nil, err
		}
		if productList == "" {
			productList = filterReturnProductList(order.ProductList, input.ProductIDs)
		}
	}
	if input.SubOrderID != "" {
		if err := ensureNoActiveReturnOrder(tx, input.OrderID, input.SubOrderID); err != nil {
			return nil, nil, err
		}
	}
	if productList == "" {
		return nil, nil, fmt.Errorf("售后商品信息不能为空")
	}

	now := time.Now()
	returnOrder := models.ReturnOrder{
		ReturnID:            GenerateReturnOrderNo(),
		UserID:              input.UserID,
		OrderID:             input.OrderID,
		SubOrderID:          input.SubOrderID,
		SubOrderProductInfo: subOrderProductInfo,
		OrderStatus:         orderStatus,
		ProductList:         productList,
		Type:                input.Type,
		Status:              ReturnOrderStatusPending,
		RequestTime:         &now,
		Reason:              input.Reason,
		SpecificReasons:     input.SpecificReasons,
		BuyerProvince:       input.BuyerProvince,
		BuyerCity:           input.BuyerCity,
		BuyerCounty:         input.BuyerCounty,
		BuyerAddress:        input.BuyerAddress,
		BuyerPhone:          input.BuyerPhone,
		Remarks:             input.Remark,
		JushuitanPushStatus: JushuitanPushStatusPending,
	}
	return &order, &returnOrder, nil
}

func normalizeReturnType(returnType string) string {
	switch strings.TrimSpace(returnType) {
	case "", "return", "return_refund":
		return "return"
	case "exchange":
		return "exchange"
	case "refund", "仅退款":
		return "refund"
	case "replacement", "reissue", "补发":
		return "replacement"
	default:
		return strings.TrimSpace(returnType)
	}
}

func loadReturnSubOrder(tx *gorm.DB, orderID, subOrderID string) (*models.SubOrder, error) {
	var subOrder models.SubOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sub_order_id = ?", subOrderID).First(&subOrder).Error; err != nil {
		return nil, err
	}
	if subOrder.OrderID != orderID {
		return nil, fmt.Errorf("子订单不属于当前主订单")
	}
	return &subOrder, nil
}

func ensureNoActiveReturnOrder(tx *gorm.DB, orderID, subOrderID string) error {
	query := tx.Model(&models.ReturnOrder{}).
		Where("order_id = ? AND status NOT IN ?", orderID, []string{
			ReturnOrderStatusRejected,
			ReturnOrderStatusCompleted,
			ReturnOrderStatusCanceled,
		})
	if subOrderID != "" {
		query = query.Where("sub_order_id = ?", subOrderID)
	} else {
		query = query.Where("sub_order_id = '' OR sub_order_id IS NULL")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该商品或子订单存在未结束售后")
	}
	return nil
}

func filterReturnProductList(orderProductList string, productIDs []string) string {
	if len(productIDs) == 0 {
		return orderProductList
	}

	productIDMap := make(map[string]bool, len(productIDs))
	for _, productID := range productIDs {
		productIDMap[strings.TrimSpace(productID)] = true
	}

	var products []map[string]interface{}
	if err := json.Unmarshal([]byte(orderProductList), &products); err != nil {
		productIDsJSON, _ := json.Marshal(productIDs)
		return string(productIDsJSON)
	}

	filteredProducts := make([]map[string]interface{}, 0, len(products))
	for _, product := range products {
		productID := firstStringValue(product, "commodity_id", "sku_id", "product_id", "id")
		if productIDMap[productID] {
			filteredProducts = append(filteredProducts, product)
		}
	}
	if len(filteredProducts) == 0 {
		return ""
	}

	filteredProductData, _ := json.Marshal(filteredProducts)
	return string(filteredProductData)
}

// GetReturnOrderByID 根据ID获取退货订单
func GetReturnOrderByID(returnOrderID string) (*models.ReturnOrder, error) {
	var returnOrder models.ReturnOrder
	err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error
	if err != nil {
		return nil, err
	}
	return &returnOrder, nil
}

// GetReturnOrderByOrderID 根据订单ID获取退货订单
func GetReturnOrderByOrderID(orderID string) ([]models.ReturnOrder, error) {
	var returnOrders []models.ReturnOrder
	err := db.DB.Where("order_id = ?", orderID).Find(&returnOrders).Error
	if err != nil {
		return nil, err
	}
	return returnOrders, nil
}

// ReturnOrderApprove 审核退货订单
func ReturnOrderApprove(returnOrderID, approveStatus, remark string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var returnOrder models.ReturnOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
			return err
		}

		if returnOrder.Status != ReturnOrderStatusPending {
			return fmt.Errorf("退货订单状态不允许审核")
		}

		if approveStatus != ReturnOrderStatusApproved && approveStatus != ReturnOrderStatusRejected {
			return fmt.Errorf("审核状态不正确")
		}
		if approveStatus == ReturnOrderStatusRejected && strings.TrimSpace(remark) == "" {
			return fmt.Errorf("拒绝售后必须填写拒绝原因")
		}

		if err := validateReturnOrderBeforeApprove(tx, returnOrder); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":  approveStatus,
			"remarks": remark,
		}
		if err := tx.Model(&models.ReturnOrder{}).Where("return_id = ?", returnOrderID).Updates(updates).Error; err != nil {
			return err
		}
		if approveStatus == ReturnOrderStatusApproved {
			return markOrderAfterSaleProcessingTx(tx, returnOrder)
		}
		return nil
	})
}

func validateReturnOrderBeforeApprove(tx *gorm.DB, returnOrder models.ReturnOrder) error {
	var order models.Order
	if err := tx.Where("order_id = ?", returnOrder.OrderID).First(&order).Error; err != nil {
		return err
	}
	if order.Status == "canceled" || order.Status == "cancelled" {
		return fmt.Errorf("订单已取消，不能审核通过售后")
	}
	if returnOrder.SubOrderID != "" {
		if _, err := loadReturnSubOrder(tx, returnOrder.OrderID, returnOrder.SubOrderID); err != nil {
			return err
		}
	}
	return ensureNoOtherActiveReturnOrder(tx, returnOrder)
}

func ensureNoOtherActiveReturnOrder(tx *gorm.DB, returnOrder models.ReturnOrder) error {
	query := tx.Model(&models.ReturnOrder{}).
		Where("return_id <> ? AND order_id = ? AND status NOT IN ?", returnOrder.ReturnID, returnOrder.OrderID, []string{
			ReturnOrderStatusRejected,
			ReturnOrderStatusCompleted,
			ReturnOrderStatusCanceled,
		})
	if returnOrder.SubOrderID != "" {
		query = query.Where("sub_order_id = ?", returnOrder.SubOrderID)
	} else {
		query = query.Where("sub_order_id = '' OR sub_order_id IS NULL")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该商品或子订单存在未结束售后")
	}
	return nil
}

func markOrderAfterSaleProcessingTx(tx *gorm.DB, returnOrder models.ReturnOrder) error {
	// 审核通过才把订单或子订单标记为售后中，申请阶段只保留售后单本身的 pending 状态。
	if returnOrder.SubOrderID != "" {
		status := "returning"
		if returnOrder.Type == "exchange" {
			status = "exchanging"
		} else if returnOrder.Type == "replacement" {
			status = "replacing"
		}
		if err := tx.Model(&models.SubOrder{}).
			Where("sub_order_id = ?", returnOrder.SubOrderID).
			Update("status", status).Error; err != nil {
			return err
		}
		if err := syncMainOrderSubOrderStatusTx(tx, returnOrder.OrderID); err != nil {
			return err
		}
	}
	return tx.Model(&models.Order{}).
		Where("order_id = ?", returnOrder.OrderID).
		Updates(map[string]interface{}{
			"status":      "processing",
			"process_num": returnOrder.ReturnID,
		}).Error
}

func syncMainOrderSubOrderStatusTx(tx *gorm.DB, orderID string) error {
	var subOrders []models.SubOrder
	if err := tx.Where("order_id = ?", orderID).Find(&subOrders).Error; err != nil {
		return err
	}
	if len(subOrders) == 0 {
		return nil
	}

	subOrderIDs := make([]string, 0, len(subOrders))
	for _, subOrder := range subOrders {
		subOrderIDs = append(subOrderIDs, subOrder.SubOrderID+":"+subOrder.Status)
	}
	subOrderIDsJSON, _ := json.Marshal(subOrderIDs)
	return tx.Model(&models.Order{}).Where("order_id = ?", orderID).Update("sub_order_ids", string(subOrderIDsJSON)).Error
}

// GetReturnOrders 获取退货订单列表
func GetReturnOrders(returnOrderID, orderID string, userID int, status string, page, pageSize int) ([]models.ReturnOrder, int64, error) {
	var returnOrders []models.ReturnOrder
	var total int64

	query := db.DB.Model(&models.ReturnOrder{})

	if returnOrderID != "" {
		query = query.Where("return_id = ?", returnOrderID)
	}
	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("request_time DESC").Offset(offset).Limit(pageSize).Find(&returnOrders).Error; err != nil {
		return nil, 0, err
	}

	return returnOrders, total, nil
}

func decorateOrderAfterSaleFields(result map[string]interface{}, orderID string) {
	status := latestAfterSaleStatus("order_id = ?", orderID)
	setAfterSaleDisplayFields(result, status)
}

func decorateSubOrderAfterSaleFields(result map[string]interface{}, subOrderID string) {
	status := latestAfterSaleStatus("sub_order_id = ?", subOrderID)
	setAfterSaleDisplayFields(result, status)
}

func latestAfterSaleStatusByOrderIDs(orderIDs []string) map[string]string {
	return latestAfterSaleStatusByIDs("order_id", orderIDs)
}

func latestAfterSaleStatusBySubOrderIDs(subOrderIDs []string) map[string]string {
	return latestAfterSaleStatusByIDs("sub_order_id", subOrderIDs)
}

func latestAfterSaleStatusByIDs(column string, values []string) map[string]string {
	result := make(map[string]string)
	if len(values) == 0 {
		return result
	}

	uniqueValues := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		uniqueValues = append(uniqueValues, value)
	}
	if len(uniqueValues) == 0 {
		return result
	}

	var returnOrders []models.ReturnOrder
	if err := db.DB.Model(&models.ReturnOrder{}).
		Where(column+" IN ?", uniqueValues).
		Order(column + " ASC, request_time DESC, id DESC").
		Find(&returnOrders).Error; err != nil {
		return result
	}

	for _, returnOrder := range returnOrders {
		var key string
		if column == "sub_order_id" {
			key = returnOrder.SubOrderID
		} else {
			key = returnOrder.OrderID
		}
		if key != "" {
			if _, ok := result[key]; !ok {
				result[key] = returnOrder.Status
			}
		}
	}

	return result
}

func latestAfterSaleStatus(query string, value interface{}) string {
	if value == "" {
		return ""
	}

	var returnOrder models.ReturnOrder
	err := db.DB.Where(query, value).
		Order("request_time DESC, id DESC").
		First(&returnOrder).Error
	if err != nil {
		return ""
	}
	return returnOrder.Status
}

func setAfterSaleDisplayFields(result map[string]interface{}, status string) {
	// 这三个字段专门给前端做售后展示：状态负责文案，completed/display_gray 负责置灰判断。
	result["after_sale_status"] = status
	result["is_after_sale_completed"] = isAfterSaleCompleted(status)
	result["display_gray"] = shouldDisplayGrayForAfterSale(status)
}

func isAfterSaleCompleted(status string) bool {
	return returnOrderTerminalStatuses[status] && status != ReturnOrderStatusRejected && status != ReturnOrderStatusCanceled
}

func shouldDisplayGrayForAfterSale(status string) bool {
	return status == ReturnOrderStatusCompleted || status == "returned"
}

func PushReturnOrderToJushuitan(returnOrderID string) error {
	var returnOrder models.ReturnOrder
	if err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
		return err
	}

	var order models.Order
	if err := db.DB.Where("order_id = ?", returnOrder.OrderID).First(&order).Error; err != nil {
		return err
	}

	token, err := jushuitan.GetToken()
	if err != nil {
		_ = updateReturnOrderJushuitanPushResult(returnOrder.ReturnID, JushuitanPushStatusFailed, "", err.Error())
		return err
	}

	payload := buildJushuitanAfterSaleData(order, returnOrder)
	resp, err := jushuitan.SendAfterSale(token, payload)
	if err != nil {
		_ = updateReturnOrderJushuitanPushResult(returnOrder.ReturnID, JushuitanPushStatusFailed, "", err.Error())
		return err
	}

	jushuitanAfterSaleID := extractJushuitanAfterSaleID(resp)
	return updateReturnOrderJushuitanPushResult(returnOrder.ReturnID, JushuitanPushStatusSuccess, jushuitanAfterSaleID, resp)
}

func updateReturnOrderJushuitanPushResult(returnID, status, jushuitanAfterSaleID, response string) error {
	updates := map[string]interface{}{
		"jushuitan_push_status":   status,
		"jushuitan_push_response": response,
		"jushuitan_updated_time":  time.Now(),
	}
	if jushuitanAfterSaleID != "" {
		updates["jushuitan_after_sale_id"] = jushuitanAfterSaleID
	}
	return db.DB.Model(&models.ReturnOrder{}).Where("return_id = ?", returnID).Updates(updates).Error
}

func buildJushuitanAfterSaleData(order models.Order, returnOrder models.ReturnOrder) jushuitan.AfterSaleData {
	requestTime := time.Now()
	if returnOrder.RequestTime != nil {
		requestTime = *returnOrder.RequestTime
	}
	receiverState := firstNonEmpty(returnOrder.BuyerProvince, order.Province)
	receiverCity := firstNonEmpty(returnOrder.BuyerCity, order.City)
	receiverDistrict := firstNonEmpty(returnOrder.BuyerCounty, order.County)
	receiverAddress := strings.TrimSpace(returnOrder.BuyerAddress)
	if receiverAddress == "" {
		receiverAddress = fmt.Sprintf("%s_%s_%s_%s", order.Province, order.City, order.County, order.DetailedAddress)
	}

	return jushuitan.AfterSaleData{
		ShopID:           10395227,
		OuterASID:        returnOrder.ReturnID,
		SoID:             returnOrder.OrderID,
		Type:             jushuitanAfterSaleType(returnOrder.Type),
		ShopStatus:       "TRADE_AFTER_SALE",
		QuestionType:     returnOrder.Reason,
		Reason:           returnOrder.SpecificReasons,
		Remark:           returnOrder.Remarks,
		Created:          requestTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		Modified:         time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		BuyerAccount:     strconv.Itoa(returnOrder.UserID),
		ReceiverState:    receiverState,
		ReceiverCity:     receiverCity,
		ReceiverDistrict: receiverDistrict,
		ReceiverAddress:  receiverAddress,
		ReceiverPhone:    firstNonEmpty(returnOrder.BuyerPhone, order.ReceiverPhone),
		Items:            buildJushuitanAfterSaleItems(returnOrder),
	}
}

func buildJushuitanAfterSaleItems(returnOrder models.ReturnOrder) []jushuitan.AfterSaleItem {
	productList := strings.TrimSpace(returnOrder.ProductList)
	if returnOrder.SubOrderProductInfo != "" {
		productList = returnOrder.SubOrderProductInfo
	}

	rawItems := parseAfterSaleProductItems(productList)
	if len(rawItems) == 0 {
		return []jushuitan.AfterSaleItem{{
			OuterOiID: returnOrder.SubOrderID,
			SkuID:     returnOrder.SubOrderID,
			Qty:       1,
			Type:      jushuitanAfterSaleItemType(returnOrder.Type),
		}}
	}

	items := make([]jushuitan.AfterSaleItem, 0, len(rawItems))
	for _, raw := range rawItems {
		qty := firstIntValue(raw, 1, "qty", "quantity", "num")
		if qty <= 0 {
			qty = 1
		}
		items = append(items, jushuitan.AfterSaleItem{
			OuterOiID: firstNonEmpty(returnOrder.SubOrderID, firstStringValue(raw, "outer_oi_id", "sub_order_id")),
			SkuID:     firstStringValue(raw, "sku_id", "commodity_id", "product_id", "id"),
			ShopSkuID: firstStringValue(raw, "shop_sku_id", "commodity_id", "sku_id"),
			Name:      firstStringValue(raw, "name", "product_name", "commodity_name"),
			Qty:       qty,
			Amount:    firstFloatValue(raw, "amount", "sub_amount", "price"),
			Type:      jushuitanAfterSaleItemType(returnOrder.Type),
		})
	}
	if len(items) == 0 {
		return []jushuitan.AfterSaleItem{{
			OuterOiID: returnOrder.SubOrderID,
			SkuID:     returnOrder.SubOrderID,
			Qty:       1,
			Type:      jushuitanAfterSaleItemType(returnOrder.Type),
		}}
	}
	return items
}

func parseAfterSaleProductItems(productList string) []map[string]interface{} {
	var rawItems []map[string]interface{}
	if err := json.Unmarshal([]byte(productList), &rawItems); err == nil {
		return rawItems
	}

	var rawItem map[string]interface{}
	if err := json.Unmarshal([]byte(productList), &rawItem); err == nil && len(rawItem) > 0 {
		return []map[string]interface{}{rawItem}
	}
	return nil
}

func firstFloatValue(m map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		value, ok := m[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case json.Number:
			parsed, _ := v.Float64()
			return parsed
		case string:
			parsed, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
			return parsed
		}
	}
	return 0
}

func jushuitanAfterSaleType(returnType string) string {
	switch normalizeReturnType(returnType) {
	case "refund":
		return "仅退款"
	case "exchange":
		return "换货"
	case "replacement":
		return "补发"
	default:
		return "退货退款"
	}
}

func jushuitanAfterSaleItemType(returnType string) string {
	switch normalizeReturnType(returnType) {
	case "exchange":
		return "换货"
	case "replacement":
		return "补发"
	case "refund":
		return "其它"
	default:
		return "退货"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func extractJushuitanAfterSaleID(resp string) string {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		return ""
	}
	for _, key := range []string{"as_id", "jushuitan_after_sale_id"} {
		if value, ok := parsed[key]; ok {
			return fmt.Sprint(value)
		}
	}
	if data, ok := parsed["data"].(map[string]interface{}); ok {
		for _, key := range []string{"as_id", "jushuitan_after_sale_id"} {
			if value, ok := data[key]; ok {
				return fmt.Sprint(value)
			}
		}
	}
	return ""
}

type JushuitanAfterSaleUpdateInput struct {
	ReturnID             string
	JushuitanAfterSaleID string
	OrderID              string
	Status               string
	Response             string
}

func ApplyJushuitanAfterSaleUpdate(input JushuitanAfterSaleUpdateInput) error {
	normalizedStatus := NormalizeJushuitanAfterSaleStatus(input.Status)
	if normalizedStatus == "" {
		return fmt.Errorf("聚水潭售后状态不能为空")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		returnOrder, err := lockReturnOrderForJushuitanUpdate(tx, input)
		if err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":                  normalizedStatus,
			"jushuitan_push_response": input.Response,
			"jushuitan_updated_time":  time.Now(),
		}
		if input.JushuitanAfterSaleID != "" {
			updates["jushuitan_after_sale_id"] = input.JushuitanAfterSaleID
		}

		switch normalizedStatus {
		case ReturnOrderStatusApproved:
			if err := markOrderAfterSaleProcessingTx(tx, *returnOrder); err != nil {
				return err
			}
		case ReturnOrderStatusReceived:
			if err := RestoreInventoryForReturn(tx, *returnOrder); err != nil {
				return err
			}
		case ReturnOrderStatusCompleted:
			now := time.Now()
			updates["completed_time"] = &now
		case ReturnOrderStatusCanceled:
			now := time.Now()
			updates["canceled_time"] = &now
		}

		return tx.Model(&models.ReturnOrder{}).Where("return_id = ?", returnOrder.ReturnID).Updates(updates).Error
	})
}

func ApplyJushuitanAfterSaleReceivedResponse(response string) (int, error) {
	decoder := json.NewDecoder(strings.NewReader(response))
	decoder.UseNumber()
	var payload interface{}
	if err := decoder.Decode(&payload); err != nil {
		return 0, fmt.Errorf("解析聚水潭实际收货响应失败: %v", err)
	}

	updates := extractJushuitanAfterSaleUpdates(payload)
	applied := 0
	for _, update := range updates {
		if update.ReturnID == "" && update.JushuitanAfterSaleID == "" && update.OrderID == "" {
			continue
		}
		if update.Status == "" {
			update.Status = ReturnOrderStatusReceived
		}
		if update.Response == "" {
			update.Response = response
		}
		if err := ApplyJushuitanAfterSaleUpdate(update); err != nil {
			return applied, err
		}
		applied++
	}
	return applied, nil
}

func extractJushuitanAfterSaleUpdates(value interface{}) []JushuitanAfterSaleUpdateInput {
	switch typed := value.(type) {
	case []interface{}:
		result := make([]JushuitanAfterSaleUpdateInput, 0, len(typed))
		for _, item := range typed {
			result = append(result, extractJushuitanAfterSaleUpdates(item)...)
		}
		return result
	case map[string]interface{}:
		if update, ok := jushuitanAfterSaleUpdateFromMap(typed); ok {
			return []JushuitanAfterSaleUpdateInput{update}
		}
		result := []JushuitanAfterSaleUpdateInput{}
		for _, item := range typed {
			result = append(result, extractJushuitanAfterSaleUpdates(item)...)
		}
		return result
	default:
		return nil
	}
}

func jushuitanAfterSaleUpdateFromMap(item map[string]interface{}) (JushuitanAfterSaleUpdateInput, bool) {
	update := JushuitanAfterSaleUpdateInput{
		ReturnID:             firstStringValue(item, "outer_as_id", "return_id", "return_order_id", "as_outer_id"),
		JushuitanAfterSaleID: firstStringValue(item, "as_id", "aftersale_id", "jushuitan_after_sale_id"),
		OrderID:              firstStringValue(item, "so_id", "order_id", "tid"),
		Status:               firstNonEmpty(firstStringValue(item, "status"), firstStringValue(item, "shop_status"), firstStringValue(item, "refund_status")),
	}
	if update.Status == "" && looksLikeReceivedAfterSale(item) {
		update.Status = ReturnOrderStatusReceived
	}
	raw, _ := json.Marshal(item)
	update.Response = string(raw)
	return update, update.ReturnID != "" || update.JushuitanAfterSaleID != "" || update.OrderID != ""
}

func looksLikeReceivedAfterSale(item map[string]interface{}) bool {
	for _, key := range []string{"received_date", "receive_date", "inout_date", "io_date", "wms_co_id"} {
		if firstStringValue(item, key) != "" {
			return true
		}
	}
	return false
}

func lockReturnOrderForJushuitanUpdate(tx *gorm.DB, input JushuitanAfterSaleUpdateInput) (*models.ReturnOrder, error) {
	var returnOrder models.ReturnOrder

	if input.ReturnID != "" {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("return_id = ?", input.ReturnID).First(&returnOrder).Error; err == nil {
			return &returnOrder, nil
		}
	}
	if input.JushuitanAfterSaleID != "" {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("jushuitan_after_sale_id = ?", input.JushuitanAfterSaleID).First(&returnOrder).Error; err == nil {
			return &returnOrder, nil
		}
	}
	if input.OrderID != "" {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", input.OrderID).Order("request_time DESC, id DESC").First(&returnOrder).Error; err == nil {
			return &returnOrder, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func NormalizeJushuitanAfterSaleStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "reject") || strings.Contains(status, "拒"):
		return ReturnOrderStatusRejected
	case strings.Contains(lower, "cancel") || strings.Contains(lower, "close") || strings.Contains(status, "取消") || strings.Contains(status, "关闭"):
		return ReturnOrderStatusCanceled
	case strings.Contains(lower, "received") || strings.Contains(lower, "stockin") || strings.Contains(status, "入库") || strings.Contains(status, "收货"):
		return ReturnOrderStatusReceived
	case strings.Contains(lower, "complete") || strings.Contains(lower, "finish") || strings.Contains(status, "完成") || strings.Contains(status, "退款成功"):
		return ReturnOrderStatusCompleted
	case strings.Contains(lower, "approve") || strings.Contains(lower, "agree") || strings.Contains(status, "审核通过") || strings.Contains(status, "同意"):
		return ReturnOrderStatusApproved
	default:
		return ReturnOrderStatusPending
	}
}

// UpdateReturnOrderStatus 更新退货订单状态
func UpdateReturnOrderStatus(returnOrderID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态更新时间
	switch status {
	case "completed":
		updates["completed_time"] = time.Now()
	case "canceled":
		updates["canceled_time"] = time.Now()
	}

	return db.DB.Model(&models.ReturnOrder{}).Where("return_id = ?", returnOrderID).Updates(updates).Error
}

// DeleteReturnOrder 删除退货订单
func DeleteReturnOrder(returnOrderID string) error {
	return db.DB.Where("return_id = ?", returnOrderID).Delete(&models.ReturnOrder{}).Error
}

// GetReturnOrderDetail 获取退货订单详情
func GetReturnOrderDetail(returnOrderID string) (*models.ReturnOrder, error) {
	var returnOrder models.ReturnOrder
	err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error
	if err != nil {
		return nil, err
	}
	return &returnOrder, nil
}
