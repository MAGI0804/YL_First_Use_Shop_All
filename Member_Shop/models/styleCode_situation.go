package models

import "time"

// StyleCodeSituation 款式状态模型
type StyleCodeSituation struct {
	StyleCode     string     `gorm:"column:style_code;primaryKey;size:50;comment:款号" json:"style_code"`
	Status        string     `gorm:"column:status;size:20;not null;index;comment:状态" json:"status"`
	OnlineTime    *time.Time `gorm:"column:online_time;autoCreateTime;comment:上线时间" json:"online_time"`
	OfflineTime   *time.Time `gorm:"column:offline_time;null;comment:下线时间" json:"offline_time"`
	SyncDataCount int        `gorm:"column:sync_data_count;default:0;comment:同步数据次数" json:"sync_data_count"`
}

// TableName 设置表名
func (StyleCodeSituation) TableName() string {
	return "StyleCode_Situation"
}
