package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InventoryChangeOrderDeduct        = "order_deduct"         // 订单扣减库存
	InventoryChangeOrderCancelRestore = "order_cancel_restore" // 订单取消恢复库存
	InventoryChangeReturnRestore      = "return_restore"       // 退货恢复库存
	InventoryChangeManualAdjust       = "manual_adjust"        // 手动调整库存
	InventoryChangeSyncJushuitan      = "sync_jushuitan"       // 同步聚水潭库存
)

// ChangeInventoryInput 库存变动输入参数
type ChangeInventoryInput struct {
	CommodityID       string // 商品ID
	ChangeQty         int    // 变动数量，正数表示增加，负数表示减少
	ChangeType        string // 变动类型
	RelatedOrderID    string // 关联订单ID
	RelatedSubOrderID string // 关联子订单ID
	RelatedReturnID   string // 关联退货单ID
	OperatorID        string // 操作员ID
	Remark            string // 备注
	WarehouseCode     string // 仓库编码
}

// OrderInventoryItem 订单库存项
type OrderInventoryItem struct {
	CommodityID string // 商品ID
	Qty         int    // 数量
}

// ChangeInventory 变更库存（带事务）
func ChangeInventory(input ChangeInventoryInput) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		return changeInventoryTx(tx, input)
	})
}

// QueryInventory 查询库存信息
// 根据商品ID或款号查询库存信息
// 返回结果包含：商品信息、商品情况、款式数据（按商品ID查询）或款式数据、款式下所有商品及总库存（按款式编码查询）
func QueryInventory(commodityID, styleCode string) (map[string]any, error) {
	commodityID = strings.TrimSpace(commodityID)
	styleCode = strings.TrimSpace(styleCode)
	if commodityID == "" && styleCode == "" {
		return nil, fmt.Errorf("commodity_id或style_code不能为空")
	}

	result := map[string]any{}
	if commodityID != "" {
		var commodity models.Commodity
		if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
			return nil, err
		}

		var situation models.CommoditySituation
		_ = db.DB.Where("commodity_id = ?", commodityID).First(&situation).Error

		result["commodity"] = commodity
		result["commodity_situation"] = situation
		if commodity.StyleCode != "" {
			var styleCodeData models.StyleCodeData
			if err := db.DB.Where("style_code = ?", commodity.StyleCode).First(&styleCodeData).Error; err == nil {
				result["style_code_data"] = styleCodeData
			}
		}
		return result, nil
	}

	var styleCodeData models.StyleCodeData
	_ = db.DB.Where("style_code = ?", styleCode).First(&styleCodeData).Error

	var commodities []models.Commodity
	if err := db.DB.Where("style_code = ?", styleCode).Order("commodity_id ASC").Find(&commodities).Error; err != nil {
		return nil, err
	}

	totalInventory := 0
	for _, commodity := range commodities {
		totalInventory += commodity.Inventory
	}
	result["style_code"] = styleCode
	result["style_code_data"] = styleCodeData
	result["total_inventory"] = totalInventory
	result["commodities"] = commodities
	return result, nil
}

// AdjustInventory 手动调整库存
func AdjustInventory(input ChangeInventoryInput) error {
	input.ChangeType = InventoryChangeManualAdjust
	if strings.TrimSpace(input.Remark) == "" {
		input.Remark = "手动调整库存"
	}
	return ChangeInventory(input)
}

// QueryInventoryLogs 查询库存变更日志
// 支持按商品ID、款式编码、变动类型、关联订单/子订单/退货单等条件筛选
// 返回库存变动日志列表、总数量、当前页码、每页数量
func QueryInventoryLogs(input InventoryLogQueryInput) ([]models.InventoryLog, int64, int, int, error) {
	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := db.DB.Model(&models.InventoryLog{})
	if input.CommodityID != "" {
		query = query.Where("commodity_id = ?", strings.TrimSpace(input.CommodityID))
	}
	if input.StyleCode != "" {
		query = query.Where("style_code = ?", strings.TrimSpace(input.StyleCode))
	}
	if input.ChangeType != "" {
		query = query.Where("change_type = ?", strings.TrimSpace(input.ChangeType))
	}
	if input.RelatedOrderID != "" {
		query = query.Where("related_order_id = ?", strings.TrimSpace(input.RelatedOrderID))
	}
	if input.RelatedSubOrderID != "" {
		query = query.Where("related_sub_order_id = ?", strings.TrimSpace(input.RelatedSubOrderID))
	}
	if input.RelatedReturnID != "" {
		query = query.Where("related_return_id = ?", strings.TrimSpace(input.RelatedReturnID))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	var logs []models.InventoryLog
	if err := query.Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, page, pageSize, err
	}
	return logs, total, page, pageSize, nil
}

