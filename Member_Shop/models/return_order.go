package models

import (
	"time"
)

// ReturnOrder 退换货订单模型
type ReturnOrder struct {
	ID                  uint       `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID              int        `gorm:"column:user_id;not null;default:0" json:"user_id"`
	ReturnID            string     `gorm:"column:return_id;size:30;not null;uniqueIndex;comment:退换货订单号" json:"return_id"`
	OrderID             string     `gorm:"column:order_id;size:20;not null;comment:关联订单号" json:"order_id"`
	SubOrderID          string     `gorm:"column:sub_order_id;size:30;null;comment:关联子订单号" json:"sub_order_id"`
	SubOrderProductInfo string     `gorm:"column:sub_order_product_info;type:text;null;comment:子订单商品信息" json:"sub_order_product_info"`
	OrderStatus         string     `gorm:"column:order_status;size:20;not null;default:'pending';comment:状态" json:"order_status"`
	ProductList         string     `gorm:"column:product_list;type:text;not null;comment:商品列表" json:"product_list"`
	Type                string     `gorm:"column:type;size:20;not null;comment:类型" json:"type"`
	Status              string     `gorm:"column:status;size:20;not null;default:'pending';comment:状态" json:"status"`
	RequestTime         *time.Time `gorm:"column:request_time;autoCreateTime;comment:申请时间" json:"request_time"`
	ShippedTime         *time.Time `gorm:"column:shipped_time;null;comment:发货时间" json:"shipped_time"`
	CanceledTime        *time.Time `gorm:"column:canceled_time;null;comment:取消时间" json:"canceled_time"`
	CompletedTime       *time.Time `gorm:"column:completed_time;null;comment:完成时间" json:"completed_time"`
	ExpressCompany      string     `gorm:"column:express_company;size:50;null;comment:退货物流公司" json:"express_company"`
	ExpressNumber       string     `gorm:"column:express_number;size:50;null;comment:退货物流单号" json:"express_number"`
	Reason              string     `gorm:"column:reason;type:text;not null;comment:退换货原因" json:"reason"`
	SpecificReasons     string     `gorm:"column:specific_reasons;type:text;not null;comment:退换货具体原因" json:"specific_reasons"`
	BuyerProvince       string     `gorm:"column:buyer_province;size:50;null;comment:买方省" json:"buyer_province"`
	BuyerCity           string     `gorm:"column:buyer_city;size:50;null;comment:买方市" json:"buyer_city"`
	BuyerCounty         string     `gorm:"column:buyer_county;size:50;null;comment:买方县" json:"buyer_county"`
	BuyerAddress        string     `gorm:"column:buyer_address;size:255;null;comment:买方具体地址" json:"buyer_address"`
	BuyerPhone          string     `gorm:"column:buyer_phone;size:15;null;comment:买方联系电话" json:"buyer_phone"`
	Remarks             string     `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
}

// TableName 设置表名
func (ReturnOrder) TableName() string {
	return "return_order_data"
}
