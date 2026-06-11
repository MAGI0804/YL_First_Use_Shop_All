package method

import (
	"Member_shop/models"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type openInventoryChangeResult struct {
	WarehouseCode string
	BeforeQty     int
	AfterQty      int
}

func applyOpenInventoryChangeTx(tx *gorm.DB, input ChangeInventoryInput, commodity models.Commodity) (*openInventoryChangeResult, error) {
	if err := ensureOpenInventoryWarehouseTx(tx); err != nil {
		return nil, err
	}
	if err := ensureOpenInventorySKUTx(tx, commodity); err != nil {
		return nil, err
	}

	warehouseCode, err := resolveOpenInventoryWarehouseCodeTx(tx, input.CommodityID, input.WarehouseCode)
	if err != nil {
		return nil, err
	}

	idempotencyKey, stableID := openInventoryIdempotencyKey(input)
	if stableID {
		var count int64
		if err := tx.Model(&models.InventoryStockMovement{}).
			Where("idempotency_key = ?", idempotencyKey).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, nil
		}
	}

	balance, err := lockOpenInventoryBalanceTx(tx, commodity, warehouseCode)
	if err != nil {
		return nil, err
	}

	beforeQty := balance.AvailableQty
	afterQty := beforeQty + input.ChangeQty
	if afterQty < 0 {
		return nil, fmt.Errorf("商品%s库存不足，当前库存%d，需要%d", input.CommodityID, beforeQty, -input.ChangeQty)
	}

	if err := tx.Model(&models.InventoryStockBalance{}).
		Where("commodity_id = ? AND warehouse_code = ?", input.CommodityID, warehouseCode).
		Updates(map[string]any{
			"on_hand_qty":   afterQty,
			"available_qty": afterQty,
			"locked_qty":    0,
			"version":       gorm.Expr("version + 1"),
		}).Error; err != nil {
		return nil, err
	}

	movement := models.InventoryStockMovement{
		MovementNo:     openInventoryMovementNo(idempotencyKey),
		CommodityID:    input.CommodityID,
		StyleCode:      commodity.StyleCode,
		WarehouseCode:  warehouseCode,
		MovementType:   input.ChangeType,
		BizType:        openInventoryBizType(input),
		BizID:          openInventoryBizID(input),
		IdempotencyKey: idempotencyKey,
		BeforeQty:      beforeQty,
		ChangeQty:      input.ChangeQty,
		AfterQty:       afterQty,
		OnHandDelta:    input.ChangeQty,
		LockedDelta:    0,
		AvailableDelta: input.ChangeQty,
		OperatorID:     input.OperatorID,
		Remark:         input.Remark,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}).Create(&movement).Error; err != nil {
		return nil, err
	}

	return &openInventoryChangeResult{
		WarehouseCode: warehouseCode,
		BeforeQty:     beforeQty,
		AfterQty:      afterQty,
	}, nil
}

func setOpenInventoryAvailableTx(tx *gorm.DB, commodity models.Commodity, warehouseCode string, afterQty int, movementType, bizType, bizID, idempotencyKey, operatorID, remark string) (*openInventoryChangeResult, error) {
	if afterQty < 0 {
		return nil, fmt.Errorf("库存不能小于0")
	}
	if err := ensureOpenInventoryWarehouseTx(tx); err != nil {
		return nil, err
	}
	if err := ensureOpenInventorySKUTx(tx, commodity); err != nil {
		return nil, err
	}

	warehouseCode, err := resolveOpenInventoryWarehouseCodeTx(tx, commodity.CommodityID, warehouseCode)
	if err != nil {
		return nil, err
	}
	if idempotencyKey != "" {
		var count int64
		if err := tx.Model(&models.InventoryStockMovement{}).
			Where("idempotency_key = ?", idempotencyKey).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, nil
		}
	} else {
		idempotencyKey = fmt.Sprintf("inventory:%s:%s:%s:%d", movementType, commodity.CommodityID, warehouseCode, time.Now().UnixNano())
	}

	balance, err := lockOpenInventoryBalanceTx(tx, commodity, warehouseCode)
	if err != nil {
		return nil, err
	}

	beforeQty := balance.AvailableQty
	changeQty := afterQty - beforeQty
	if changeQty == 0 {
		return &openInventoryChangeResult{
			WarehouseCode: warehouseCode,
			BeforeQty:     beforeQty,
			AfterQty:      afterQty,
		}, nil
	}

	if err := tx.Model(&models.InventoryStockBalance{}).
		Where("commodity_id = ? AND warehouse_code = ?", commodity.CommodityID, warehouseCode).
		Updates(map[string]any{
			"on_hand_qty":   afterQty,
			"available_qty": afterQty,
			"locked_qty":    0,
			"version":       gorm.Expr("version + 1"),
		}).Error; err != nil {
		return nil, err
	}

	if err := createInventoryStockMovementTx(tx, commodity, warehouseCode, movementType, bizType, bizID, idempotencyKey, beforeQty, changeQty, afterQty, operatorID, remark); err != nil {
		return nil, err
	}

	return &openInventoryChangeResult{
		WarehouseCode: warehouseCode,
		BeforeQty:     beforeQty,
		AfterQty:      afterQty,
	}, nil
}

