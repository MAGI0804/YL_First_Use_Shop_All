package models

import (
	"time"
)

// ActivityImage 活动图片模型
// 与Django项目中的ActivityImage模型完全同步
type ActivityImage struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Image            string     `gorm:"column:image;size:255;not null;comment:图片地址" json:"image"`
	Status           string     `gorm:"column:status;size:20;default:'pending';comment:状态" json:"status"`
	OnlineTime       *time.Time `gorm:"column:online_time;null;comment:上线时间" json:"online_time"`
	OfflineTime      *time.Time `gorm:"column:offline_time;null;comment:下线时间" json:"offline_time"`
	Commodities      string     `gorm:"column:commodities;type:text;null;comment:关联商品" json:"commodities"`
	StyleCodes       string     `gorm:"column:style_codes;size:100;null;comment:关联款式" json:"style_codes"`
	Category         string     `gorm:"column:category;size:100;null;comment:分类" json:"category"`
	Notes            string     `gorm:"column:notes;type:text;null;comment:备注" json:"notes"`
	PromotionalPics  string     `gorm:"column:promotional_pics;type:text;null;comment:宣传图信息" json:"promotional_pics"`
	HasActivityDetail bool       `gorm:"column:has_activity_detail;default:false;comment:是否有活动详情" json:"has_activity_detail"`
	Order            *int       `gorm:"column:order;null;comment:图片顺序" json:"order"` // 图片顺序字段，改为指针类型支持null
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 设置表名，与Django模型保持一致
func (ActivityImage) TableName() string {
	return "Activity_Image"
}
