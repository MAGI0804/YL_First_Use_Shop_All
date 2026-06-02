package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// Member 存储会员业务数据
// 故意与User分离：
// - User存储微信身份
// - Member存储会员手机号、唯一会员号、订单总额和平台ID
type Member struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MemberNo         string    `gorm:"column:member_no;size:32;uniqueIndex;not null;comment:会员唯一编号" json:"member_no"`
	UserID           int       `gorm:"column:user_id;index;default:0;comment:关联的微信用户ID" json:"user_id"`
	OpenID           string    `gorm:"column:openid;size:100;index;null;comment:关联的微信OpenID" json:"openid"`
	Mobile           string    `gorm:"column:mobile;size:11;uniqueIndex;not null;comment:会员手机号" json:"mobile"`
	Nickname         string    `gorm:"column:nickname;size:100;null;comment:会员昵称" json:"nickname"`
	TotalOrderAmount float64   `gorm:"column:total_order_amount;type:decimal(10,2);default:0;comment:订单总金额" json:"total_order_amount"`
	TotalPaidAmount  float64   `gorm:"column:total_paid_amount;type:decimal(10,2);default:0;comment:实付总金额" json:"total_paid_amount"`
	TmallID          string    `gorm:"column:tmall_id;size:100;null;comment:天猫ID" json:"tmall_id"`
	TmallAmount      float64   `gorm:"column:tmall_amount;type:decimal(10,2);default:0;comment:天猫金额" json:"tmall_amount"`
	YouzanID         string    `gorm:"column:youzan_id;size:100;null;comment:有赞ID" json:"youzan_id"`
	YouzanAmount     float64   `gorm:"column:youzan_amount;type:decimal(10,2);default:0;comment:有赞金额" json:"youzan_amount"`
	Remarks          string    `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Member) TableName() string {
	return "member_info"
}

// BeforeCreate 在调用者未提供会员号时自动分配一个
func (m *Member) BeforeCreate(tx *gorm.DB) error {
	if m.MemberNo == "" {
		m.MemberNo = GenerateMemberNo()
	}
	return nil
}

// GenerateMemberNo 返回一个带日期前缀的、近乎唯一的会员号
func GenerateMemberNo() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return fmt.Sprintf("M%s0000", time.Now().Format("20060102150405"))
	}
	return fmt.Sprintf("M%s%04d", time.Now().Format("20060102150405"), n.Int64())
}
