package models

import "time"

const (
	DefaultInventoryWarehouseCode = "DEFAULT"

	InventoryWarehouseStatusActive   = "active"
	InventoryWarehouseStatusDisabled = "disabled"

	InventorySourceLocal     = "local"
	InventorySourceJushuitan = "jushuitan"

	InventorySKUStatusActive   = "active"
	InventorySKUStatusDisabled = "disabled"

	InventoryMovementOpeningBalance = "opening_balance"
	InventoryMovementBizMigration   = "migration"
)

type InventoryWarehouse struct {
	WarehouseCode string    `gorm:"column:warehouse_code;primaryKey;size:50;comment:仓库编码" json:"warehouse_code"`
	WarehouseName string    `gorm:"column:warehouse_name;size:100;not null;comment:仓库名称" json:"warehouse_name"`
	Source        string    `gorm:"column:source;size:30;index;not null;comment:库存来源" json:"source"`
	IsDefault     bool      `gorm:"column:is_default;index;not null;default:false;comment:是否默认仓" json:"is_default"`
	Status        string    `gorm:"column:status;size:20;index;not null;comment:状态" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (InventoryWarehouse) TableName() string {
	return "inventory_warehouses"
}

type InventorySKU struct {
	CommodityID string    `gorm:"column:commodity_id;primaryKey;size:100;comment:本地SKU商品ID" json:"commodity_id"`
	StyleCode   string    `gorm:"column:style_code;size:50;index;comment:款号" json:"style_code"`
	SpecCode    string    `gorm:"column:spec_code;size:100;index;comment:外部规格码" json:"spec_code"`
	Name        string    `gorm:"column:name;size:255;not null;comment:商品名称" json:"name"`
	Size        string    `gorm:"column:size;size:50;comment:尺码" json:"size"`
	Color       string    `gorm:"column:color;size:50;comment:颜色" json:"color"`
	Category    string    `gorm:"column:category;size:100;index;comment:分类" json:"category"`
	Status      string    `gorm:"column:status;size:20;index;not null;comment:状态" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (InventorySKU) TableName() string {
	return "inventory_skus"
}

type InventoryStockBalance struct {
	ID            uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommodityID   string    `gorm:"column:commodity_id;size:100;uniqueIndex:idx_inventory_balance_sku_warehouse;index;not null;comment:本地SKU商品ID" json:"commodity_id"`
	WarehouseCode string    `gorm:"column:warehouse_code;size:50;uniqueIndex:idx_inventory_balance_sku_warehouse;index;not null;comment:仓库编码" json:"warehouse_code"`
	OnHandQty     int       `gorm:"column:on_hand_qty;not null;default:0;comment:实物库存" json:"on_hand_qty"`
	LockedQty     int       `gorm:"column:locked_qty;not null;default:0;comment:锁定库存" json:"locked_qty"`
	AvailableQty  int       `gorm:"column:available_qty;not null;default:0;comment:可用库存" json:"available_qty"`
	Version       int       `gorm:"column:version;not null;default:1;comment:乐观锁版本" json:"version"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (InventoryStockBalance) TableName() string {
	return "inventory_stock_balances"
}

type InventoryStockMovement struct {
	ID             uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MovementNo     string    `gorm:"column:movement_no;size:160;uniqueIndex;not null;comment:库存流水号" json:"movement_no"`
	CommodityID    string    `gorm:"column:commodity_id;size:100;index;not null;comment:本地SKU商品ID" json:"commodity_id"`
	StyleCode      string    `gorm:"column:style_code;size:50;index;comment:款号" json:"style_code"`
	WarehouseCode  string    `gorm:"column:warehouse_code;size:50;index;not null;comment:仓库编码" json:"warehouse_code"`
	MovementType   string    `gorm:"column:movement_type;size:50;index;not null;comment:库存动作类型" json:"movement_type"`
	BizType        string    `gorm:"column:biz_type;size:50;index;not null;comment:业务类型" json:"biz_type"`
	BizID          string    `gorm:"column:biz_id;size:100;index;comment:业务ID" json:"biz_id"`
	IdempotencyKey string    `gorm:"column:idempotency_key;size:160;uniqueIndex;not null;comment:幂等键" json:"idempotency_key"`
	BeforeQty      int       `gorm:"column:before_qty;not null;default:0;comment:变更前可用库存" json:"before_qty"`
	ChangeQty      int       `gorm:"column:change_qty;not null;default:0;comment:可用库存变更量" json:"change_qty"`
	AfterQty       int       `gorm:"column:after_qty;not null;default:0;comment:变更后可用库存" json:"after_qty"`
	OnHandDelta    int       `gorm:"column:on_hand_delta;not null;default:0;comment:实物库存变更量" json:"on_hand_delta"`
	LockedDelta    int       `gorm:"column:locked_delta;not null;default:0;comment:锁定库存变更量" json:"locked_delta"`
	AvailableDelta int       `gorm:"column:available_delta;not null;default:0;comment:可用库存变更量" json:"available_delta"`
	OperatorID     string    `gorm:"column:operator_id;size:50;comment:操作人" json:"operator_id"`
	Remark         string    `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

func (InventoryStockMovement) TableName() string {
	return "inventory_stock_movements"
}
