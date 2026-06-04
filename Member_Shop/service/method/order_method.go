package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ValidShopName = "youlan_kids"

var ValidOrderStatuses = []string{"pending", "unpaid", "paid", "partial_paid", "shipped", "delivered", "canceled", "processing", "returning", "exchanging"}
var validChangeStatusMap = map[string]bool{
	"pending":      true,
	"unpaid":       true,
	"paid":         true,
	"partial_paid": true,
	"processing":   true,
	"shipped":      true,
	"delivered":    true,
	"canceled":     true,
}

// ValidateShopName 验证店铺名称
func ValidateShopName(shopname string) bool {
	return shopname == ValidShopName
}

// ValidateOrderStatus 验证订单状态
func ValidateOrderStatus(status string, validStatuses []string) bool {
	if status == "" {
		return true
	}
	for _, validStatus := range validStatuses {
		if validStatus == status {
			return true
		}
	}
	return false
}

// ValidateChangeStatus 验证可变更的状态
func ValidateChangeStatus(status string) bool {
	return validChangeStatusMap[status]
}

// ConvertOrderToMap 将订单对象转换为Map
func ConvertOrderToMap(order models.Order) map[string]interface{} {
	return convertOrderToMap(order, nil)
}

func convertOrderToMap(order models.Order, afterSaleStatus *string) map[string]interface{} {
	result := make(map[string]interface{})
	result["order_id"] = order.OrderID
	result["user_id"] = order.UserID
	result["receiver_name"] = order.ReceiverName
	result["receiver_phone"] = order.ReceiverPhone
	result["province"] = order.Province
	result["city"] = order.City
	result["county"] = order.County
	result["detailed_address"] = order.DetailedAddress
	result["order_amount"] = order.OrderAmount
	result["final_pay_amount"] = normalizeFinalPayAmount(order)
	result["discount_amount"] = order.DiscountAmount
	result["discount_reason"] = order.DiscountReason
	result["status"] = order.Status
	result["pay_status"] = order.PayStatus
	result["payment_operator_id"] = order.PaymentOperatorID
	result["payment_remark"] = order.PaymentRemark
	result["price_adjusted_by"] = order.PriceAdjustedBy

	if order.ProductList != "" {
		var productList []string
		if err := json.Unmarshal([]byte(order.ProductList), &productList); err == nil {
			result["product_list"] = productList
		} else {
			result["product_list"] = []string{}
		}
	} else {
		result["product_list"] = []string{}
	}

	result["order_time"] = order.OrderTime.Format("2006-01-02 15:04:05")
	result["express_company"] = order.ExpressCompany
	result["express_number"] = order.ExpressNumber

	if !order.ShippedTime.IsZero() {
		result["shipped_time"] = order.ShippedTime.Format("2006-01-02 15:04:05")
	} else {
		result["shipped_time"] = ""
	}

	if !order.DeliveredTime.IsZero() {
		result["delivered_time"] = order.DeliveredTime.Format("2006-01-02 15:04:05")
	} else {
		result["delivered_time"] = ""
	}

	if !order.CanceledTime.IsZero() {
		result["canceled_time"] = order.CanceledTime.Format("2006-01-02 15:04:05")
	} else {
		result["canceled_time"] = ""
	}

	if !order.ProcessingTime.IsZero() {
		result["processing_time"] = order.ProcessingTime.Format("2006-01-02 15:04:05")
	} else {
		result["processing_time"] = ""
	}

	result["process_num"] = order.ProcessNum
	result["remarks"] = order.Remarks

	if order.LogisticsProcess != "" {
		var logisticsProcess []interface{}
		if err := json.Unmarshal([]byte(order.LogisticsProcess), &logisticsProcess); err == nil {
			result["logistics_process"] = logisticsProcess
		} else {
			result["logistics_process"] = []interface{}{}
		}
	} else {
		result["logistics_process"] = []interface{}{}
	}

	result["payment_method"] = order.PaymentMethod
	result["delivery_method"] = order.DeliveryMethod

	if !order.PaymentTime.IsZero() {
		result["payment_time"] = order.PaymentTime.Format("2006-01-02 15:04:05")
	} else {
		result["payment_time"] = ""
	}

	if order.SubOrderIDs != "" {
		var subOrderIDs []string
		if err := json.Unmarshal([]byte(order.SubOrderIDs), &subOrderIDs); err == nil {
			result["sub_order_ids"] = subOrderIDs
		} else {
			result["sub_order_ids"] = []string{}
		}
	} else {
		result["sub_order_ids"] = []string{}
	}

	result["jushuitan_order_id"] = order.JushuitanOrderID
	result["order_from"] = order.OrderFrom
	result["wms_co_id"] = order.WmsCoID
	result["lcid"] = order.LCID
	result["is_send_all"] = order.IsSendAll
	if afterSaleStatus != nil {
		setAfterSaleDisplayFields(result, *afterSaleStatus)
	} else {
		decorateOrderAfterSaleFields(result, order.OrderID)
	}

	return result
}

// ConvertOrdersToMap 将订单数组转换为Map数组
func ConvertOrdersToMap(orders []models.Order) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(orders))
	afterSaleStatuses := latestAfterSaleStatusByOrderIDs(extractOrderIDs(orders))
	for _, order := range orders {
		status := afterSaleStatuses[order.OrderID]
		result = append(result, convertOrderToMap(order, &status))
	}
	return result
}

