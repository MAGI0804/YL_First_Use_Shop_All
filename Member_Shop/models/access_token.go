package models

import (
	"time"

	_ "gorm.io/gorm"
)

// AccessToken 访问令牌模型
// 对应数据库表: access_token
type AccessToken struct {
	ID           int       `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	IPAddress    string    `gorm:"column:ip_address;type:varchar(45);not null;comment:IP地址" json:"ip_address"`
	AccessToken  string    `gorm:"column:access_token;type:varchar(100);uniqueIndex;not null;comment:访问令牌" json:"access_token"`
	RegisterTime time.Time `gorm:"column:register_time;autoCreateTime;comment:注册时间" json:"register_time"`
}

// TableName 设置表名
func (AccessToken) TableName() string {
	return "access_token"
}
