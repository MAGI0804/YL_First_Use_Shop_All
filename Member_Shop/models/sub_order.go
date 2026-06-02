package models

import (
	"time"

	"gorm.io/gorm"
)

// SubOrder 子订单模型
// 用于存储订单中的子订单信息，每个主订单可以包含多个子订单，主要用于拆单场景
type SubOrder struct {
	SubOrderID          string     `gorm:"column:sub_order_id;primaryKey;size:30;comment:子订单ID/子订单号" json:"sub_order_id"`                             //子订单ID
	OrderID             string     `gorm:"column:order_id;size:20;not null;index;comment:关联的主订单ID" json:"order_id"`                                   //关联的主订单ID
	ProductInfo         string     `gorm:"column:product_info;type:text;not null;comment:商品信息JSON" json:"product_info"`                               //商品信息JSON
	ProductName         string     `gorm:"column:product_name;size:255;not null;comment:商品名称" json:"product_name"`                                    //商品名称
	Status              string     `gorm:"column:status;size:20;not null;default:'pending';comment:sub-order lifecycle status" json:"status"`         // Lifecycle status, not payment status.
	PayStatus           string     `gorm:"column:pay_status;size:20;not null;default:'unpaid';comment:payment status: unpaid/paid" json:"pay_status"` // Payment status after signed delivery.
	SubAmount           float64    `gorm:"column:sub_amount;type:decimal(10,2);not null;comment:子订单金额" json:"sub_amount"`                             //子订单金额
	ExpressCompany      string     `gorm:"column:express_company;size:50;null;comment:物流公司" json:"express_company"`                                   //物流公司
	ExpressNumber       string     `gorm:"column:express_number;size:50;null;comment:物流单号" json:"express_number"`                                     //物流单号
	IsShipped           string     `gorm:"column:is_shipped;size:10;not null;default:'no';comment:是否已发货(yes/no)" json:"is_shipped"`                   //是否已发货
	ShippedTime         *time.Time `gorm:"column:shipped_time;null;comment:发货时间" json:"shipped_time"`                                                 //发货时间
	JushuitanSubOrderID string     `gorm:"column:jushuitan_sub_order_id;size:50;null;comment:聚水潭子订单ID(oi_id)" json:"jushuitan_sub_order_id"`          //聚水潭子订单ID
	WmsCoID             string     `gorm:"column:wms_co_id;size:50;null;comment:仓库编码(wms_co_id)" json:"wms_co_id"`                                    //仓库编码
	Qty                 int        `gorm:"column:qty;default:0;comment:商品数量(qty)" json:"qty"`                                                         //商品数量
	CommodityID         string     `gorm:"column:commodity_id;size:50;null;comment:商品编码(sku_id)" json:"commodity_id"`                                 //商品编码
	CreateTime          time.Time  `gorm:"column:create_time;autoCreateTime;comment:创建时间" json:"create_time"`                                         //创建时间
	UpdateTime          time.Time  `gorm:"column:update_time;autoUpdateTime;comment:更新时间" json:"update_time"`                                         //更新时间
}

// TableName 指定对应的数据库表名
func (SubOrder) TableName() string {
	return "sub_order_data"
}

// BeforeSave GORM保存前钩子，可用于数据预处理或验证
func (s *SubOrder) BeforeSave(*gorm.DB) error {
	return nil
}
