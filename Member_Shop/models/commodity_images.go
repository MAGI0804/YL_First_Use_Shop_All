package models

import "time"

// CommodityImage 商品图片模型
type CommodityImage struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CommodityID string    `gorm:"column:commodity_id;size:100;index;comment:商品ID" json:"commodity_id"`
	Image       string    `gorm:"column:image;size:255;not null;comment:图片地址" json:"image"`
	IsMain      bool      `gorm:"column:is_main;default:false;comment:是否主图" json:"is_main"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

// TableName 设置表名
func (CommodityImage) TableName() string {
	return "Commodity_Images"
}