func extractOrderIDs(orders []models.Order) []string {
	orderIDs := make([]string, 0, len(orders))
	for _, order := range orders {
		if order.OrderID != "" {
			orderIDs = append(orderIDs, order.OrderID)
		}
	}
	return orderIDs
}

// GenerateOrderNo 生成订单号
func GenerateOrderNo() string {
	currentDate := time.Now().Format("20060102")
	maxRetries := 5
	var orderID string

	for retry := 0; retry < maxRetries; retry++ {
		var randomNum string
		for i := 0; i < 8; i++ {
			randomNum += fmt.Sprintf("%d", time.Now().UnixNano()%10)
			time.Sleep(time.Nanosecond)
		}

		orderID = fmt.Sprintf("Y%s%s", currentDate, randomNum)

		var count int64
		err := db.DB.Model(&models.Order{}).Where("order_id = ?", orderID).Count(&count).Error
		if err == nil && count == 0 {
			return orderID
		}
	}

	return fmt.Sprintf("Y%s%d", currentDate, time.Now().UnixNano()%100000000)
}

// GenerateReturnOrderNo 生成退货单号
func GenerateReturnOrderNo() string {
	currentDate := time.Now().Format("20060102")
	maxRetries := 5
	var returnOrderID string

	for retry := 0; retry < maxRetries; retry++ {
		var randomNum string
		for i := 0; i < 8; i++ {
			randomNum += fmt.Sprintf("%d", time.Now().UnixNano()%10)
			time.Sleep(time.Nanosecond)
		}

		returnOrderID = fmt.Sprintf("T%s%s", currentDate, randomNum)

		var count int64
		err := db.DB.Model(&models.ReturnOrder{}).Where("return_id = ?", returnOrderID).Count(&count).Error
		if err == nil && count == 0 {
			return returnOrderID
		}
	}

	return fmt.Sprintf("T%s%d", currentDate, time.Now().UnixNano()%100000000)
}

// GenerateSubOrderNo 生成子订单号
func GenerateSubOrderNo() string {
	currentDate := time.Now().Format("20060102")
	maxRetries := 5
	var subOrderID string

	for retry := 0; retry < maxRetries; retry++ {
		var randomNum string
		for i := 0; i < 8; i++ {
			randomNum += fmt.Sprintf("%d", time.Now().UnixNano()%10)
			time.Sleep(time.Nanosecond)
		}

		subOrderID = fmt.Sprintf("S%s%s", currentDate, randomNum)

		var count int64
		err := db.DB.Model(&models.SubOrder{}).Where("sub_order_id = ?", subOrderID).Count(&count).Error
		if err == nil && count == 0 {
			return subOrderID
		}
	}

	return fmt.Sprintf("S%s%d", currentDate, time.Now().UnixNano()%100000000)
}

// CreateSubOrder 创建子订单
func CreateSubOrder(orderID, productName, productInfo string, subAmount float64) (*models.SubOrder, error) {
	return createSubOrderTx(db.DB, orderID, productName, productInfo, subAmount, "", 0)
}

func createSubOrderTx(tx *gorm.DB, orderID, productName, productInfo string, subAmount float64, commodityID string, qty int) (*models.SubOrder, error) {
	if qty <= 0 {
		qty = 1
	}

	subOrder := models.SubOrder{
		SubOrderID:  GenerateSubOrderNo(),
		OrderID:     orderID,
		ProductName: productName,
		ProductInfo: productInfo,
		SubAmount:   subAmount,
		Status:      "pending",
		PayStatus:   "unpaid",
		CommodityID: commodityID,
		Qty:         qty,
	}

	if err := tx.Create(&subOrder).Error; err != nil {
		return nil, err
	}

	return &subOrder, nil
}

// QueryOrdersResult 查询订单结果
type QueryOrdersResult struct {
	Orders []map[string]interface{}
	Total  int64
}

// QueryOrdersByUserID 根据用户ID查询订单
func QueryOrdersByUserID(userID int, status string, page, pageSize int) (*QueryOrdersResult, error) {
	var orders []models.Order
	query := db.DB.Table("order_data").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	offset := (page - 1) * pageSize

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("获取订单总数失败: %v", err)
		return nil, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("order_time DESC").Find(&orders).Error; err != nil {
		log.Printf("查询用户订单失败: %v", err)
		return nil, err
	}

	result := ConvertOrdersToMap(orders)

	return &QueryOrdersResult{
		Orders: result,
		Total:  total,
	}, nil
}

// GetOrderList 获取订单列表
func GetOrderList(userID int, status, beginTime, endTime, tid string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var orders []models.Order
	query := db.DB.Model(&models.Order{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if tid != "" {
		query = query.Where("order_id LIKE ?", "%"+tid+"%")
	}

	if beginTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", beginTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", beginTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour)
				beginTimeUTC := t.In(time.UTC)
				query = query.Where("order_time >= ?", beginTimeUTC)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			beginTimeUTC := t.In(time.UTC)
			query = query.Where("order_time >= ?", beginTimeUTC)
		}
	}

	if endTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", endTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour).Add(24 * time.Hour)
				endTimeUTC := t.In(time.UTC)
				query = query.Where("order_time < ?", endTimeUTC)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			endTimeUTC := t.In(time.UTC)
			query = query.Where("order_time < ?", endTimeUTC)
		}
	}

	offset := (page - 1) * pageSize

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("order_time DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	result := ConvertOrdersToMap(orders)
	for _, orderMap := range result {
		orderMap["logistics_process"] = []interface{}{}
	}

	return result, total, nil
}

