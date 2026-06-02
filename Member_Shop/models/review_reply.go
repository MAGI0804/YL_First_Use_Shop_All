package models

import "time"

// ReviewReply stores a backend operator reply for a product review.
type ReviewReply struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReviewID   uint      `gorm:"column:review_id;not null;index" json:"review_id"`
	OperatorID string    `gorm:"column:operator_id;size:50;not null;index" json:"operator_id"`
	Content    string    `gorm:"column:content;type:text;not null" json:"content"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ReviewReply) TableName() string {
	return "review_replies"
}
