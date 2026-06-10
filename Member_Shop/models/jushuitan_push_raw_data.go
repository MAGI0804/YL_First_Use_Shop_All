package models

import (
	"time"

	"gorm.io/gorm"
)

// JushuitanPushRawData 聚水潭推送原始数据模型
// 用于存储从聚水潭系统推送过来的原始请求和响应数据，便于问题排查和数据追溯
type JushuitanPushRawData struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement;comment:序号/主键ID" json:"id"`
	RequestURL  string     `gorm:"column:request_url;size:500;null;comment:请求URL" json:"request_url"`
	RequestIP   string     `gorm:"column:request_ip;size:50;null;comment:请求IP" json:"request_ip"`
	RequestTime *time.Time `gorm:"column:request_time;null;comment:请求时间" json:"request_time"`
	Response    string     `gorm:"column:response;type:text;null;comment:响应情况/响应结果" json:"response"`
	RawData     string     `gorm:"column:raw_data;type:longtext;null;comment:原始数据/请求body" json:"raw_data"`
	Remarks     string     `gorm:"column:remarks;size:500;null;comment:备注" json:"remarks"`
	CreateTime  time.Time  `gorm:"column:create_time;autoCreateTime;comment:创建时间" json:"create_time"`
}

// TableName 指定对应的数据库表名
func (JushuitanPushRawData) TableName() string {
	return "jushuitan_push_raw_data"
}

// BeforeSave GORM保存前钩子，可用于数据预处理或验证
func (j *JushuitanPushRawData) BeforeSave(tx *gorm.DB) error {
	return nil
}
