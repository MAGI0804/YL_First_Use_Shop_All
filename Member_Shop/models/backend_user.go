package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// BackendUser stores staff accounts for the web management backend.
type BackendUser struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperatorNo  string    `gorm:"column:operator_no;size:32;uniqueIndex;not null;comment:backend operator number" json:"operator_no"`
	Nickname    string    `gorm:"column:nickname;size:100;not null;comment:display name" json:"nickname"`
	Mobile      string    `gorm:"column:mobile;size:11;uniqueIndex;not null;comment:login mobile" json:"mobile"`
	Password    string    `gorm:"column:password;size:255;not null;comment:bcrypt password" json:"-"`
	Role        string    `gorm:"column:role;size:32;default:operation;comment:operation,customer_service,admin" json:"role"`
	Level       int       `gorm:"column:level;default:1;comment:legacy permission level" json:"-"`
	Permissions string    `gorm:"column:permissions;type:text;null;comment:allowed web pages" json:"permissions"`
	Status      string    `gorm:"column:status;size:20;default:active;comment:account status" json:"status"`
	Remarks     string    `gorm:"column:remarks;type:text;null;comment:remarks" json:"remarks"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (BackendUser) TableName() string {
	return "backend_operation_user"
}

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

func generateOperatorNo() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return fmt.Sprintf("OP%s0000", time.Now().Format("20060102150405"))
	}
	return fmt.Sprintf("OP%s%04d", time.Now().Format("20060102150405"), n.Int64())
}