// GetOrderDetail 获取订单详情
func GetOrderDetail(orderID string, userID int) (*models.Order, error) {
	query := db.DB.Model(&models.Order{}).Where("order_id = ?", orderID)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	var order models.Order
	if err := query.First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// ChangeStatusResult 变更状态结果
type ChangeStatusResult struct {
	Order              models.Order
	OldStatus          string
	UpdatedExpressInfo map[string]interface{}
}

// ChangeOrderStatus 变更订单状态
func ChangeOrderStatus(orderID, status, expressCompany, expressNumber string, logisticsProcess interface{}) (*ChangeStatusResult, error) {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	oldStatus := order.Status
	order.Status = status
	updatedExpressInfo := make(map[string]interface{})

	if expressCompany != "" {
		order.ExpressCompany = expressCompany
		updatedExpressInfo["express_company"] = expressCompany
	}

	if expressNumber != "" {
		order.ExpressNumber = expressNumber
		updatedExpressInfo["express_number"] = expressNumber
	}

	if logisticsProcess != nil {
		logisticsJSON, err := json.Marshal(logisticsProcess)
		if err != nil {
			return nil, err
		}
		order.LogisticsProcess = string(logisticsJSON)
		updatedExpressInfo["logistics_process"] = logisticsProcess
	}

	if err := db.DB.Select("status", "express_company", "express_number", "logistics_process").Save(&order).Error; err != nil {
		return nil, err
	}

	db.DB.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("status", status)

	log.Printf("订单状态变更: order_id=%s, 旧状态=%s, 新状态=%s", orderID, oldStatus, status)

	return &ChangeStatusResult{
		Order:              order,
		OldStatus:          oldStatus,
		UpdatedExpressInfo: updatedExpressInfo,
	}, nil
}

// CreateOrder 创建订单
func CreateOrder(userID int, receiverName, receiverPhone, province, city, county, detailedAddress string, orderAmount float64, productList interface{}, expressCompany, expressNumber, remark string) (*models.Order, error) {
	orderID := GenerateOrderNo()

	productListJSON, err := json.Marshal(productList)
	if err != nil {
		return nil, err
	}

	// 提取商品名称列表
	var productNames []string
	log.Printf("商品列表类型: %T, 长度: %d", productList, len(productList.([]interface{})))
	if productListArray, ok := productList.([]map[string]interface{}); ok {
		log.Printf("商品列表类型: []map[string]interface{}")
		for _, product := range productListArray {
			log.Printf("商品数据: %v", product)
			if productName, ok := product["product_name"].(string); ok {
				productNames = append(productNames, productName)
				log.Printf("提取到商品名称: %s (product_name)", productName)
			} else if productName, ok := product["name"].(string); ok {
				productNames = append(productNames, productName)
				log.Printf("提取到商品名称: %s (name)", productName)
			} else {
				log.Printf("商品数据中没有product_name或name字段: %v", product)
			}
		}
	} else if productListArray, ok := productList.([]map[string]string); ok {
		log.Printf("商品列表类型: []map[string]string")
		for _, product := range productListArray {
			log.Printf("商品数据: %v", product)
			if productName, ok := product["product_name"]; ok {
				productNames = append(productNames, productName)
				log.Printf("提取到商品名称: %s (product_name)", productName)
			} else if productName, ok := product["name"]; ok {
				productNames = append(productNames, productName)
				log.Printf("提取到商品名称: %s (name)", productName)
			} else {
				log.Printf("商品数据中没有product_name或name字段: %v", product)
			}
		}
	} else if productListArray, ok := productList.([]interface{}); ok {
		log.Printf("商品列表类型: []interface{}")
		for i, item := range productListArray {
			log.Printf("商品[%d]类型: %T, 数据: %v", i, item, item)
			if productMap, ok := item.(map[string]interface{}); ok {
				log.Printf("商品[%d]转换为map成功: %v", i, productMap)
				if productName, ok := productMap["product_name"].(string); ok {
					productNames = append(productNames, productName)
					log.Printf("提取到商品名称: %s (product_name)", productName)
				} else if productName, ok := productMap["name"].(string); ok {
					productNames = append(productNames, productName)
					log.Printf("提取到商品名称: %s (name)", productName)
				} else {
					log.Printf("商品数据中没有product_name或name字段: %v", productMap)
				}
			} else if productID, ok := item.(string); ok {
				// 如果商品数据是字符串，根据商品ID查询商品名称
				commodityName, err := GetCommodityNameByID(productID)
				if err != nil {
					// 如果查询失败，将商品ID作为商品名称
					productNames = append(productNames, productID)
					log.Printf("商品[%d]查询失败，使用商品ID作为名称: %s, 错误: %v", i, productID, err)
				} else {
					// 查询成功，使用商品名称
					productNames = append(productNames, commodityName)
					log.Printf("商品[%d]查询成功，商品名称: %s", i, commodityName)
				}
			} else {
				log.Printf("商品[%d]转换为map失败，类型: %T", i, item)
			}
		}
	} else {
		log.Printf("商品列表类型不支持: %T", productList)
	}

	productNamesJSON, err := json.Marshal(productNames)
	if err != nil {
		return nil, err
	}
	log.Printf("商品名称列表: %v, JSON: %s", productNames, string(productNamesJSON))

	order := models.Order{
		OrderID:         orderID,
		UserID:          userID,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		Province:        province,
		City:            city,
		County:          county,
		DetailedAddress: detailedAddress,
		OrderAmount:     orderAmount,
		ProductList:     string(productListJSON),
		ProdoctNameList: string(productNamesJSON),
		ExpressCompany:  expressCompany,
		ExpressNumber:   expressNumber,
		FinalPayAmount:  orderAmount,
		DiscountAmount:  0,
		DiscountReason:  "",
		Status:          "pending",
		PayStatus:       "unpaid",
		OrderTime:       time.Now(),
		Remarks:         remark,
	}

	inventoryItems, err := ParseOrderInventoryItems(productList)
	if err != nil {
		return nil, err
	}
	var subOrderIDs []string
	var createdSubOrders []models.SubOrder

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("order_id", "user_id", "receiver_name", "receiver_phone", "province", "city", "county", "detailed_address", "order_amount", "final_pay_amount", "discount_amount", "discount_reason", "product_list", "prodoct_name_list", "express_company", "express_number", "status", "pay_status", "order_time", "remarks").Create(&order).Error; err != nil {
			return err
		}
		if userID > 0 && orderAmount > 0 {
			if err := tx.Model(&models.Member{}).
				Where("user_id = ?", userID).
				Update("total_order_amount", gorm.Expr("total_order_amount + ?", orderAmount)).Error; err != nil {
				return err
			}
		}

		if productListArray, ok := productList.([]interface{}); ok {
			for index, item := range productListArray {
				var productName, productInfo string
				var subAmount float64 = 0.0

				if productMap, ok := item.(map[string]interface{}); ok {
					if name, ok := productMap["product_name"].(string); ok {
						productName = name
					} else if name, ok := productMap["name"].(string); ok {
						productName = name
					}
					productInfoJSON, _ := json.Marshal(productMap)
					productInfo = string(productInfoJSON)
					if amount, ok := productMap["price"].(float64); ok {
						subAmount = amount
					} else if amount, ok := productMap["sub_amount"].(float64); ok {
						subAmount = amount
					}
				} else if productID, ok := item.(string); ok {
					commodity, err := GetCommodityInfoByID(productID)
					if err != nil {
						productName = productID
						productInfo = productID
					} else {
						productName = commodity.Name
						productInfo = productID
						subAmount = commodity.Price
					}
				}

				if productName == "" && index < len(inventoryItems) {
					commodity, err := GetCommodityInfoByID(inventoryItems[index].CommodityID)
					if err == nil {
						productName = commodity.Name
						if subAmount == 0 {
							subAmount = commodity.Price
						}
					}
				}
				if productName == "" {
					return fmt.Errorf("第%d个商品缺少商品名称", index+1)
				}
				if index >= len(inventoryItems) {
					return fmt.Errorf("第%d个商品缺少库存扣减信息", index+1)
				}

				subOrder, err := createSubOrderTx(tx, orderID, productName, productInfo, subAmount, inventoryItems[index].CommodityID, inventoryItems[index].Qty)
				if err != nil {
					return err
				}
				createdSubOrders = append(createdSubOrders, *subOrder)
				subOrderIDs = append(subOrderIDs, subOrder.SubOrderID+":pending")
			}
		}

		if len(createdSubOrders) == 0 {
			return fmt.Errorf("订单缺少有效子订单")
		}
		if err := DeductInventoryForOrder(tx, orderID, createdSubOrders); err != nil {
			return err
		}
		subOrderIDsJSON, _ := json.Marshal(subOrderIDs)
		order.SubOrderIDs = string(subOrderIDsJSON)
		return tx.Model(&order).Update("sub_order_ids", order.SubOrderIDs).Error
	}); err != nil {
		return nil, err
	}

	return &order, nil
}

