package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InventoryChangeOrderDeduct        = "order_create_deduct"
	InventoryChangeOrderCancelRestore = "order_cancel_restore"
	InventoryChangeReturnRestore      = "return_completed_restore"
	InventoryChangeManualAdjust       = "manual_adjust"
	InventoryChangeSyncJushuitan      = "jushuitan_sync"
	InventoryChangeStockTransfer      = "stock_transfer"
	InventoryChangeStockCheck         = "stock_check"

	legacyInventoryChangeOrderDeduct = "order_deduct"
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

type InventoryTransferInput struct {
	CommodityID         string
	Qty                 int
	SourceWarehouseCode string
	TargetWarehouseCode string
	OperatorID          string
	Remark              string
}

type InventoryStockCheckInput struct {
	CommodityID   string
	ActualQty     int
	WarehouseCode string
	OperatorID    string
	Remark        string
}

type JushuitanInventorySyncInput struct {
	SkuID         string
	IID           string
	Name          string
	Qty           int
	VirtualQty    int
	OrderLock     int
	PickLock      int
	Modified      string
	WarehouseCode string
}

type JushuitanInventorySyncResult struct {
	CommodityID string `json:"commodity_id"`
	SkuID       string `json:"sku_id"`
	BeforeQty   int    `json:"before_qty"`
	AfterQty    int    `json:"after_qty"`
	ChangeQty   int    `json:"change_qty"`
	Modified    string `json:"modified"`
	Skipped     bool   `json:"skipped"`
	SkipReason  string `json:"skip_reason"`
}

// ChangeInventory 变更库存（带事务）
func ChangeInventory(input ChangeInventoryInput) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		return changeInventoryTx(tx, input)
	})
}

func ApplyJushuitanInventorySync(input JushuitanInventorySyncInput) (*JushuitanInventorySyncResult, error) {
	input.SkuID = strings.TrimSpace(input.SkuID)
	input.IID = strings.TrimSpace(input.IID)
	input.Modified = strings.TrimSpace(input.Modified)
	input.WarehouseCode = strings.TrimSpace(input.WarehouseCode)
	if input.SkuID == "" && input.IID == "" {
		return nil, fmt.Errorf("sku_id或i_id不能为空")
	}

	var result *JushuitanInventorySyncResult
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var commodity models.Commodity
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"})
		if input.SkuID != "" && input.IID != "" {
			query = query.Where("commodity_id = ? OR spec_code = ? OR style_code = ?", input.SkuID, input.SkuID, input.IID)
		} else if input.SkuID != "" {
			query = query.Where("commodity_id = ? OR spec_code = ?", input.SkuID, input.SkuID)
		} else {
			query = query.Where("style_code = ?", input.IID)
		}
		if err := query.First(&commodity).Error; err != nil {
			return err
		}

		if input.Modified != "" {
			var count int64
			if err := tx.Model(&models.InventoryLog{}).
				Where("commodity_id = ? AND change_type = ? AND related_sub_order_id = ?", commodity.CommodityID, InventoryChangeSyncJushuitan, input.Modified).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				result = &JushuitanInventorySyncResult{
					CommodityID: commodity.CommodityID,
					SkuID:       input.SkuID,
					BeforeQty:   commodity.Inventory,
					AfterQty:    commodity.Inventory,
					Modified:    input.Modified,
					Skipped:     true,
					SkipReason:  "duplicate_modified",
				}
				return nil
			}
		}

		afterQty := calculateJushuitanAvailableQty(input.Qty, input.OrderLock, input.VirtualQty)
		remark := fmt.Sprintf("聚水潭库存同步 sku_id=%s i_id=%s qty=%d order_lock=%d virtual_qty=%d pick_lock=%d modified=%s",
			input.SkuID, input.IID, input.Qty, input.OrderLock, input.VirtualQty, input.PickLock, input.Modified)
		idempotencyKey := ""
		if input.Modified != "" {
			idempotencyKey = fmt.Sprintf("inventory:%s:%s:modified:%s", InventoryChangeSyncJushuitan, commodity.CommodityID, input.Modified)
		}
		openChange, err := setOpenInventoryAvailableTx(
			tx,
			commodity,
			input.WarehouseCode,
			afterQty,
			InventoryChangeSyncJushuitan,
			"jushuitan",
			input.Modified,
			idempotencyKey,
			"jushuitan",
			remark,
		)
		if err != nil {
			return err
		}
		if openChange == nil {
			result = &JushuitanInventorySyncResult{
				CommodityID: commodity.CommodityID,
				SkuID:       input.SkuID,
				BeforeQty:   commodity.Inventory,
				AfterQty:    commodity.Inventory,
				Modified:    input.Modified,
				Skipped:     true,
				SkipReason:  "duplicate_modified",
			}
			return nil
		}
		beforeQty := openChange.BeforeQty
		afterQty = openChange.AfterQty
		changeQty := afterQty - beforeQty
		result = &JushuitanInventorySyncResult{
			CommodityID: commodity.CommodityID,
			SkuID:       input.SkuID,
			BeforeQty:   beforeQty,
			AfterQty:    afterQty,
			ChangeQty:   changeQty,
			Modified:    input.Modified,
		}
		if changeQty == 0 {
			result.Skipped = true
			result.SkipReason = "inventory_unchanged"
			return nil
		}

		if _, err := refreshLegacyInventoryFromOpenTx(tx, commodity.CommodityID, commodity.StyleCode); err != nil {
			return err
		}

		log := models.InventoryLog{
			CommodityID:       commodity.CommodityID,
			StyleCode:         commodity.StyleCode,
			WarehouseCode:     openChange.WarehouseCode,
			BeforeQty:         beforeQty,
			ChangeQty:         changeQty,
			AfterQty:          afterQty,
			ChangeType:        InventoryChangeSyncJushuitan,
			RelatedSubOrderID: input.Modified,
			OperatorID:        "jushuitan",
			Remark:            remark,
		}
		return tx.Create(&log).Error
	})
	return result, err
}

