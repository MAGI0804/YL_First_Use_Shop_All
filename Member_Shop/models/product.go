package models

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID        uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	Name      string    `gorm:"size:100;not null;comment:商品名称" json:"name"`
	Price     float64   `gorm:"not null;comment:商品价格" json:"price"`
	ImageURL  string    `gorm:"size:255;comment:图片URL" json:"image_url"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 设置表名
func (p *Product) TableName() string {
	return "product_product"
}