// CancelOrder 取消订单
func CancelOrder(orderID string) error {
	log.Printf("CancelOrder 开始执行, orderID: %s", orderID)

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).First(&order).Error; err != nil {
			log.Printf("CancelOrder 查询订单失败, orderID: %s, 错误: %v", orderID, err)
			return err
		}

		log.Printf("CancelOrder 查询到订单, orderID: %s, 当前状态: %s", orderID, order.Status)

		if order.Status != "pending" && order.Status != "paid" {
			log.Printf("CancelOrder 订单状态不允许取消, orderID: %s, 状态: %s", orderID, order.Status)
			return fmt.Errorf("订单状态不允许取消")
		}

		var subOrders []models.SubOrder
		if err := tx.Where("order_id = ?", orderID).Find(&subOrders).Error; err != nil {
			return err
		}
		if err := RestoreInventoryForOrderCancel(tx, orderID, subOrders); err != nil {
			return err
		}

		canceledTime := time.Now()
		log.Printf("CancelOrder 开始更新订单状态, orderID: %s", orderID)
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":        "canceled",
			"canceled_time": canceledTime,
		}).Error; err != nil {
			log.Printf("CancelOrder 更新订单失败, orderID: %s, 错误: %v", orderID, err)
			return err
		}

		log.Printf("CancelOrder 开始更新子订单状态, orderID: %s", orderID)
		if err := tx.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("status", "canceled").Error; err != nil {
			return err
		}
		canceledSubOrderIDs := make([]string, 0, len(subOrders))
		for _, subOrder := range subOrders {
			canceledSubOrderIDs = append(canceledSubOrderIDs, subOrder.SubOrderID+":canceled")
		}
		if len(canceledSubOrderIDs) > 0 {
			subOrderIDsJSON, _ := json.Marshal(canceledSubOrderIDs)
			if err := tx.Model(&models.Order{}).Where("order_id = ?", orderID).Update("sub_order_ids", string(subOrderIDsJSON)).Error; err != nil {
				return err
			}
		}
		log.Printf("CancelOrder 取消成功并回滚库存, orderID: %s", orderID)
		return nil
	})
}

