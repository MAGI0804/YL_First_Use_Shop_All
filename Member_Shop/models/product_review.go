package models

import "time"

const (
	ReviewStatusPending  = "pending"
	ReviewStatusApproved = "approved"
	ReviewStatusRejected = "rejected"
	ReviewStatusHidden   = "hidden"
)

// ProductReview stores a user review for a purchased commodity.
type ProductReview struct {
	ID            uint          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID        int           `gorm:"column:user_id;not null;index" json:"user_id"`
	OrderID       string        `gorm:"column:order_id;size:20;not null;index" json:"order_id"`
	SubOrderID    string        `gorm:"column:sub_order_id;size:30;not null;uniqueIndex:idx_review_sub_order_commodity" json:"sub_order_id"`
	CommodityID   string        `gorm:"column:commodity_id;size:100;not null;index;uniqueIndex:idx_review_sub_order_commodity" json:"commodity_id"`
	StyleCode     string        `gorm:"column:style_code;size:50;index" json:"style_code"`
	Rating        int           `gorm:"column:rating;not null;default:5" json:"rating"`
	Content       string        `gorm:"column:content;type:text" json:"content"`
	Images        string        `gorm:"column:images;type:text" json:"images"`
	Tags          string        `gorm:"column:tags;type:text" json:"tags"`
	Status        string        `gorm:"column:status;size:20;not null;default:'pending';index" json:"status"`
	AuditRemark   string        `gorm:"column:audit_remark;size:255" json:"audit_remark"`
	CreatedAt     time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	ReviewReplies []ReviewReply `gorm:"foreignKey:ReviewID;references:ID" json:"replies,omitempty"`
}

func (ProductReview) TableName() string {
	return "product_reviews"
}