// QueryInventoryWarnings 查询库存预警商品
// 查询库存低于设定阈值的商品列表，用于及时补充库存
// 默认阈值为5，可自定义设置
func QueryInventoryWarnings(threshold, page, pageSize int) ([]models.Commodity, int64, int, int, int, error) {
	if threshold <= 0 {
		threshold = 5
	}
	page, pageSize = normalizePage(page, pageSize)

	query := db.DB.Model(&models.Commodity{}).Where("inventory <= ?", threshold)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, threshold, page, pageSize, err
	}

	var commodities []models.Commodity
	if err := query.Order("inventory ASC, commodity_id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&commodities).Error; err != nil {
		return nil, 0, threshold, page, pageSize, err
	}
	return commodities, total, threshold, page, pageSize, nil
}

// InventoryLogQueryInput 库存日志查询输入参数
type InventoryLogQueryInput struct {
	CommodityID       string // 商品ID
	StyleCode         string // 款号
	ChangeType        string // 变动类型
	RelatedOrderID    string // 关联订单ID
	RelatedSubOrderID string // 关联子订单ID
	RelatedReturnID   string // 关联退货单ID
	Page              int    // 页码
	PageSize          int    // 每页大小
}

// DeductInventoryForOrder 为订单扣减库存
// 在订单创建成功时调用，遍历所有子订单，扣减对应商品的库存
// 库存不足时返回错误，不执行扣减
func DeductInventoryForOrder(tx *gorm.DB, orderID string, subOrders []models.SubOrder) error {
	for _, subOrder := range subOrders {
		if subOrder.CommodityID == "" || subOrder.Qty <= 0 {
			return fmt.Errorf("子订单%s缺少商品或数量，无法扣库存", subOrder.SubOrderID)
		}
		if err := changeInventoryTx(tx, ChangeInventoryInput{
			CommodityID:       subOrder.CommodityID,
			ChangeQty:         -subOrder.Qty,
			ChangeType:        InventoryChangeOrderDeduct,
			RelatedOrderID:    orderID,
			RelatedSubOrderID: subOrder.SubOrderID,
			WarehouseCode:     subOrder.WmsCoID,
			Remark:            "订单创建成功扣库存",
		}); err != nil {
			return err
		}
	}
	return nil
}

// RestoreInventoryForOrderCancel 订单取消时恢复库存
// 当订单被取消时，遍历所有子订单，恢复之前扣减的库存
func RestoreInventoryForOrderCancel(tx *gorm.DB, orderID string, subOrders []models.SubOrder) error {
	for _, subOrder := range subOrders {
		if subOrder.CommodityID == "" || subOrder.Qty <= 0 {
			continue
		}
		if err := restoreSubOrderInventoryTx(tx, subOrder, InventoryChangeOrderCancelRestore, "", "订单取消回滚库存"); err != nil {
			return err
		}
	}
	return nil
}

// RestoreInventoryForReturn 退货完成时恢复库存
// 仅处理退货（return）类型，不处理仅退款（refund）类型
// 根据退货单关联的子订单恢复库存
func RestoreInventoryForReturn(tx *gorm.DB, returnOrder models.ReturnOrder) error {
	if returnOrder.Type == "refund" || returnOrder.SubOrderID == "" {
		return nil
	}

	var subOrder models.SubOrder
	if err := tx.Where("sub_order_id = ?", returnOrder.SubOrderID).First(&subOrder).Error; err != nil {
		return err
	}
	return restoreSubOrderInventoryTx(tx, subOrder, InventoryChangeReturnRestore, returnOrder.ReturnID, "售后完成回滚库存")
}

// ParseOrderInventoryItems 解析订单商品列表为库存项
// 支持两种格式：1.字符串数组，每个元素为商品ID，默认数量为1；2.对象数组，每个对象包含商品ID和数量
// 支持的商品ID字段：commodity_id、sku_id、product_id、id
// 支持的数量字段：qty、quantity、num
func ParseOrderInventoryItems(productList interface{}) ([]OrderInventoryItem, error) {
	items, ok := productList.([]interface{})
	if !ok {
		return nil, fmt.Errorf("product_list格式不正确")
	}

	result := make([]OrderInventoryItem, 0, len(items))
	for index, item := range items {
		switch value := item.(type) {
		case string:
			if strings.TrimSpace(value) == "" {
				return nil, fmt.Errorf("第%d个商品缺少commodity_id", index+1)
			}
			result = append(result, OrderInventoryItem{CommodityID: strings.TrimSpace(value), Qty: 1})
		case map[string]interface{}:
			commodityID := firstStringValue(value, "commodity_id", "sku_id", "product_id", "id")
			if commodityID == "" {
				return nil, fmt.Errorf("第%d个商品缺少commodity_id", index+1)
			}
			qty := firstIntValue(value, 1, "qty", "quantity", "num")
			if qty <= 0 {
				return nil, fmt.Errorf("第%d个商品数量必须大于0", index+1)
			}
			result = append(result, OrderInventoryItem{CommodityID: commodityID, Qty: qty})
		default:
			return nil, fmt.Errorf("第%d个商品格式不正确", index+1)
		}
	}
	return result, nil
}