func createInventoryStockMovementTx(tx *gorm.DB, commodity models.Commodity, warehouseCode, movementType, bizType, bizID, idempotencyKey string, beforeQty, changeQty, afterQty int, operatorID, remark string) error {
	movement := models.InventoryStockMovement{
		MovementNo:     openInventoryMovementNo(idempotencyKey),
		CommodityID:    commodity.CommodityID,
		StyleCode:      commodity.StyleCode,
		WarehouseCode:  normalizeOpenInventoryWarehouseCode(warehouseCode),
		MovementType:   movementType,
		BizType:        bizType,
		BizID:          bizID,
		IdempotencyKey: idempotencyKey,
		BeforeQty:      beforeQty,
		ChangeQty:      changeQty,
		AfterQty:       afterQty,
		OnHandDelta:    changeQty,
		LockedDelta:    0,
		AvailableDelta: changeQty,
		OperatorID:     operatorID,
		Remark:         remark,
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}).Create(&movement).Error
}

func ensureOpenInventoryWarehouseTx(tx *gorm.DB) error {
	return ensureOpenInventoryWarehouseCodeTx(tx, models.DefaultInventoryWarehouseCode)
}

func ensureOpenInventoryWarehouseCodeTx(tx *gorm.DB, warehouseCode string) error {
	now := time.Now()
	warehouseCode = normalizeOpenInventoryWarehouseCode(warehouseCode)
	warehouseName := warehouseCode
	isDefault := warehouseCode == models.DefaultInventoryWarehouseCode
	if isDefault {
		warehouseName = "默认仓"
	}
	warehouse := models.InventoryWarehouse{
		WarehouseCode: warehouseCode,
		WarehouseName: warehouseName,
		Source:        models.InventorySourceLocal,
		IsDefault:     isDefault,
		Status:        models.InventoryWarehouseStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "warehouse_code"}},
		DoUpdates: clause.Assignments(map[string]any{
			"warehouse_name": warehouse.WarehouseName,
			"source":         warehouse.Source,
			"is_default":     warehouse.IsDefault,
			"status":         warehouse.Status,
			"updated_at":     now,
		}),
	}).Create(&warehouse).Error
}

func ensureOpenInventorySKUTx(tx *gorm.DB, commodity models.Commodity) error {
	now := time.Now()
	sku := models.InventorySKU{
		CommodityID: commodity.CommodityID,
		StyleCode:   commodity.StyleCode,
		SpecCode:    commodity.SpecCode,
		Name:        commodity.Name,
		Size:        commodity.Size,
		Color:       commodity.Color,
		Category:    commodity.Category,
		Status:      models.InventorySKUStatusActive,
		CreatedAt:   commodity.CreatedAt,
		UpdatedAt:   now,
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commodity_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"style_code": sku.StyleCode,
			"spec_code":  sku.SpecCode,
			"name":       sku.Name,
			"size":       sku.Size,
			"color":      sku.Color,
			"category":   sku.Category,
			"status":     sku.Status,
			"updated_at": now,
		}),
	}).Create(&sku).Error
}

func resolveOpenInventoryWarehouseCodeTx(tx *gorm.DB, commodityID, warehouseCode string) (string, error) {
	warehouseCode = normalizeOpenInventoryWarehouseCode(warehouseCode)
	if warehouseCode == "" || warehouseCode == models.DefaultInventoryWarehouseCode {
		return models.DefaultInventoryWarehouseCode, nil
	}

	var count int64
	if err := tx.Model(&models.InventoryStockBalance{}).
		Where("commodity_id = ? AND warehouse_code = ?", commodityID, warehouseCode).
		Count(&count).Error; err != nil {
		return "", err
	}
	if count == 0 {
		return models.DefaultInventoryWarehouseCode, nil
	}
	return warehouseCode, nil
}

