package models

import "time"

// Message 消息模型
// 用于存储用户的消息通知，包括系统通知、订单通知等各类消息
type Message struct {
	MessageID       string    `column:"message_id" gorm:"primaryKey"`                                        //消息ID
	UserID          int64     `column:"user_id" gorm:"primaryKey"`                                           //用户ID
	MessageType     string    `column:"message_type" gorm:"type:varchar(255);not null"`                      //消息类型
	MessageTitleOne string    `column:"message_title_one" gorm:"type:varchar(255);not null"`                 //消息标题一级
	MessageTitleTwo string    `column:"message_title_two" gorm:"type:varchar(255);not null"`                 //消息标题二级
	MessageBody     string    `column:"message_body" gorm:"type:text;not null"`                              //消息内容
	RelatedNum      string    `column:"related_num" gorm:"type:varchar(255);not null"`                       //相关单号
	DisplayImg      string    `column:"display_img" gorm:"type:varchar(255);not null"`                       //显示图片
	CreatedAt       time.Time `column:"created_at" gorm:"type:timestamp;not null;default:current_timestamp"` //创建时间
}

// TableName 指定对应的数据库表名
func (Message) TableName() string {
	return "messages_data"
}