func calculateJushuitanAvailableQty(qty, orderLock, virtualQty int) int {
	available := qty - orderLock + virtualQty
	if available < 0 {
		return 0
	}
	return available
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

		openInventory, err := QueryOpenInventory(OpenInventoryQueryInput{CommodityID: commodityID})
		if err != nil {
			return nil, err
		}
		if len(openInventory.Items) > 0 {
			commodity.Inventory = openInventory.Summary.TotalAvailableQty
			situation.Inventory = openInventory.Summary.TotalAvailableQty
		}

		result["commodity"] = commodity
		result["commodity_situation"] = situation
		result["open_inventory"] = openInventory
		if commodity.StyleCode != "" {
			var styleCodeData models.StyleCodeData
			if err := db.DB.Where("style_code = ?", commodity.StyleCode).First(&styleCodeData).Error; err == nil {
				if styleOpenInventory, err := QueryOpenInventory(OpenInventoryQueryInput{StyleCode: commodity.StyleCode}); err == nil && len(styleOpenInventory.Items) > 0 {
					styleCodeData.Inventory = styleOpenInventory.Summary.TotalAvailableQty
				}
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

	openInventory, err := QueryOpenInventory(OpenInventoryQueryInput{StyleCode: styleCode})
	if err != nil {
		return nil, err
	}
	availableByCommodity := openInventoryAvailableByCommodity(openInventory.Items)
	totalInventory := 0
	if len(openInventory.Items) > 0 {
		totalInventory = openInventory.Summary.TotalAvailableQty
		styleCodeData.Inventory = totalInventory
		for i := range commodities {
			if availableQty, ok := availableByCommodity[commodities[i].CommodityID]; ok {
				commodities[i].Inventory = availableQty
			}
		}
	} else {
		for _, commodity := range commodities {
			totalInventory += commodity.Inventory
		}
	}
	result["style_code"] = styleCode
	result["style_code_data"] = styleCodeData
	result["total_inventory"] = totalInventory
	result["commodities"] = commodities
	result["open_inventory"] = openInventory
	return result, nil
}

func openInventoryAvailableByCommodity(items []OpenInventoryBalanceView) map[string]int {
	availableByCommodity := make(map[string]int)
	for _, item := range items {
		availableByCommodity[item.CommodityID] += item.AvailableQty
	}
	return availableByCommodity
}

// AdjustInventory 手动调整库存
func AdjustInventory(input ChangeInventoryInput) error {
	input.ChangeType = InventoryChangeManualAdjust
	if strings.TrimSpace(input.Remark) == "" {
		input.Remark = "手动调整库存"
	}
	return ChangeInventory(input)
}

func TransferInventory(input InventoryTransferInput) error {
	if err := validateInventoryTransferInput(&input); err != nil {
		return err
	}

	remark := strings.TrimSpace(input.Remark)
	if remark == "" {
		remark = fmt.Sprintf("stock transfer %s -> %s", input.SourceWarehouseCode, input.TargetWarehouseCode)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var commodity models.Commodity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("commodity_id = ?", input.CommodityID).
			First(&commodity).Error; err != nil {
			return err
		}
		if err := ensureOpenInventoryWarehouseCodeTx(tx, input.SourceWarehouseCode); err != nil {
			return err
		}
		if err := ensureOpenInventoryWarehouseCodeTx(tx, input.TargetWarehouseCode); err != nil {
			return err
		}
		if err := ensureOpenInventorySKUTx(tx, commodity); err != nil {
			return err
		}

		sourceInitialQty := 0
		if input.SourceWarehouseCode == models.DefaultInventoryWarehouseCode {
			sourceInitialQty = commodity.Inventory
		}
		sourceBalance, err := lockOrCreateOpenInventoryBalanceTx(tx, commodity, input.SourceWarehouseCode, sourceInitialQty)
		if err != nil {
			return err
		}
		if sourceBalance.AvailableQty < input.Qty {
			return fmt.Errorf("商品%s源仓库存不足，当前库存%d，需要%d", input.CommodityID, sourceBalance.AvailableQty, input.Qty)
		}
		targetBalance, err := lockOrCreateOpenInventoryBalanceTx(tx, commodity, input.TargetWarehouseCode, 0)
		if err != nil {
			return err
		}

		sourceAfter := sourceBalance.AvailableQty - input.Qty
		targetAfter := targetBalance.AvailableQty + input.Qty
		if err := updateOpenInventoryBalanceQtyTx(tx, input.CommodityID, input.SourceWarehouseCode, sourceAfter); err != nil {
			return err
		}
		if err := updateOpenInventoryBalanceQtyTx(tx, input.CommodityID, input.TargetWarehouseCode, targetAfter); err != nil {
			return err
		}

		transferKey := fmt.Sprintf("inventory:%s:%s:%s:%s:%d", InventoryChangeStockTransfer, input.CommodityID, input.SourceWarehouseCode, input.TargetWarehouseCode, time.Now().UnixNano())
		if err := createInventoryStockMovementTx(tx, commodity, input.SourceWarehouseCode, InventoryChangeStockTransfer, "inventory", "transfer", transferKey+":out", sourceBalance.AvailableQty, -input.Qty, sourceAfter, input.OperatorID, remark+" out"); err != nil {
			return err
		}
		if err := createInventoryStockMovementTx(tx, commodity, input.TargetWarehouseCode, InventoryChangeStockTransfer, "inventory", "transfer", transferKey+":in", targetBalance.AvailableQty, input.Qty, targetAfter, input.OperatorID, remark+" in"); err != nil {
			return err
		}

		if _, err := refreshLegacyInventoryFromOpenTx(tx, input.CommodityID, commodity.StyleCode); err != nil {
			return err
		}
		if err := tx.Create(&models.InventoryLog{
			CommodityID:   input.CommodityID,
			StyleCode:     commodity.StyleCode,
			WarehouseCode: input.SourceWarehouseCode,
			BeforeQty:     sourceBalance.AvailableQty,
			ChangeQty:     -input.Qty,
			AfterQty:      sourceAfter,
			ChangeType:    InventoryChangeStockTransfer,
			OperatorID:    input.OperatorID,
			Remark:        remark + " out",
		}).Error; err != nil {
			return err
		}
		return tx.Create(&models.InventoryLog{
			CommodityID:   input.CommodityID,
			StyleCode:     commodity.StyleCode,
			WarehouseCode: input.TargetWarehouseCode,
			BeforeQty:     targetBalance.AvailableQty,
			ChangeQty:     input.Qty,
			AfterQty:      targetAfter,
			ChangeType:    InventoryChangeStockTransfer,
			OperatorID:    input.OperatorID,
			Remark:        remark + " in",
		}).Error
	})
}

func StockCheckInventory(input InventoryStockCheckInput) (map[string]any, error) {
	if err := validateInventoryStockCheckInput(&input); err != nil {
		return nil, err
	}

	var result map[string]any
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var commodity models.Commodity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("commodity_id = ?", input.CommodityID).
			First(&commodity).Error; err != nil {
			return err
		}

		if err := ensureOpenInventoryWarehouseCodeTx(tx, input.WarehouseCode); err != nil {
			return err
		}
		if err := ensureOpenInventorySKUTx(tx, commodity); err != nil {
			return err
		}
		initialQty := 0
		if input.WarehouseCode == models.DefaultInventoryWarehouseCode {
			initialQty = commodity.Inventory
		}
		balance, err := lockOrCreateOpenInventoryBalanceTx(tx, commodity, input.WarehouseCode, initialQty)
		if err != nil {
			return err
		}

		beforeQty := balance.AvailableQty
		changeQty := input.ActualQty - beforeQty
		result = map[string]any{
			"commodity_id":    input.CommodityID,
			"before_qty":      beforeQty,
			"actual_qty":      input.ActualQty,
			"change_qty":      changeQty,
			"warehouse_code":  input.WarehouseCode,
			"inventory_equal": changeQty == 0,
		}
		if changeQty == 0 {
			return nil
		}

		remark := strings.TrimSpace(input.Remark)
		if remark == "" {
			remark = "stock check adjustment"
		}
		if err := updateOpenInventoryBalanceQtyTx(tx, input.CommodityID, input.WarehouseCode, input.ActualQty); err != nil {
			return err
		}
		idempotencyKey := fmt.Sprintf("inventory:%s:%s:%s:%d", InventoryChangeStockCheck, input.CommodityID, input.WarehouseCode, time.Now().UnixNano())
		if err := createInventoryStockMovementTx(tx, commodity, input.WarehouseCode, InventoryChangeStockCheck, "inventory", "stock_check", idempotencyKey, beforeQty, changeQty, input.ActualQty, input.OperatorID, remark); err != nil {
			return err
		}
		if _, err := refreshLegacyInventoryFromOpenTx(tx, input.CommodityID, commodity.StyleCode); err != nil {
			return err
		}
		return tx.Create(&models.InventoryLog{
			CommodityID:   input.CommodityID,
			StyleCode:     commodity.StyleCode,
			WarehouseCode: input.WarehouseCode,
			BeforeQty:     beforeQty,
			ChangeQty:     changeQty,
			AfterQty:      input.ActualQty,
			ChangeType:    InventoryChangeStockCheck,
			OperatorID:    input.OperatorID,
			Remark:        remark,
		}).Error
	})
	return result, err
}