// PayOrder 支付订单
func PayOrder(orderID string) error {
	return ConfirmOrderPayment(orderID, 0, "")

	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}
	if order.PayStatus == "paid" {
		return fmt.Errorf("订单已支付")
	}
	if order.Status != "delivered" {
		return fmt.Errorf("订单必须先送达才能支付")
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"pay_status":   "paid",
			"payment_time": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("pay_status", "paid").Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.Member{}).
		Where("user_id = ?", order.UserID).
		Update("total_paid_amount", gorm.Expr("total_paid_amount + ?", order.OrderAmount)).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// payOrderLegacy 旧版支付订单方法
func payOrderLegacy(orderID string) error {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	if order.Status != "delivered" {
		return fmt.Errorf("订单状态不允许支付")
	}

	order.PayStatus = "paid"
	order.PaymentTime = time.Now()
	if err := db.DB.Save(&order).Error; err != nil {
		return err
	}
	db.DB.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("pay_status", "paid")
	return nil
}

// DeliverOrder 发货订单
func DeliverOrder(orderID, expressCompany, expressNumber string) error {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	if order.Status != "pending" {
		return fmt.Errorf("订单状态不允许发货")
	}

	order.Status = "shipped"
	order.ExpressCompany = expressCompany
	order.ExpressNumber = expressNumber
	order.ShippedTime = time.Now()
	if err := db.DB.Select("status", "express_company", "express_number", "shipped_time").Save(&order).Error; err != nil {
		return err
	}
	db.DB.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("status", "shipped")
	return nil
}

// ReceiveOrder 签收订单
func ReceiveOrder(orderID string) error {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	if order.Status != "shipped" {
		return fmt.Errorf("订单状态不允许签收")
	}

	order.Status = "delivered"
	order.DeliveredTime = time.Now()
	if err := db.DB.Select("status", "delivered_time").Save(&order).Error; err != nil {
		return err
	}
	db.DB.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("status", "delivered")
	return nil
}

// ReturnOrderResult 退货结果
type ReturnOrderResult struct {
	ReturnOrder *models.ReturnOrder
	Order       *models.Order
	ReturnID    string
}

// RequestReturn 申请退货
func RequestReturn(userID int, orderID, OrderStatus, returnType, reason, SpecificReasons, buyerProvince, buyerCity, buyerCounty, buyerAddress, buyerPhone string, productIDs []string, orderProductList string) (*ReturnOrderResult, error) {
	// 旧接口只负责兼容请求格式，真实创建逻辑统一收敛到 CreateReturnOrderFromInput。
	return CreateReturnOrderFromInput(ReturnOrderCreateInput{
		UserID:          userID,
		OrderID:         orderID,
		OrderStatus:     OrderStatus,
		Type:            returnType,
		Reason:          reason,
		SpecificReasons: SpecificReasons,
		ProductIDs:      productIDs,
		ProductList:     orderProductList,
		BuyerProvince:   buyerProvince,
		BuyerCity:       buyerCity,
		BuyerCounty:     buyerCounty,
		BuyerAddress:    buyerAddress,
		BuyerPhone:      buyerPhone,
	})
}

// SyncLogisticsResult 同步物流结果
type SyncLogisticsResult struct {
	Order            *models.Order
	LogisticsProcess []map[string]interface{}
}

// SyncLogisticsInfo 同步物流信息
func SyncLogisticsInfo(orderID string) (*SyncLogisticsResult, error) {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	logisticsProcess := []map[string]interface{}{
		{"time": "2023-01-01 12:00:00", "location": "上海市", "description": "包裹已发出"},
		{"time": "2023-01-02 10:00:00", "location": "北京市", "description": "包裹已到达中转中心"},
		{"time": "2023-01-03 08:00:00", "location": "广州市", "description": "包裹已派送"},
	}

	logisticsJSON, err := json.Marshal(logisticsProcess)
	if err != nil {
		return nil, err
	}

	order.LogisticsProcess = string(logisticsJSON)

	if len(logisticsProcess) > 0 {
		lastStatus := logisticsProcess[len(logisticsProcess)-1]
		if desc, ok := lastStatus["description"].(string); ok {
			if strings.Contains(strings.ToLower(desc), "已签收") ||
				strings.Contains(strings.ToLower(desc), "已送达") {
				order.Status = "delivered"
			}
		}
	}

	// 构建更新字段映射
	updateData := map[string]interface{}{
		"logistics_process": order.LogisticsProcess,
	}

	// 如果状态发生变化，也更新状态字段
	if order.Status != "" {
		updateData["status"] = order.Status
	}

	if err := db.DB.Model(&order).Updates(updateData).Error; err != nil {
		return nil, err
	}

	return &SyncLogisticsResult{
		Order:            &order,
		LogisticsProcess: logisticsProcess,
	}, nil
}

