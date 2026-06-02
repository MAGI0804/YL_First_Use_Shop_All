package models

import (
	"time"

	"gorm.io/gorm"
)

// Commodity 商品模型
type Commodity struct {
	CommodityID    string    `gorm:"column:commodity_id;primaryKey;size:100;comment:商品ID" json:"commodity_id"`
	Name           string    `gorm:"column:name;size:255;not null;comment:商品名称" json:"name"`
	StyleCode      string    `gorm:"column:style_code;size:50;index;comment:款号" json:"style_code"`
	Category       string    `gorm:"column:category;size:100;not null;comment:分类" json:"category"`
	CategoryDetail string    `gorm:"column:category_detail;size:100;null;comment:详细分类" json:"category_detail"`
	Price          float64   `gorm:"column:price;not null;comment:价格" json:"price"`
	Image          string    `gorm:"column:image;size:255;not null;comment:图片" json:"image"`
	PromoImage     string    `gorm:"column:promo_image;size:255;null;comment:促销图片" json:"promo_image"`
	Size           string    `gorm:"column:size;size:50;null;comment:尺寸" json:"size"`
	Color          string    `gorm:"column:color;size:50;null;comment:颜色" json:"color"`
	Height         string    `gorm:"column:height;size:50;null;comment:高度" json:"height"`
	SpecCode       string    `gorm:"column:spec_code;size:100;null;comment:规格码" json:"spec_code"`
	ColorImage     string    `gorm:"column:color_image;size:255;null;comment:颜色图片" json:"color_image"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	Inventory      int       `gorm:"column:inventory;default:0;comment:库存" json:"inventory"`
	Notes          string    `gorm:"column:notes;type:text;null;comment:备注" json:"notes"`
}

// TableName 设置表名，与Django模型保持一致
func (Commodity) TableName() string {
	return "Commodity_data"
}

// BeforeSave GORM钩子，确保gorm包被使用
func (c *Commodity) BeforeSave(*gorm.DB) error {
	return nil
}
