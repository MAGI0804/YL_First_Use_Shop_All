package models

import "time"

// StyleCodeData 款式数据模型
type StyleCodeData struct {
	StyleCode       string    `gorm:"column:style_code;primaryKey;size:50;comment:款号" json:"style_code"`
	Name            string    `gorm:"column:name;size:255;not null;comment:名称" json:"name"`
	Image           string    `gorm:"column:image;size:255;not null;comment:图片" json:"image"`
	Category        string    `gorm:"column:category;size:100;not null;index;comment:分类" json:"category"`
	CategoryDetail  string    `gorm:"column:category_detail;size:100;null;comment:详细分类" json:"category_detail"`
	Price           float64   `gorm:"column:price;not null;comment:价格" json:"price"`
	DisplayPictures string    `gorm:"column:display_pictures;type:json;null;comment:展示图片" json:"display_pictures"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	LabelOne        string    `gorm:"column:label_one;size:50;null;index;comment:标签一" json:"label_one"`
	LabelTwo        string    `gorm:"column:label_two;size:50;null;index;comment:标签二" json:"label_two"`
	LabelThree      string    `gorm:"column:label_three;size:50;null;index;comment:标签三" json:"label_three"`
	LabelFour       string    `gorm:"column:label_four;size:50;null;index;comment:标签四" json:"label_four"`
	LabelFive       string    `gorm:"column:label_five;size:50;null;comment:标签五" json:"label_five"`
	LabelSix        string    `gorm:"column:label_six;size:50;null;comment:标签六" json:"label_six"`
	LabelSeven      string    `gorm:"column:label_seven;size:50;null;index;comment:标签七" json:"label_seven"`
	Inventory       int       `gorm:"column:inventory;default:0;comment:款式库存" json:"inventory"`
}

// TableName 设置表名
func (StyleCodeData) TableName() string {
	return "StyleCode_Data"
}
