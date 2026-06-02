package models

import (
	"time"

	"gorm.io/gorm"
)

// Address 用户地址模型
type Address struct {
	AddressID       int       `gorm:"column:address_id;primaryKey;autoIncrement;comment:地址ID" json:"address_id"`
	UserID          int       `gorm:"column:user_id;index;comment:用户ID" json:"user_id"`
	ReceiverName    string    `gorm:"column:receiver_name;size:100;not null;comment:收货人姓名" json:"receiver_name"`
	PhoneNumber     string    `gorm:"column:phone_number;size:20;not null;comment:电话号码" json:"phone_number"`
	Province        string    `gorm:"column:province;size:50;not null;comment:省份" json:"province"`
	City            string    `gorm:"column:city;size:50;not null;comment:城市" json:"city"`
	County          string    `gorm:"column:county;size:50;not null;comment:区县" json:"county"`
	DetailedAddress string    `gorm:"column:detailed_address;size:255;not null;comment:详细地址" json:"detailed_address"`
	IsDefault       bool      `gorm:"column:is_default;default:false;comment:是否默认" json:"is_default"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 设置表名
func (Address) TableName() string {
	return "addresses"
}

// BeforeSave GORM钩子，确保gorm包被使用
func (a *Address) BeforeSave(*gorm.DB) error {
	return nil
}
