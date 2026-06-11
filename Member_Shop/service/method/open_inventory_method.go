package method

import (
	"Member_shop/db"
	"fmt"
	"strings"
	"time"
)

type OpenInventoryQueryInput struct {
	CommodityID   string
	StyleCode     string
	WarehouseCode string
}

type OpenInventoryBalanceView struct {
	CommodityID   string    `json:"commodity_id"`
	StyleCode     string    `json:"style_code"`
	SpecCode      string    `json:"spec_code"`
	Name          string    `json:"name"`
	Size          string    `json:"size"`
	Color         string    `json:"color"`
	Category      string    `json:"category"`
	WarehouseCode string    `json:"warehouse_code"`
	OnHandQty     int       `json:"on_hand_qty"`
	LockedQty     int       `json:"locked_qty"`
	AvailableQty  int       `json:"available_qty"`
	Version       int       `json:"version"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OpenInventorySummary struct {
	TotalOnHandQty    int `json:"total_on_hand_qty"`
	TotalLockedQty    int `json:"total_locked_qty"`
	TotalAvailableQty int `json:"total_available_qty"`
}

type OpenInventoryQueryResult struct {
	CommodityID   string                     `json:"commodity_id"`
	StyleCode     string                     `json:"style_code"`
	WarehouseCode string                     `json:"warehouse_code"`
	Summary       OpenInventorySummary       `json:"summary"`
	Items         []OpenInventoryBalanceView `json:"items"`
}

func QueryOpenInventory(input OpenInventoryQueryInput) (*OpenInventoryQueryResult, error) {
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	input.StyleCode = strings.TrimSpace(input.StyleCode)
	input.WarehouseCode = strings.TrimSpace(input.WarehouseCode)
	if input.CommodityID == "" && input.StyleCode == "" {
		return nil, fmt.Errorf("commodity_id或style_code不能为空")
	}

	query := db.DB.Table("inventory_stock_balances AS b").
		Select(strings.Join([]string{
			"b.commodity_id",
			"s.style_code",
			"s.spec_code",
			"s.name",
			"s.size",
			"s.color",
			"s.category",
			"b.warehouse_code",
			"b.on_hand_qty",
			"b.locked_qty",
			"b.available_qty",
			"b.version",
			"b.updated_at",
		}, ", ")).
		Joins("JOIN inventory_skus AS s ON s.commodity_id = b.commodity_id")

	if input.CommodityID != "" {
		query = query.Where("b.commodity_id = ?", input.CommodityID)
	}
	if input.StyleCode != "" {
		query = query.Where("s.style_code = ?", input.StyleCode)
	}
	if input.WarehouseCode != "" {
		query = query.Where("b.warehouse_code = ?", input.WarehouseCode)
	}

	var items []OpenInventoryBalanceView
	if err := query.Order("s.style_code ASC, b.commodity_id ASC, b.warehouse_code ASC").Scan(&items).Error; err != nil {
		return nil, err
	}

	result := &OpenInventoryQueryResult{
		CommodityID:   input.CommodityID,
		StyleCode:     input.StyleCode,
		WarehouseCode: input.WarehouseCode,
		Items:         items,
	}
	for _, item := range items {
		result.Summary.TotalOnHandQty += item.OnHandQty
		result.Summary.TotalLockedQty += item.LockedQty
		result.Summary.TotalAvailableQty += item.AvailableQty
	}

	return result, nil
}
