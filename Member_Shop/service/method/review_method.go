package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReviewCreateInput 创建评价输入参数
type ReviewCreateInput struct {
	UserID      int      // 用户ID
	OrderID     string   // 订单ID
	SubOrderID  string   // 子订单ID
	CommodityID string   // 商品ID
	StyleCode   string   // 款式编码
	Rating      int      // 评分，1-5分
	Content     string   // 评价内容
	Images      []string // 评价图片列表
	Tags        []string // 评价标签列表
}

// ReviewProductQueryInput 商品评价查询输入参数
type ReviewProductQueryInput struct {
	CommodityID string // 商品ID
	StyleCode   string // 款式编码
	Page        int    // 页码
	PageSize    int    // 每页数量
}

// ReviewBackendQueryInput 后台评价管理查询输入参数
type ReviewBackendQueryInput struct {
	UserID      int    // 用户ID
	OrderID     string // 订单ID
	SubOrderID  string // 子订单ID
	CommodityID string // 商品ID
	StyleCode   string // 款式编码
	Status      string // 评价状态
	Page        int    // 页码
	PageSize    int    // 每页数量
}

// ReviewStatistics 评价统计数据结构
type ReviewStatistics struct {
	CommodityID        string        `json:"commodity_id,omitempty"` // 商品ID
	StyleCode          string        `json:"style_code,omitempty"`   // 款式编码
	Total              int64         `json:"total"`                  // 评价总数
	PendingCount       int64         `json:"pending_count"`          // 待审核评价数
	AverageRating      float64       `json:"average_rating"`         // 平均评分
	GoodRate           float64       `json:"good_rate"`              // 好评率（4-5分占比）
	RatingDistribution map[int]int64 `json:"rating_distribution"`    // 评分分布，1-5分各等级数量
}

