package models

import (
	"time"

	"gorm.io/gorm"
)

// User 存储微信/小程序用户身份
// 保持此模型专注于原始的 users_user 表：
// - 微信OpenID、头像、昵称，以及可选的授权手机号
// - 收件人/地址字段保留以兼容现有调用者
// - 会员业务数据存储在Member中
// - 后台运营账号存储在BackendUser中
type User struct {
	UserID           int        `gorm:"column:user_id;primaryKey;autoIncrement;comment:用户ID" json:"user_id"`
	OpenID           string     `gorm:"column:openid;size:100;uniqueIndex;null;comment:微信OpenID" json:"openid"`
	UserImg          string     `gorm:"column:user_img;size:255;null;comment:用户头像" json:"user_img"`
	Mobile           string     `gorm:"column:mobile;size:11;uniqueIndex;null;comment:微信授权手机号（如有）" json:"mobile"`
	Nickname         string     `gorm:"column:nickname;size:100;not null;comment:微信昵称" json:"nickname"`
	Password         string     `gorm:"column:password;size:128;null;comment:密码" json:"password"`
	DefaultReceiver  string     `gorm:"column:default_receiver;size:100;null;comment:默认收件人" json:"default_receiver"`
	Province         string     `gorm:"column:province;size:50;null;comment:省份" json:"province"`
	City             string     `gorm:"column:city;size:50;null;comment:城市" json:"city"`
	County           string     `gorm:"column:county;size:50;null;comment:区县" json:"county"`
	DetailedAddress  string     `gorm:"column:detailed_address;size:255;null;comment:详细地址" json:"detailed_address"`
	MembershipLevel  int        `gorm:"column:membership_level;default:0;comment:旧版会员等级" json:"membership_level"`
	RegistrationDate time.Time  `gorm:"column:registration_date;autoCreateTime;comment:注册日期" json:"registration_date"`
	TotalSpending    float64    `gorm:"column:total_spending;type:decimal(10,2);default:0;comment:旧版总消费额" json:"total_spending"`
	Remarks          string     `gorm:"column:remarks;type:text;null;comment:备注" json:"remarks"`
	LastLogin        *time.Time `gorm:"column:last_login;null;comment:最后登录时间" json:"last_login"`
	IsActive         bool       `gorm:"column:is_active;default:true;comment:是否激活" json:"is_active"`
	IsStaff          bool       `gorm:"column:is_staff;default:false;comment:是否员工" json:"is_staff"`
	IsSuperuser      bool       `gorm:"column:is_superuser;default:false;comment:是否超级用户" json:"is_superuser"`
}

func (User) TableName() string {
	return "users_user"
}

// BeforeCreate 在微信未提供昵称时填充一个可读的昵称
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.OpenID != "" && u.Nickname == "" {
		if len(u.OpenID) > 8 {
			u.Nickname = "wechat_user_" + u.OpenID[:8]
		} else {
			u.Nickname = "wechat_user_" + u.OpenID
		}
	} else if u.Mobile != "" && u.Nickname == "" {
		if len(u.Mobile) > 4 {
			u.Nickname = "mobile_user_" + u.Mobile[len(u.Mobile)-4:]
		} else {
			u.Nickname = "mobile_user_" + u.Mobile
		}
	}
	u.RegistrationDate = time.Now()
	return nil
}

// BeforeSave 保留旧版 last_login 自动填充行为
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.LastLogin == nil {
		now := time.Now()
		u.LastLogin = &now
	}
	return nil
}