// ChangeReceivingData 变更收货信息
func ChangeReceivingData(orderID, receiverName, receiverPhone, province, city, county, detailedAddress string) error {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return err
	}

	// 构建更新字段映射
	updateData := map[string]interface{}{
		"receiver_name":    receiverName,
		"province":         province,
		"city":             city,
		"county":           county,
		"detailed_address": detailedAddress,
	}

	// 只在receiverPhone不为空时更新该字段
	if receiverPhone != "" {
		updateData["receiver_phone"] = receiverPhone
	}

	// 使用Updates方法只更新必要的字段
	return db.DB.Model(&order).Updates(updateData).Error
}

// ReturnOrderDeliver 退货发货
func ReturnOrderDeliver(returnOrderID, expressCompany, expressNumber string) error {
	var returnOrder models.ReturnOrder
	if err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
		return err
	}

	if returnOrder.Status != ReturnOrderStatusApproved {
		return fmt.Errorf("退货订单状态不允许发货")
	}

	returnOrder.Status = ReturnOrderStatusBuyerShipped
	timeNow := time.Now()
	returnOrder.ShippedTime = &timeNow
	returnOrder.ExpressCompany = expressCompany
	returnOrder.ExpressNumber = expressNumber

	return db.DB.Save(&returnOrder).Error
}

// ReturnOrderReceive 退货签收
func ReturnOrderReceive(returnOrderID string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var returnOrder models.ReturnOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
			return err
		}

		if returnOrder.Status != ReturnOrderStatusBuyerShipped &&
			returnOrder.Status != ReturnOrderStatusReceived &&
			returnOrder.Status != "shipped" &&
			!(returnOrder.Type == "refund" && returnOrder.Status == ReturnOrderStatusApproved) {
			return fmt.Errorf("退货订单状态不允许签收")
		}

		// ERP 入库推送会先把售后置为 received 并回滚库存；
		// 旧接口仍兼容 buyer_shipped/shipped 直接完成的场景。
		if returnOrder.Status != ReturnOrderStatusReceived {
			if err := RestoreInventoryForReturn(tx, returnOrder); err != nil {
				return err
			}
		}

		returnOrder.Status = "completed"
		timeNow := time.Now()
		returnOrder.CompletedTime = &timeNow

		return tx.Save(&returnOrder).Error
	})
}

// ReturnOrderCancel 取消退货
func ReturnOrderCancel(returnOrderID, reason string) error {
	var returnOrder models.ReturnOrder
	if err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
		return err
	}

	if returnOrder.Status != ReturnOrderStatusPending && returnOrder.Status != ReturnOrderStatusApproved {
		return fmt.Errorf("退货订单状态不允许取消")
	}

	returnOrder.Status = "canceled"
	timeNow := time.Now()
	returnOrder.CanceledTime = &timeNow
	if returnOrder.Remarks != "" {
		returnOrder.Remarks += "\n取消原因: " + reason
	} else {
		returnOrder.Remarks = "取消原因: " + reason
	}

	return db.DB.Save(&returnOrder).Error
}

// ReturnOrderUpdateBuyerInfo 更新退货买家信息
func ReturnOrderUpdateBuyerInfo(returnOrderID, buyerProvince, buyerCity, buyerCounty, buyerAddress, buyerPhone string) error {
	var returnOrder models.ReturnOrder
	if err := db.DB.Where("return_id = ?", returnOrderID).First(&returnOrder).Error; err != nil {
		return err
	}

	if returnOrder.Status == ReturnOrderStatusBuyerShipped || returnOrder.Status == "shipped" ||
		returnOrder.Status == ReturnOrderStatusCompleted || returnOrder.Status == ReturnOrderStatusCanceled {
		return fmt.Errorf("退货订单状态不允许修改买家信息")
	}

	returnOrder.BuyerProvince = buyerProvince
	returnOrder.BuyerCity = buyerCity
	returnOrder.BuyerCounty = buyerCounty
	returnOrder.BuyerAddress = buyerAddress
	returnOrder.BuyerPhone = buyerPhone

	return db.DB.Save(&returnOrder).Error
}

// BatchOrdersQueryResult 批量查询订单结果
type BatchOrdersQueryResult struct {
	Orders []map[string]interface{}
	Total  int64
}

// BatchOrdersQuery 批量查询订单
func BatchOrdersQuery(userID int, status, beginTime, endTime string, page, pageSize int) (*BatchOrdersQueryResult, error) {
	var orders []models.Order
	query := db.DB.Table("order_data")
	query = query.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if beginTime != "" {
		t, err := time.Parse("2006-01-02", beginTime)
		if err != nil {
			return nil, err
		}
		query = query.Where("order_time >= ?", t.Add(-8*time.Hour))
	}

	if endTime != "" {
		t, err := time.Parse("2006-01-02", endTime)
		if err != nil {
			return nil, err
		}
		query = query.Where("order_time < ?", t.Add(-8*time.Hour+24*time.Hour))
	}

	offset := (page - 1) * pageSize

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("order_time DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	result := ConvertOrdersToMap(orders)

	return &BatchOrdersQueryResult{
		Orders: result,
		Total:  total,
	}, nil
}

// GetOrderByID 根据ID获取订单
func GetOrderByID(orderID string) (*models.Order, error) {
	var order models.Order
	if err := db.DB.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新订单
func UpdateOrder(order *models.Order, fields []string) error {
	return db.DB.Select(fields).Save(order).Error
}

// SearchOrdersByProductName 根据商品名称搜索订单
func SearchOrdersByProductName(userID int, productName, status, beginTime, endTime, tid string, page, pageSize int) (*BatchOrdersQueryResult, error) {
	var orders []models.Order
	query := db.DB.Table("order_data")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if tid != "" {
		query = query.Where("order_id LIKE ?", "%"+tid+"%")
	}

	// 按商品名称模糊搜索
	if productName != "" {
		query = query.Where("prodoct_name_list LIKE ?", "%"+productName+"%")
	}

	if beginTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", beginTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", beginTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour)
				beginTimeUTC := t.In(time.UTC)
				query = query.Where("order_time >= ?", beginTimeUTC)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			beginTimeUTC := t.In(time.UTC)
			query = query.Where("order_time >= ?", beginTimeUTC)
		}
	}

	if endTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", endTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour).Add(24 * time.Hour)
				endTimeUTC := t.In(time.UTC)
				query = query.Where("order_time < ?", endTimeUTC)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			endTimeUTC := t.In(time.UTC)
			query = query.Where("order_time < ?", endTimeUTC)
		}
	}

	offset := (page - 1) * pageSize

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("order_time DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	result := ConvertOrdersToMap(orders)

	return &BatchOrdersQueryResult{
		Orders: result,
		Total:  total,
	}, nil
}