// CreateReview 创建评价
// 验证用户、订单、子订单关系，检查是否已评价，检查子订单状态是否可评价
// 评价创建后状态为待审核（pending），需要后台审核后才能显示
func CreateReview(input ReviewCreateInput) (*models.ProductReview, error) {
	var review models.ProductReview
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		created, err := createReviewTx(tx, input)
		if err != nil {
			return err
		}
		review = *created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// QueryReviewsByProduct 查询商品评价（前台展示）
// 根据商品ID或款式编码查询已通过审核的评价列表
// 仅返回状态为已通过（approved）的评价，按时间倒序排列
func QueryReviewsByProduct(input ReviewProductQueryInput) ([]models.ProductReview, int64, int, int, error) {
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	input.StyleCode = strings.TrimSpace(input.StyleCode)
	if input.CommodityID == "" && input.StyleCode == "" {
		return nil, 0, 0, 0, fmt.Errorf("commodity_id or style_code is required")
	}

	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := db.DB.Model(&models.ProductReview{}).
		Preload("ReviewReplies").
		Where("status = ?", models.ReviewStatusApproved)
	if input.CommodityID != "" {
		query = query.Where("commodity_id = ?", input.CommodityID)
	}
	if input.StyleCode != "" {
		query = query.Where("style_code = ?", input.StyleCode)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	reviews := make([]models.ProductReview, 0)
	if err := query.Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&reviews).Error; err != nil {
		return nil, 0, page, pageSize, err
	}
	return reviews, total, page, pageSize, nil
}

// QueryReviewsForBackend 查询评价列表（后台管理）
// 支持多种筛选条件：用户ID、订单ID、子订单ID、商品ID、款式编码、评价状态
// 返回所有状态的评价，用于后台审核管理
func QueryReviewsForBackend(input ReviewBackendQueryInput) ([]models.ProductReview, int64, int, int, error) {
	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := db.DB.Model(&models.ProductReview{}).Preload("ReviewReplies")
	if input.UserID > 0 {
		query = query.Where("user_id = ?", input.UserID)
	}
	if strings.TrimSpace(input.OrderID) != "" {
		query = query.Where("order_id = ?", strings.TrimSpace(input.OrderID))
	}
	if strings.TrimSpace(input.SubOrderID) != "" {
		query = query.Where("sub_order_id = ?", strings.TrimSpace(input.SubOrderID))
	}
	if strings.TrimSpace(input.CommodityID) != "" {
		query = query.Where("commodity_id = ?", strings.TrimSpace(input.CommodityID))
	}
	if strings.TrimSpace(input.StyleCode) != "" {
		query = query.Where("style_code = ?", strings.TrimSpace(input.StyleCode))
	}
	if strings.TrimSpace(input.Status) != "" {
		status := normalizeReviewStatus(input.Status)
		if !validReviewStatus(status) {
			return nil, 0, page, pageSize, fmt.Errorf("invalid review status")
		}
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	reviews := make([]models.ProductReview, 0)
	if err := query.Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&reviews).Error; err != nil {
		return nil, 0, page, pageSize, err
	}
	return reviews, total, page, pageSize, nil
}

// AuditReview 审核评价
// 后台管理员审核评价，可将评价状态设置为：approved（通过）、rejected（拒绝）、hidden（隐藏）
// 使用行锁确保并发安全
func AuditReview(reviewID uint, status, remark string) (*models.ProductReview, error) {
	status = normalizeReviewStatus(status)
	if !validReviewStatus(status) {
		return nil, fmt.Errorf("invalid review status")
	}

	var review models.ProductReview
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", reviewID).
			First(&review).Error; err != nil {
			return err
		}
		review.Status = status
		review.AuditRemark = strings.TrimSpace(remark)
		return tx.Save(&review).Error
	})
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// ReplyReview 回复评价
// 运营人员或客服对用户评价进行回复
// 回复内容会关联到原评价上
func ReplyReview(reviewID uint, operatorID, content string) (*models.ReviewReply, error) {
	operatorID = strings.TrimSpace(operatorID)
	content = strings.TrimSpace(content)
	if operatorID == "" {
		return nil, fmt.Errorf("operator_id is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var reply models.ReviewReply
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var review models.ProductReview
		if err := tx.Where("id = ?", reviewID).First(&review).Error; err != nil {
			return err
		}

		reply = models.ReviewReply{
			ReviewID:   reviewID,
			OperatorID: operatorID,
			Content:    content,
		}
		return tx.Create(&reply).Error
	})
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

// GetReviewStatistics 获取评价统计数据
// 根据商品ID或款式编码统计评价数据
// 返回评价总数、平均评分、好评率、评分分布
func GetReviewStatistics(commodityID, styleCode string) (*ReviewStatistics, error) {
	commodityID = strings.TrimSpace(commodityID)
	styleCode = strings.TrimSpace(styleCode)

	stats := &ReviewStatistics{
		CommodityID:        commodityID,
		StyleCode:          styleCode,
		RatingDistribution: map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0},
	}
	if err := reviewStatisticsQuery(commodityID, styleCode).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	if err := reviewStatisticsStatusQuery(commodityID, styleCode, models.ReviewStatusPending).Count(&stats.PendingCount).Error; err != nil {
		return nil, err
	}
	if stats.Total == 0 {
		return stats, nil
	}

	var average float64
	if err := reviewStatisticsQuery(commodityID, styleCode).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&average).Error; err != nil {
		return nil, err
	}
	stats.AverageRating = average

	var goodCount int64
	if err := reviewStatisticsQuery(commodityID, styleCode).
		Where("rating >= ?", 4).
		Count(&goodCount).Error; err != nil {
		return nil, err
	}
	stats.GoodRate = float64(goodCount) / float64(stats.Total)

	type ratingCount struct {
		Rating int
		Total  int64
	}
	rows := make([]ratingCount, 0, 5)
	if err := reviewStatisticsQuery(commodityID, styleCode).
		Select("rating, COUNT(*) AS total").
		Group("rating").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		stats.RatingDistribution[row.Rating] = row.Total
	}
	return stats, nil
}

