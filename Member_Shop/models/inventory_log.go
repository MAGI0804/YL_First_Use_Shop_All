package models

import "time"

type InventoryLog struct {
	ID                uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommodityID       string    `gorm:"column:commodity_id;size:100;index;not null;comment:SKU商品ID" json:"commodity_id"`
	StyleCode         string    `gorm:"column:style_code;size:50;index;comment:款号" json:"style_code"`
	WarehouseCode     string    `gorm:"column:warehouse_code;size:50;index;comment:仓库编码" json:"warehouse_code"`
	BeforeQty         int       `gorm:"column:before_qty;not null;comment:变动前库存" json:"before_qty"`
	ChangeQty         int       `gorm:"column:change_qty;not null;comment:变动数量" json:"change_qty"`
	AfterQty          int       `gorm:"column:after_qty;not null;comment:变动后库存" json:"after_qty"`
	ChangeType        string    `gorm:"column:change_type;size:40;index;not null;comment:变动类型" json:"change_type"`
	RelatedOrderID    string    `gorm:"column:related_order_id;size:30;index;comment:关联主订单" json:"related_order_id"`
	RelatedSubOrderID string    `gorm:"column:related_sub_order_id;size:30;index;comment:关联子订单" json:"related_sub_order_id"`
	RelatedReturnID   string    `gorm:"column:related_return_id;size:30;index;comment:关联售后单" json:"related_return_id"`
	OperatorID        string    `gorm:"column:operator_id;size:50;comment:操作人" json:"operator_id"`
	Remark            string    `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

func (InventoryLog) TableName() string {
	return "inventory_log"
}
