package models

import "time"

// BackendOperationLog records every backend write action with the staff account that performed it.
type BackendOperationLog struct {
	ID             uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperatorID     uint      `gorm:"column:operator_id;index;not null;comment:后台操作人ID" json:"operator_id"`
	OperatorNo     string    `gorm:"column:operator_no;size:32;index;null;comment:后台操作人工号" json:"operator_no"`
	OperatorMobile string    `gorm:"column:operator_mobile;size:11;index;null;comment:后台操作人手机号" json:"operator_mobile"`
	OperatorRole   string    `gorm:"column:operator_role;size:32;null;comment:后台操作人角色" json:"operator_role"`
	Action         string    `gorm:"column:action;size:80;index;not null;comment:操作动作" json:"action"`
	Module         string    `gorm:"column:module;size:50;index;not null;comment:业务模块" json:"module"`
	TargetType     string    `gorm:"column:target_type;size:50;index;null;comment:操作对象类型" json:"target_type"`
	TargetID       string    `gorm:"column:target_id;size:80;index;null;comment:操作对象ID" json:"target_id"`
	MemberID       uint      `gorm:"column:member_id;index;default:0;comment:关联会员ID" json:"member_id"`
	UserID         int       `gorm:"column:user_id;index;default:0;comment:关联微信用户ID" json:"user_id"`
	OrderID        string    `gorm:"column:order_id;size:30;index;null;comment:关联订单号" json:"order_id"`
	BeforeData     string    `gorm:"column:before_data;type:text;null;comment:操作前关键数据JSON" json:"before_data"`
	AfterData      string    `gorm:"column:after_data;type:text;null;comment:操作后关键数据JSON" json:"after_data"`
	RequestID      string    `gorm:"column:request_id;size:64;index;null;comment:请求ID" json:"request_id"`
	ClientIP       string    `gorm:"column:client_ip;size:64;null;comment:客户端IP" json:"client_ip"`
	UserAgent      string    `gorm:"column:user_agent;size:255;null;comment:User-Agent" json:"user_agent"`
	Remark         string    `gorm:"column:remark;type:text;null;comment:备注" json:"remark"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (BackendOperationLog) TableName() string {
	return "backend_operation_log"
}
