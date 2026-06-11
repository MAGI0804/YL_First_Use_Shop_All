package db

import (
	"Member_shop/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const openInventoryOpeningBizID = "inventory_opening_20260611"

func seedOpenInventorySnapshot() {
	if DB == nil {
		return
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		return seedOpenInventorySnapshotTx(tx)
	}); err != nil {
		log.Printf("seed open inventory snapshot failed: %v", err)
		return
	}

	log.Println("seed open inventory snapshot completed")
}

func seedOpenInventorySnapshotTx(tx *gorm.DB) error {
	if err := seedDefaultInventoryWarehouse(tx); err != nil {
		return err
	}

	var commodities []models.Commodity
	if err := tx.Order("commodity_id ASC").Find(&commodities).Error; err != nil {
		return fmt.Errorf("query commodities for open inventory seed: %w", err)
	}

	for _, commodity := range commodities {
		if err := seedOpenInventoryCommodity(tx, commodity); err != nil {
			return err
		}
	}

	return nil
}

func seedDefaultInventoryWarehouse(tx *gorm.DB) error {
	now := time.Now()
	warehouse := models.InventoryWarehouse{
		WarehouseCode: models.DefaultInventoryWarehouseCode,
		WarehouseName: "默认仓",
		Source:        models.InventorySourceLocal,
		IsDefault:     true,
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

func seedOpenInventoryCommodity(tx *gorm.DB, commodity models.Commodity) error {
	if commodity.CommodityID == "" {
		return nil
	}

	if err := upsertOpenInventorySKU(tx, commodity); err != nil {
		return fmt.Errorf("seed inventory sku %s: %w", commodity.CommodityID, err)
	}
	if err := upsertOpenInventoryBalance(tx, commodity); err != nil {
		return fmt.Errorf("seed inventory balance %s: %w", commodity.CommodityID, err)
	}
	if err := createOpeningInventoryMovement(tx, commodity); err != nil {
		return fmt.Errorf("seed opening inventory movement %s: %w", commodity.CommodityID, err)
	}

	return nil
}

func upsertOpenInventorySKU(tx *gorm.DB, commodity models.Commodity) error {
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

func upsertOpenInventoryBalance(tx *gorm.DB, commodity models.Commodity) error {
	now := time.Now()
	qty := nonNegativeInventoryQty(commodity.Inventory)
	balance := models.InventoryStockBalance{
		CommodityID:   commodity.CommodityID,
		WarehouseCode: models.DefaultInventoryWarehouseCode,
		OnHandQty:     qty,
		LockedQty:     0,
		AvailableQty:  qty,
		Version:       1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commodity_id"}, {Name: "warehouse_code"}},
		DoUpdates: clause.Assignments(map[string]any{
			"on_hand_qty":   balance.OnHandQty,
			"locked_qty":    balance.LockedQty,
			"available_qty": balance.AvailableQty,
			"updated_at":    now,
		}),
	}).Create(&balance).Error
}

func createOpeningInventoryMovement(tx *gorm.DB, commodity models.Commodity) error {
	qty := nonNegativeInventoryQty(commodity.Inventory)
	idempotencyKey := fmt.Sprintf("migration:opening:%s", commodity.CommodityID)
	movement := models.InventoryStockMovement{
		MovementNo:     fmt.Sprintf("OPENING:%s", commodity.CommodityID),
		CommodityID:    commodity.CommodityID,
		StyleCode:      commodity.StyleCode,
		WarehouseCode:  models.DefaultInventoryWarehouseCode,
		MovementType:   models.InventoryMovementOpeningBalance,
		BizType:        models.InventoryMovementBizMigration,
		BizID:          openInventoryOpeningBizID,
		IdempotencyKey: idempotencyKey,
		BeforeQty:      0,
		ChangeQty:      qty,
		AfterQty:       qty,
		OnHandDelta:    qty,
		LockedDelta:    0,
		AvailableDelta: qty,
		OperatorID:     "migration",
		Remark:         "库存开放重构初始化补数，来源 Commodity_data.inventory",
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}).Create(&movement).Error
}

func nonNegativeInventoryQty(qty int) int {
	if qty < 0 {
		return 0
	}
	return qty
}
