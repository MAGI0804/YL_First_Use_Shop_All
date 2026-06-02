package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// BackendUser 存储后台运营账号
// 这些账号是给员工/运营人员使用的，不是会员或微信用户
type BackendUser struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperatorNo string    `gorm:"column:operator_no;size:32;uniqueIndex;not null;comment:后台运营编号" json:"operator_no"`
	Nickname   string    `gorm:"column:nickname;size:100;not null;comment:昵称" json:"nickname"`
	Mobile     string    `gorm:"column:mobile;size:11;uniqueIndex;not null;comment:登录手机号" json:"mobile"`
	Password   string    `gorm:"column:password;size:255;not null;com                                                                            ment:bcrypt加密密码" json:"-"`
	Role       string    `gorm:"column:role;size:32;default:operation;comment:后台角色" json:"role"`
	Level      int       `gorm:"column:level;default:1;comment:权限等级" json:"level"`
	Status     string    `gorm:"column:status;size:20;default:active;comment:账号状态" json:"status"`
	Remarks    string    `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (BackendUser) TableName() string {
	return "backend_operation_user"
}

// BeforeCreate 填充后台账号的默认值
func (u *BackendUser) BeforeCreate(tx *gorm.DB) error {
	if u.OperatorNo == "" {
		u.OperatorNo = generateOperatorNo()
	}
	if u.Role == "" {
		u.Role = "operation"
	}
	if u.Status == "" {
		u.Status = "active"
	}
	if u.Level == 0 {
		u.Level = 1
	}
	return nil
}

// generateOperatorNo 为后台用户生成一个易读的运营编号
func generateOperatorNo() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return fmt.Sprintf("OP%s0000", time.Now().Format("20060102150405"))
	}
	return fmt.Sprintf("OP%s%04d", time.Now().Format("20060102150405"), n.Int64())
}