// ParsePhoneNumber 解析手机号
func ParsePhoneNumber(phone interface{}) (string, error) {
	if phone == nil {
		return "", fmt.Errorf("手机号不能为空")
	}
	switch v := phone.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64), nil
	default:
		return "", fmt.Errorf("手机号格式不正确")
	}
}

// ConvertSubOrderToMap 将子订单对象转换为Map
func ConvertSubOrderToMap(subOrder models.SubOrder) map[string]interface{} {
	result := make(map[string]interface{})
	result["sub_order_id"] = subOrder.SubOrderID
	result["order_id"] = subOrder.OrderID
	result["product_info"] = subOrder.ProductInfo
	result["product_name"] = subOrder.ProductName
	result["status"] = subOrder.Status
	result["pay_status"] = subOrder.PayStatus
	result["sub_amount"] = subOrder.SubAmount
	result["express_company"] = subOrder.ExpressCompany
	result["express_number"] = subOrder.ExpressNumber
	result["is_shipped"] = subOrder.IsShipped
	result["qty"] = subOrder.Qty
	result["commodity_id"] = subOrder.CommodityID
	result["create_time"] = subOrder.CreateTime.Format("2006-01-02 15:04:05")
	result["update_time"] = subOrder.UpdateTime.Format("2006-01-02 15:04:05")
	result["jushuitan_sub_order_id"] = subOrder.JushuitanSubOrderID
	result["wms_co_id"] = subOrder.WmsCoID
	decorateSubOrderAfterSaleFields(result, subOrder.SubOrderID)

	if subOrder.ShippedTime != nil {
		result["shipped_time"] = subOrder.ShippedTime.Format("2006-01-02 15:04:05")
	} else {
		result["shipped_time"] = ""
	}

	return result
}

// ConvertSubOrdersToMap 将子订单数组转换为Map数组
func ConvertSubOrdersToMap(subOrders []models.SubOrder) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(subOrders))
	for _, subOrder := range subOrders {
		result = append(result, ConvertSubOrderToMap(subOrder))
	}
	return result
}

// GetSubOrdersByOrderID 根据订单ID获取子订单
func GetSubOrdersByOrderID(orderID string) ([]models.SubOrder, error) {
	var subOrders []models.SubOrder
	if err := db.DB.Where("order_id = ?", orderID).Find(&subOrders).Error; err != nil {
		return nil, err
	}
	return subOrders, nil
}

// GetSubOrderByID 根据ID获取子订单
func GetSubOrderByID(subOrderID string) (*models.SubOrder, error) {
	var subOrder models.SubOrder
	if err := db.DB.Where("sub_order_id = ?", subOrderID).First(&subOrder).Error; err != nil {
		return nil, err
	}
	return &subOrder, nil
}

// ChangeSubOrderStatus 变更子订单状态
func ChangeSubOrderStatus(subOrderID, status string) error {
	return db.DB.Model(&models.SubOrder{}).Where("sub_order_id = ?", subOrderID).Update("status", status).Error
}

// UpdateSubOrderShipInfo 更新子订单发货信息
func UpdateSubOrderShipInfo(subOrderID, expressNumber, expressCompany, sendDate string) error {
	shippedTime := time.Now()
	if sendDate != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", sendDate); err == nil {
			shippedTime = t
		} else if t, err := time.Parse("2006-01-02 15:04:05.000", sendDate); err == nil {
			shippedTime = t
		}
	}

	updates := map[string]interface{}{
		"express_number":  expressNumber,
		"express_company": expressCompany,
		"is_shipped":      "yes",
		"shipped_time":    shippedTime,
	}

	return db.DB.Model(&models.SubOrder{}).Where("sub_order_id = ?", subOrderID).Updates(updates).Error
}

// CancelSubOrder 取消子订单
func CancelSubOrder(subOrderID, reason string) error {
	var orderID string
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		var subOrder models.SubOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sub_order_id = ?", subOrderID).First(&subOrder).Error; err != nil {
			return err
		}
		orderID = subOrder.OrderID

		if subOrder.Status != "pending" && subOrder.Status != "paid" {
			return fmt.Errorf("子订单状态不允许取消")
		}
		if subOrder.IsShipped == "yes" {
			return fmt.Errorf("已发货子订单不能直接取消，请走售后流程")
		}
		if err := restoreSubOrderInventoryTx(tx, subOrder, InventoryChangeOrderCancelRestore, "", "子订单取消回滚库存"); err != nil {
			return err
		}
		return tx.Model(&subOrder).Update("status", "canceled").Error
	}); err != nil {
		return err
	}

	UpdateMainOrderSubOrderStatus(orderID)
	return nil
}

