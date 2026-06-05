package models

import "time"

// MemberTag is a backend-managed label that can be assigned to members.
type MemberTag struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:50;uniqueIndex;not null;comment:标签名称" json:"name"`
	Color     string    `gorm:"column:color;size:20;null;comment:标签颜色" json:"color"`
	Remarks   string    `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
	CreatedBy uint      `gorm:"column:created_by;default:0;comment:创建操作人ID" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (MemberTag) TableName() string {
	return "member_tag"
}

// MemberTagRelation links a member to a custom backend tag.
type MemberTagRelation struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberID  uint      `gorm:"column:member_id;uniqueIndex:idx_member_tag_relation;index;not null" json:"member_id"`
	TagID     uint      `gorm:"column:tag_id;uniqueIndex:idx_member_tag_relation;index;not null" json:"tag_id"`
	CreatedBy uint      `gorm:"column:created_by;default:0;comment:操作人ID" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Tag       MemberTag `gorm:"foreignKey:TagID;references:ID" json:"tag,omitempty"`
}

func (MemberTagRelation) TableName() string {
	return "member_tag_relation"
}
