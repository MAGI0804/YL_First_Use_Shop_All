package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
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
	ReturnOrderStatusCompleted    = "completed"
	ReturnOrderStatusCanceled     = "canceled"
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
	return result, err
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