func validateInventoryTransferInput(input *InventoryTransferInput) error {
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	if input.CommodityID == "" {
		return fmt.Errorf("commodity_id不能为空")
	}
	if input.Qty <= 0 {
		return fmt.Errorf("调拨数量必须大于0")
	}
	input.SourceWarehouseCode = strings.TrimSpace(input.SourceWarehouseCode)
	input.TargetWarehouseCode = strings.TrimSpace(input.TargetWarehouseCode)
	if input.SourceWarehouseCode == "" || input.TargetWarehouseCode == "" {
		return fmt.Errorf("源仓库和目标仓库不能为空")
	}
	input.SourceWarehouseCode = normalizeOpenInventoryWarehouseCode(input.SourceWarehouseCode)
	input.TargetWarehouseCode = normalizeOpenInventoryWarehouseCode(input.TargetWarehouseCode)
	if input.SourceWarehouseCode == input.TargetWarehouseCode {
		return fmt.Errorf("源仓库和目标仓库不能相同")
	}
	return nil
}

func validateInventoryStockCheckInput(input *InventoryStockCheckInput) error {
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	if input.CommodityID == "" {
		return fmt.Errorf("commodity_id不能为空")
	}
	if input.ActualQty < 0 {
		return fmt.Errorf("实际库存不能小于0")
	}
	input.WarehouseCode = normalizeOpenInventoryWarehouseCode(input.WarehouseCode)
	return nil
}

