package models

import "time"

// CommoditySituation 商品状态模型
type CommoditySituation struct {
	CommodityID string     `gorm:"column:commodity_id;primaryKey;size:100;comment:商品ID" json:"commodity_id"`
	Status      string     `gorm:"column:status;size:20;not null;comment:状态" json:"status"`
	OnlineTime  time.Time  `gorm:"column:online_time;autoCreateTime;comment:上线时间" json:"online_time"`
	OfflineTime *time.Time `gorm:"column:offline_time;null;comment:下线时间" json:"offline_time"`
	SalesVolume int        `gorm:"column:sales_volume;default:0;comment:销量" json:"sales_volume"`
	Remarks     string     `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
	StyleCode   string     `gorm:"column:style_code;size:50;index;comment:款号" json:"style_code"`
	Category    string     `gorm:"column:category;size:100;not null;comment:分类" json:"category"`
	Inventory   int        `gorm:"column:inventory;default:0;comment:库存" json:"inventory"`
	LabelOne    string     `gorm:"column:label_one;size:50;null;comment:标签一" json:"label_one"`
	LabelTwo    string     `gorm:"column:label_two;size:50;null;comment:标签二" json:"label_two"`
	LabelThree  string     `gorm:"column:label_three;size:50;null;comment:标签三" json:"label_three"`
	LabelFour   string     `gorm:"column:label_four;size:50;null;comment:标签四" json:"label_four"`
	LabelFive   string     `gorm:"column:label_five;size:50;null;comment:标签五" json:"label_five"`
	LabelSix    string     `gorm:"column:label_six;size:50;null;comment:标签六" json:"label_six"`
	LabelSeven  string     `gorm:"column:label_seven;size:50;null;comment:标签七" json:"label_seven"`
}

// TableName 设置表名
func (CommoditySituation) TableName() string {
	return "Commodity_Situation"
}