func lockOpenInventoryBalanceTx(tx *gorm.DB, commodity models.Commodity, warehouseCode string) (models.InventoryStockBalance, error) {
	return lockOrCreateOpenInventoryBalanceTx(tx, commodity, warehouseCode, nonNegativeOpenInventoryQty(commodity.Inventory))
}

func lockOrCreateOpenInventoryBalanceTx(tx *gorm.DB, commodity models.Commodity, warehouseCode string, initialQty int) (models.InventoryStockBalance, error) {
	warehouseCode = normalizeOpenInventoryWarehouseCode(warehouseCode)
	var balance models.InventoryStockBalance
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("commodity_id = ? AND warehouse_code = ?", commodity.CommodityID, warehouseCode).
		First(&balance).Error
	if err == nil {
		return balance, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return balance, err
	}

	qty := nonNegativeOpenInventoryQty(initialQty)
	balance = models.InventoryStockBalance{
		CommodityID:   commodity.CommodityID,
		WarehouseCode: warehouseCode,
		OnHandQty:     qty,
		LockedQty:     0,
		AvailableQty:  qty,
		Version:       1,
	}
	if err := tx.Create(&balance).Error; err != nil {
		return balance, err
	}
	return balance, nil
}

func updateOpenInventoryBalanceQtyTx(tx *gorm.DB, commodityID, warehouseCode string, afterQty int) error {
	return tx.Model(&models.InventoryStockBalance{}).
		Where("commodity_id = ? AND warehouse_code = ?", commodityID, normalizeOpenInventoryWarehouseCode(warehouseCode)).
		Updates(map[string]any{
			"on_hand_qty":   afterQty,
			"available_qty": afterQty,
			"locked_qty":    0,
			"version":       gorm.Expr("version + 1"),
		}).Error
}

func refreshLegacyInventoryFromOpenTx(tx *gorm.DB, commodityID, styleCode string) (int, error) {
	var total int64
	if err := tx.Model(&models.InventoryStockBalance{}).
		Where("commodity_id = ?", commodityID).
		Select("COALESCE(SUM(available_qty), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	totalInventory := int(total)
	if err := tx.Model(&models.Commodity{}).
		Where("commodity_id = ?", commodityID).
		Update("inventory", totalInventory).Error; err != nil {
		return 0, err
	}
	if err := tx.Model(&models.CommoditySituation{}).
		Where("commodity_id = ?", commodityID).
		Update("inventory", totalInventory).Error; err != nil {
		return 0, err
	}
	if styleCode != "" {
		if err := refreshStyleCodeInventoryTx(tx, styleCode); err != nil {
			return 0, err
		}
	}
	return totalInventory, nil
}

func openInventoryIdempotencyKey(input ChangeInventoryInput) (string, bool) {
	parts := []string{"inventory", input.ChangeType, input.CommodityID}
	switch {
	case input.RelatedSubOrderID != "":
		parts = append(parts, "sub_order", input.RelatedSubOrderID)
	case input.RelatedReturnID != "":
		parts = append(parts, "return", input.RelatedReturnID)
	case input.RelatedOrderID != "":
		parts = append(parts, "order", input.RelatedOrderID)
	default:
		parts = append(parts, "manual", fmt.Sprintf("%d", time.Now().UnixNano()))
		return strings.Join(parts, ":"), false
	}
	return strings.Join(parts, ":"), true
}

func openInventoryMovementNo(idempotencyKey string) string {
	sum := sha1.Sum([]byte(idempotencyKey))
	return "MV:" + hex.EncodeToString(sum[:])
}

func openInventoryBizType(input ChangeInventoryInput) string {
	if input.RelatedReturnID != "" {
		return "return"
	}
	if input.RelatedOrderID != "" || input.RelatedSubOrderID != "" {
		return "order"
	}
	return "inventory"
}

func openInventoryBizID(input ChangeInventoryInput) string {
	if input.RelatedReturnID != "" {
		return input.RelatedReturnID
	}
	if input.RelatedSubOrderID != "" {
		return input.RelatedSubOrderID
	}
	if input.RelatedOrderID != "" {
		return input.RelatedOrderID
	}
	return input.ChangeType
}

func nonNegativeOpenInventoryQty(qty int) int {
	if qty < 0 {
		return 0
	}
	return qty
}

func normalizeOpenInventoryWarehouseCode(warehouseCode string) string {
	warehouseCode = strings.TrimSpace(warehouseCode)
	if warehouseCode == "" || strings.EqualFold(warehouseCode, models.DefaultInventoryWarehouseCode) {
		return models.DefaultInventoryWarehouseCode
	}
	return warehouseCode
}