// UpdateMainOrderSubOrderStatus 更新主订单的子订单状态
func UpdateMainOrderSubOrderStatus(orderID string) {
	subOrders, err := GetSubOrdersByOrderID(orderID)
	if err != nil || len(subOrders) == 0 {
		return
	}

	var subOrderIDs []string
	for _, so := range subOrders {
		subOrderIDs = append(subOrderIDs, so.SubOrderID+":"+so.Status)
	}

	subOrderIDsJSON, _ := json.Marshal(subOrderIDs)
	db.DB.Model(&models.Order{}).Where("order_id = ?", orderID).Update("sub_order_ids", string(subOrderIDsJSON))
}

// ReturnSubOrder 子订单退货
func ReturnSubOrder(subOrderID, reason, specificReasons, buyerProvince, buyerCity, buyerCounty, buyerAddress, buyerPhone string) error {
	subOrder, err := GetSubOrderByID(subOrderID)
	if err != nil {
		return err
	}

	if subOrder.Status != "pending" && subOrder.Status != "paid" && subOrder.Status != "shipped" {
		return fmt.Errorf("子订单状态不允许退货")
	}

	_, err = CreateReturnOrderFromInput(ReturnOrderCreateInput{
		UserID:          subOrderIDUserFallback(subOrder),
		OrderID:         subOrder.OrderID,
		SubOrderID:      subOrderID,
		OrderStatus:     subOrder.Status,
		Type:            "return",
		Reason:          reason,
		SpecificReasons: specificReasons,
		ProductList:     subOrder.ProductInfo,
		BuyerProvince:   buyerProvince,
		BuyerCity:       buyerCity,
		BuyerCounty:     buyerCounty,
		BuyerAddress:    buyerAddress,
		BuyerPhone:      buyerPhone,
	})
	return err
}

func subOrderIDUserFallback(subOrder *models.SubOrder) int {
	var order models.Order
	if err := db.DB.Select("user_id").Where("order_id = ?", subOrder.OrderID).First(&order).Error; err != nil {
		return 0
	}
	return order.UserID
}

func normalizeFinalPayAmount(order models.Order) float64 {
	if order.FinalPayAmount > 0 {
		return order.FinalPayAmount
	}
	if order.OrderAmount > 0 {
		return order.OrderAmount
	}
	return 0
}

func ValidatePaymentAdjustment(orderAmount, finalPayAmount float64, discountReason string) (float64, error) {
	if finalPayAmount < 0 {
		return 0, fmt.Errorf("final pay amount cannot be negative")
	}
	if finalPayAmount > orderAmount {
		return 0, fmt.Errorf("final pay amount cannot exceed order amount")
	}
	discountAmount := orderAmount - finalPayAmount
	if discountAmount > 0 && strings.TrimSpace(discountReason) == "" {
		return 0, fmt.Errorf("discount reason is required")
	}
	return discountAmount, nil
}

func ValidateOrderReadyToPay(status string) error {
	if status != "delivered" {
		return fmt.Errorf("order must be delivered before payment")
	}
	return nil
}

func UpdatePaymentAmount(orderID string, finalPayAmount float64, discountReason string, operatorID int) (*models.Order, error) {
	var updatedOrder models.Order
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		if order.Status == "canceled" {
			return fmt.Errorf("order is canceled")
		}
		if order.PayStatus == "paid" {
			return fmt.Errorf("order already paid")
		}
		discountAmount, err := ValidatePaymentAdjustment(order.OrderAmount, finalPayAmount, discountReason)
		if err != nil {
			return err
		}
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"final_pay_amount":  finalPayAmount,
			"discount_amount":   discountAmount,
			"discount_reason":   strings.TrimSpace(discountReason),
			"price_adjusted_by": operatorID,
			"price_adjusted_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		return tx.Where("order_id = ?", orderID).First(&updatedOrder).Error
	})
	if err != nil {
		return nil, err
	}
	return &updatedOrder, nil
}

func ConfirmOrderPayment(orderID string, operatorID int, paymentRemark string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		if order.PayStatus == "paid" {
			return fmt.Errorf("order already paid")
		}
		if order.Status == "canceled" {
			return fmt.Errorf("order status does not allow payment")
		}
		if err := ValidateOrderReadyToPay(order.Status); err != nil {
			return err
		}
		finalPayAmount := normalizeFinalPayAmount(order)
		if _, err := ValidatePaymentAdjustment(order.OrderAmount, finalPayAmount, order.DiscountReason); err != nil {
			return err
		}
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"final_pay_amount":    finalPayAmount,
			"pay_status":          "paid",
			"payment_time":        time.Now(),
			"payment_operator_id": operatorID,
			"payment_remark":      paymentRemark,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SubOrder{}).Where("order_id = ?", orderID).Update("pay_status", "paid").Error; err != nil {
			return err
		}
		if finalPayAmount > 0 && order.UserID > 0 {
			return tx.Model(&models.Member{}).
				Where("user_id = ?", order.UserID).
				Update("total_paid_amount", gorm.Expr("total_paid_amount + ?", finalPayAmount)).Error
		}
		return nil
	})
}
