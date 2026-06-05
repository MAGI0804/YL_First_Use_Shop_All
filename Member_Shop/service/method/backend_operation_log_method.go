package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	OperationModuleMember = "member"
	OperationModuleCart   = "cart"
	OperationModuleOrder  = "order"

	ActionMemberCreate              = "member.create"
	ActionMemberUpdate              = "member.update"
	ActionMemberTagSet              = "member.tag.set"
	ActionMemberCartAdd             = "member.cart.add"
	ActionMemberCartUpdateQty       = "member.cart.update_quantity"
	ActionMemberCartDelete          = "member.cart.delete"
	ActionMemberCartClearAfterOrder = "member.cart.clear_after_order"
	ActionOrderBackendCreate        = "order.backend_create"
	ActionOrderPaymentAmountUpdate  = "order.payment_amount.update"
	ActionOrderPaymentConfirm       = "order.payment.confirm"
	ActionOrderDeliver              = "order.deliver"
	ActionOrderReceive              = "order.receive"
)

type BackendOperatorSnapshot struct {
	ID         uint
	OperatorNo string
	Mobile     string
	Role       string
}

type BackendOperationLogInput struct {
	Operator   BackendOperatorSnapshot
	Action     string
	Module     string
	TargetType string
	TargetID   string
	MemberID   uint
	UserID     int
	OrderID    string
	BeforeData any
	AfterData  any
	RequestID  string
	ClientIP   string
	UserAgent  string
	Remark     string
}

func BackendOperatorFromContext(c *gin.Context) (*models.BackendUser, error) {
	userValue, ok := c.Get("backendUser")
	if !ok {
		return nil, fmt.Errorf("backend user missing")
	}
	user, ok := userValue.(*models.BackendUser)
	if !ok || user == nil {
		return nil, fmt.Errorf("backend user invalid")
	}
	return user, nil
}

func BuildBackendOperatorSnapshot(user *models.BackendUser) BackendOperatorSnapshot {
	if user == nil {
		return BackendOperatorSnapshot{}
	}
	return BackendOperatorSnapshot{
		ID:         user.ID,
		OperatorNo: user.OperatorNo,
		Mobile:     user.Mobile,
		Role:       user.Role,
	}
}

func RecordBackendOperation(input BackendOperationLogInput) error {
	return recordBackendOperation(db.DB, input)
}

func recordBackendOperation(tx *gorm.DB, input BackendOperationLogInput) error {
	if tx == nil {
		return fmt.Errorf("db is nil")
	}
	if input.Operator.ID == 0 {
		return fmt.Errorf("operator is required")
	}
	beforeData, err := encodeOperationPayload(input.BeforeData)
	if err != nil {
		return err
	}
	afterData, err := encodeOperationPayload(input.AfterData)
	if err != nil {
		return err
	}
	log := models.BackendOperationLog{
		OperatorID:     input.Operator.ID,
		OperatorNo:     input.Operator.OperatorNo,
		OperatorMobile: input.Operator.Mobile,
		OperatorRole:   input.Operator.Role,
		Action:         strings.TrimSpace(input.Action),
		Module:         strings.TrimSpace(input.Module),
		TargetType:     strings.TrimSpace(input.TargetType),
		TargetID:       strings.TrimSpace(input.TargetID),
		MemberID:       input.MemberID,
		UserID:         input.UserID,
		OrderID:        strings.TrimSpace(input.OrderID),
		BeforeData:     beforeData,
		AfterData:      afterData,
		RequestID:      strings.TrimSpace(input.RequestID),
		ClientIP:       strings.TrimSpace(input.ClientIP),
		UserAgent:      strings.TrimSpace(input.UserAgent),
		Remark:         strings.TrimSpace(input.Remark),
	}
	if log.Action == "" || log.Module == "" {
		return fmt.Errorf("operation action and module are required")
	}
	return tx.Create(&log).Error
}

func encodeOperationPayload(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func QueryBackendOperationLogs(req requestbody.OperationLogQueryRequest) ([]models.BackendOperationLog, int64, error) {
	page, pageSize := normalizeBackendPage(req.Page, req.PageSize)
	query := db.DB.Model(&models.BackendOperationLog{})
	if req.OperatorID > 0 {
		query = query.Where("operator_id = ?", req.OperatorID)
	}
	if strings.TrimSpace(req.Action) != "" {
		query = query.Where("action = ?", strings.TrimSpace(req.Action))
	}
	if strings.TrimSpace(req.Module) != "" {
		query = query.Where("module = ?", strings.TrimSpace(req.Module))
	}
	if strings.TrimSpace(req.TargetType) != "" {
		query = query.Where("target_type = ?", strings.TrimSpace(req.TargetType))
	}
	if strings.TrimSpace(req.TargetID) != "" {
		query = query.Where("target_id = ?", strings.TrimSpace(req.TargetID))
	}
	if req.MemberID > 0 {
		query = query.Where("member_id = ?", req.MemberID)
	}
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if strings.TrimSpace(req.OrderID) != "" {
		query = query.Where("order_id = ?", strings.TrimSpace(req.OrderID))
	}
	if t, ok := parseQueryTime(req.BeginTime, false); ok {
		query = query.Where("created_at >= ?", t)
	}
	if t, ok := parseQueryTime(req.EndTime, true); ok {
		query = query.Where("created_at < ?", t)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []models.BackendOperationLog
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func normalizeBackendPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func parseQueryTime(value string, endExclusive bool) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		if endExclusive {
			t = t.Add(24 * time.Hour)
		}
		return t, true
	}
	return time.Time{}, false
}