// createReviewTx 创建评价事务处理（内部函数）
// 核心评价创建逻辑，包含以下验证：
// 1. 验证用户存在；2. 验证订单存在且属于该用户；3. 验证子订单存在且属于该订单；
// 4. 验证子订单状态可评价（delivered/completed/signed/received）；5. 验证商品存在；
// 6. 验证该子订单商品未评价过；7. 创建评价记录
func createReviewTx(tx *gorm.DB, input ReviewCreateInput) (*models.ProductReview, error) {
	input.OrderID = strings.TrimSpace(input.OrderID)
	input.SubOrderID = strings.TrimSpace(input.SubOrderID)
	input.CommodityID = strings.TrimSpace(input.CommodityID)
	input.StyleCode = strings.TrimSpace(input.StyleCode)
	input.Content = strings.TrimSpace(input.Content)

	if input.UserID <= 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if input.Rating < 1 || input.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	var user models.User
	if err := tx.Where("user_id = ?", input.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var order models.Order
	if err := tx.Where("order_id = ?", input.OrderID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if order.UserID != input.UserID {
		return nil, fmt.Errorf("order does not belong to user")
	}

	var subOrder models.SubOrder
	if err := tx.Where("sub_order_id = ?", input.SubOrderID).First(&subOrder).Error; err != nil {
		return nil, fmt.Errorf("sub_order not found: %w", err)
	}
	if subOrder.OrderID != input.OrderID {
		return nil, fmt.Errorf("sub_order does not belong to order")
	}
	if !reviewableSubOrderStatus(subOrder.Status) {
		return nil, fmt.Errorf("sub_order status is not reviewable")
	}
	if subOrder.CommodityID != input.CommodityID {
		return nil, fmt.Errorf("commodity_id does not match sub_order")
	}

	var commodity models.Commodity
	if err := tx.Where("commodity_id = ?", input.CommodityID).First(&commodity).Error; err != nil {
		return nil, fmt.Errorf("commodity not found: %w", err)
	}
	if input.StyleCode != "" && commodity.StyleCode != "" && input.StyleCode != commodity.StyleCode {
		return nil, fmt.Errorf("style_code does not match commodity")
	}
	styleCode := commodity.StyleCode
	if styleCode == "" {
		styleCode = input.StyleCode
	}

	var existingCount int64
	if err := tx.Model(&models.ProductReview{}).
		Where("sub_order_id = ? AND commodity_id = ?", input.SubOrderID, input.CommodityID).
		Count(&existingCount).Error; err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, fmt.Errorf("sub_order commodity has already been reviewed")
	}

	images, err := marshalStringList(input.Images)
	if err != nil {
		return nil, err
	}
	tags, err := marshalStringList(input.Tags)
	if err != nil {
		return nil, err
	}

	review := models.ProductReview{
		UserID:      input.UserID,
		OrderID:     input.OrderID,
		SubOrderID:  input.SubOrderID,
		CommodityID: input.CommodityID,
		StyleCode:   styleCode,
		Rating:      input.Rating,
		Content:     input.Content,
		Images:      images,
		Tags:        tags,
		Status:      models.ReviewStatusPending,
	}
	if err := tx.Create(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// validReviewStatus 验证评价状态是否有效（内部函数）
// 支持的状态：pending（待审核）、approved（已通过）、rejected（已拒绝）、hidden（已隐藏）
func validReviewStatus(status string) bool {
	switch normalizeReviewStatus(status) {
	case models.ReviewStatusPending, models.ReviewStatusApproved, models.ReviewStatusRejected, models.ReviewStatusHidden:
		return true
	default:
		return false
	}
}

// normalizeReviewStatus 规范化评价状态（内部函数）
// 统一转为小写并去除空格
func normalizeReviewStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// reviewStatisticsQuery 构建评价统计查询（内部函数）
// 仅统计已通过审核（approved）的评价
func reviewStatisticsQuery(commodityID, styleCode string) *gorm.DB {
	return reviewStatisticsStatusQuery(commodityID, styleCode, models.ReviewStatusApproved)
}

func reviewStatisticsStatusQuery(commodityID, styleCode, status string) *gorm.DB {
	query := db.DB.Model(&models.ProductReview{}).Where("status = ?", status)
	if commodityID != "" {
		query = query.Where("commodity_id = ?", commodityID)
	}
	if styleCode != "" {
		query = query.Where("style_code = ?", styleCode)
	}
	return query
}

// reviewableSubOrderStatus 检查子订单状态是否可评价（内部函数）
// 只有已发货/已完成/已签收/已收货状态可以评价
func reviewableSubOrderStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "delivered", "completed", "signed", "received":
		return true
	default:
		return false
	}
}

// marshalStringList 将字符串列表序列化为JSON字符串（内部函数）
// 用于存储图片和标签列表到数据库
func marshalStringList(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return "", fmt.Errorf("marshal string list: %w", err)
	}
	return string(bytes), nil
}