// changeInventoryTx 库存变动事务处理（内部函数）
// 核心库存变更逻辑，包含以下步骤：
// 1. 检查参数合法性；2. 检查是否存在重复扣减；3. 使用行锁查询当前库存；
// 4. 计算变动后库存并检查是否充足；5. 更新商品库存；6. 更新商品情况库存；
// 7. 如果有款式编码则更新款式总库存；8. 记录库存变动日志
func changeInventoryTx(tx *gorm.DB, input ChangeInventoryInput) error {
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	if input.CommodityID == "" {
		return fmt.Errorf("commodity_id不能为空")
	}
	if input.ChangeQty == 0 {
		return fmt.Errorf("库存变动数量不能为0")
	}
	if input.ChangeType == "" {
		return fmt.Errorf("库存变动类型不能为空")
	}

	if input.RelatedSubOrderID != "" {
		var count int64
		if err := tx.Model(&models.InventoryLog{}).
			Where("commodity_id = ? AND change_type = ? AND related_sub_order_id = ?", input.CommodityID, input.ChangeType, input.RelatedSubOrderID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
	}

	var commodity models.Commodity
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("commodity_id = ?", input.CommodityID).
		First(&commodity).Error; err != nil {
		return err
	}

	beforeQty := commodity.Inventory
	afterQty := beforeQty + input.ChangeQty
	if afterQty < 0 {
		return fmt.Errorf("商品%s库存不足，当前库存%d，需要%d", input.CommodityID, beforeQty, -input.ChangeQty)
	}

	if err := tx.Model(&models.Commodity{}).
		Where("commodity_id = ?", input.CommodityID).
		Update("inventory", afterQty).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.CommoditySituation{}).
		Where("commodity_id = ?", input.CommodityID).
		Update("inventory", afterQty).Error; err != nil {
		return err
	}

	if commodity.StyleCode != "" {
		if err := refreshStyleCodeInventoryTx(tx, commodity.StyleCode); err != nil {
			return err
		}
	}

	log := models.InventoryLog{
		CommodityID:       input.CommodityID,
		StyleCode:         commodity.StyleCode,
		WarehouseCode:     input.WarehouseCode,
		BeforeQty:         beforeQty,
		ChangeQty:         input.ChangeQty,
		AfterQty:          afterQty,
		ChangeType:        input.ChangeType,
		RelatedOrderID:    input.RelatedOrderID,
		RelatedSubOrderID: input.RelatedSubOrderID,
		RelatedReturnID:   input.RelatedReturnID,
		OperatorID:        input.OperatorID,
		Remark:            input.Remark,
	}
	return tx.Create(&log).Error
}

// restoreSubOrderInventoryTx 恢复子订单库存（内部函数）
// 检查子订单是否有过扣减记录，如果有则恢复对应数量的库存
// 避免重复恢复：如果之前没有扣减记录，则不执行恢复操作
func restoreSubOrderInventoryTx(tx *gorm.DB, subOrder models.SubOrder, changeType, returnID, remark string) error {
	var deductedCount int64
	if err := tx.Model(&models.InventoryLog{}).
		Where("change_type = ? AND related_sub_order_id = ?", InventoryChangeOrderDeduct, subOrder.SubOrderID).
		Count(&deductedCount).Error; err != nil {
		return err
	}
	if deductedCount == 0 {
		return nil
	}

	return changeInventoryTx(tx, ChangeInventoryInput{
		CommodityID:       subOrder.CommodityID,
		ChangeQty:         subOrder.Qty,
		ChangeType:        changeType,
		RelatedOrderID:    subOrder.OrderID,
		RelatedSubOrderID: subOrder.SubOrderID,
		RelatedReturnID:   returnID,
		WarehouseCode:     subOrder.WmsCoID,
		Remark:            remark,
	})
}

// refreshStyleCodeInventoryTx 刷新款号总库存（内部函数）
// 统计指定款式编码下所有商品的库存总和，并更新到款式数据表中
func refreshStyleCodeInventoryTx(tx *gorm.DB, styleCode string) error {
	var total int64
	if err := tx.Model(&models.Commodity{}).
		Where("style_code = ?", styleCode).
		Select("COALESCE(SUM(inventory), 0)").
		Scan(&total).Error; err != nil {
		return err
	}
	return tx.Model(&models.StyleCodeData{}).
		Where("style_code = ?", styleCode).
		Update("inventory", int(total)).Error
}

// normalizePage 规范化分页参数
// 页码默认1，每页数量默认20，最大100
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// firstStringValue 从map中获取第一个存在的字符串值
func firstStringValue(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := m[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			if math.Trunc(v) == v {
				return strconv.FormatInt(int64(v), 10)
			}
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.Itoa(v)
		case json.Number:
			return v.String()
		}
	}
	return ""
}

// firstIntValue 从map中获取第一个存在的整数值
func firstIntValue(m map[string]interface{}, defaultVal int, keys ...string) int {
	for _, key := range keys {
		value, ok := m[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(v))
			if err == nil {
				return parsed
			}
		case json.Number:
			parsed, err := v.Int64()
			if err == nil {
				return int(parsed)
			}
		}
	}
	return defaultVal
}