// QueryInventoryLogs 查询库存变更日志
// 支持按商品ID、款式编码、变动类型、关联订单/子订单/退货单等条件筛选
// 返回库存变动日志列表、总数量、当前页码、每页数量
func QueryInventoryLogs(input InventoryLogQueryInput) ([]InventoryLogView, int64, int, int, error) {
	page, pageSize := normalizePage(input.Page, input.PageSize)
	legacyWhere, legacyArgs := buildLegacyInventoryLogWhere(input)
	openWhere, openArgs := buildOpenInventoryMovementWhere(input)
	unionSQL := inventoryLogUnionSQL(legacyWhere, openWhere)
	countArgs := append(append([]any{}, legacyArgs...), openArgs...)
	var total int64
	if err := db.DB.Raw("SELECT COUNT(*) FROM ("+unionSQL+") AS inventory_logs", countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	var logs []InventoryLogView
	queryArgs := append(append([]any{}, countArgs...), pageSize, (page-1)*pageSize)
	querySQL := "SELECT * FROM (" + unionSQL + ") AS inventory_logs ORDER BY created_at DESC, source_id DESC LIMIT ? OFFSET ?"
	if err := db.DB.Raw(querySQL, queryArgs...).Scan(&logs).Error; err != nil {
		return nil, 0, page, pageSize, err
	}
	return logs, total, page, pageSize, nil
}

func inventoryLogUnionSQL(legacyWhere, openWhere string) string {
	legacySQL := `
SELECT
  id AS source_id,
  'legacy' AS source,
  '' AS movement_no,
  commodity_id,
  style_code,
  warehouse_code,
  before_qty,
  change_qty,
  after_qty,
  change_type,
  related_order_id,
  related_sub_order_id,
  related_return_id,
  operator_id,
  remark,
  created_at,
  '' AS biz_type,
  '' AS biz_id,
  '' AS idempotency_key
FROM inventory_log` + legacyWhere

	openSQL := `
SELECT
  id AS source_id,
  'open' AS source,
  movement_no,
  commodity_id,
  style_code,
  warehouse_code,
  before_qty,
  change_qty,
  after_qty,
  movement_type AS change_type,
  '' AS related_order_id,
  CASE WHEN biz_type = 'order' THEN biz_id ELSE '' END AS related_sub_order_id,
  CASE WHEN biz_type = 'return' THEN biz_id ELSE '' END AS related_return_id,
  operator_id,
  remark,
  created_at,
  biz_type,
  biz_id,
  idempotency_key
FROM inventory_stock_movements` + openWhere

	return legacySQL + " UNION ALL " + openSQL
}

func buildLegacyInventoryLogWhere(input InventoryLogQueryInput) (string, []any) {
	clauses := []string{}
	args := []any{}
	if strings.TrimSpace(input.CommodityID) != "" {
		clauses = append(clauses, "commodity_id = ?")
		args = append(args, strings.TrimSpace(input.CommodityID))
	}
	if strings.TrimSpace(input.StyleCode) != "" {
		clauses = append(clauses, "style_code = ?")
		args = append(args, strings.TrimSpace(input.StyleCode))
	}
	if strings.TrimSpace(input.ChangeType) != "" {
		clauses = append(clauses, "change_type = ?")
		args = append(args, strings.TrimSpace(input.ChangeType))
	}
	if strings.TrimSpace(input.RelatedOrderID) != "" {
		clauses = append(clauses, "related_order_id = ?")
		args = append(args, strings.TrimSpace(input.RelatedOrderID))
	}
	if strings.TrimSpace(input.RelatedSubOrderID) != "" {
		clauses = append(clauses, "related_sub_order_id = ?")
		args = append(args, strings.TrimSpace(input.RelatedSubOrderID))
	}
	if strings.TrimSpace(input.RelatedReturnID) != "" {
		clauses = append(clauses, "related_return_id = ?")
		args = append(args, strings.TrimSpace(input.RelatedReturnID))
	}
	return inventoryLogWhereSQL(clauses), args
}

func buildOpenInventoryMovementWhere(input InventoryLogQueryInput) (string, []any) {
	clauses := []string{}
	args := []any{}
	if strings.TrimSpace(input.CommodityID) != "" {
		clauses = append(clauses, "commodity_id = ?")
		args = append(args, strings.TrimSpace(input.CommodityID))
	}
	if strings.TrimSpace(input.StyleCode) != "" {
		clauses = append(clauses, "style_code = ?")
		args = append(args, strings.TrimSpace(input.StyleCode))
	}
	if strings.TrimSpace(input.ChangeType) != "" {
		clauses = append(clauses, "movement_type = ?")
		args = append(args, strings.TrimSpace(input.ChangeType))
	}
	if strings.TrimSpace(input.RelatedOrderID) != "" {
		clauses = append(clauses, "biz_type = ? AND biz_id = ?")
		args = append(args, "order", strings.TrimSpace(input.RelatedOrderID))
	}
	if strings.TrimSpace(input.RelatedSubOrderID) != "" {
		clauses = append(clauses, "biz_type = ? AND biz_id = ?")
		args = append(args, "order", strings.TrimSpace(input.RelatedSubOrderID))
	}
	if strings.TrimSpace(input.RelatedReturnID) != "" {
		clauses = append(clauses, "biz_type = ? AND biz_id = ?")
		args = append(args, "return", strings.TrimSpace(input.RelatedReturnID))
	}
	return inventoryLogWhereSQL(clauses), args
}

func inventoryLogWhereSQL(clauses []string) string {
	if len(clauses) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(clauses, " AND ")
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

type InventoryLogView struct {
	SourceID          uint      `gorm:"column:source_id" json:"source_id"`
	Source            string    `gorm:"column:source" json:"source"`
	MovementNo        string    `gorm:"column:movement_no" json:"movement_no"`
	CommodityID       string    `gorm:"column:commodity_id" json:"commodity_id"`
	StyleCode         string    `gorm:"column:style_code" json:"style_code"`
	WarehouseCode     string    `gorm:"column:warehouse_code" json:"warehouse_code"`
	BeforeQty         int       `gorm:"column:before_qty" json:"before_qty"`
	ChangeQty         int       `gorm:"column:change_qty" json:"change_qty"`
	AfterQty          int       `gorm:"column:after_qty" json:"after_qty"`
	ChangeType        string    `gorm:"column:change_type" json:"change_type"`
	RelatedOrderID    string    `gorm:"column:related_order_id" json:"related_order_id"`
	RelatedSubOrderID string    `gorm:"column:related_sub_order_id" json:"related_sub_order_id"`
	RelatedReturnID   string    `gorm:"column:related_return_id" json:"related_return_id"`
	OperatorID        string    `gorm:"column:operator_id" json:"operator_id"`
	Remark            string    `gorm:"column:remark" json:"remark"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	BizType           string    `gorm:"column:biz_type" json:"biz_type"`
	BizID             string    `gorm:"column:biz_id" json:"biz_id"`
	IdempotencyKey    string    `gorm:"column:idempotency_key" json:"idempotency_key"`
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

	openChange, err := applyOpenInventoryChangeTx(tx, input, commodity)
	if err != nil {
		return err
	}
	if openChange == nil {
		return nil
	}

	if _, err := refreshLegacyInventoryFromOpenTx(tx, input.CommodityID, commodity.StyleCode); err != nil {
		return err
	}

	log := models.InventoryLog{
		CommodityID:       input.CommodityID,
		StyleCode:         commodity.StyleCode,
		WarehouseCode:     openChange.WarehouseCode,
		BeforeQty:         openChange.BeforeQty,
		ChangeQty:         input.ChangeQty,
		AfterQty:          openChange.AfterQty,
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
		Where("change_type IN ? AND related_sub_order_id = ?", []string{InventoryChangeOrderDeduct, legacyInventoryChangeOrderDeduct}, subOrder.SubOrderID).
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
