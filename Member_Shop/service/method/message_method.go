package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"fmt"
	"time"
)

// CreateMessage 创建消息
func CreateMessage(userID int, messageType, messageTitleOne, messageTitleTwo, messageBody, relatedNum, displayImg string) error {
	// 生成消息ID
	messageID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 创建消息记录
	message := models.Message{
		MessageID:       messageID,
		UserID:          int64(userID),
		MessageType:     messageType,
		MessageTitleOne: messageTitleOne,
		MessageTitleTwo: messageTitleTwo,
		MessageBody:     messageBody,
		RelatedNum:      relatedNum,
		DisplayImg:      displayImg,
		CreatedAt:       time.Now(),
	}

	// 保存到数据库
	if err := db.DB.Create(&message).Error; err != nil {
		return err
	}

	return nil
}

// CreateOrderMessage 创建订单相关消息
func CreateOrderMessage(userID int, orderID, status string) error {
	var messageTitleOne, messageTitleTwo, messageBody string

	switch status {
	case "created":
		messageTitleOne = "订单创建完成"
		messageTitleTwo = "订单通知"
		messageBody = fmt.Sprintf("您的订单已创建成功，创建时间为: %s", time.Now().Format("2006-01-02 15:04:05"))
	case "shipped":
		messageTitleOne = "订单已发货"
		messageTitleTwo = "订单通知"
		messageBody = "您的订单已发货"
	case "delivered":
		messageTitleOne = "订单已签收"
		messageTitleTwo = "订单通知"
		messageBody = "您的订单已签收"
	case "canceled":
		messageTitleOne = "订单已取消"
		messageTitleTwo = "订单通知"
		messageBody = "您的订单已取消"
	case "return_requested":
		messageTitleOne = "已申请售后"
		messageTitleTwo = "订单通知"
		messageBody = "您已成功申请售后"
	default:
		return nil
	}

	return CreateMessage(userID, "Order", messageTitleOne, messageTitleTwo+orderID, messageBody, orderID, "")
}

// CreateReturnOrderMessage 创建退货订单相关消息
func CreateReturnOrderMessage(userID int, returnOrderID, status string) error {
	var messageTitleOne, messageTitleTwo, messageBody string

	switch status {
	case "created":
		messageTitleOne = "退货订单已创建"
		messageTitleTwo = "退货订单号:"
		messageBody = "您的退货订单已创建成功"
	case "shipped":
		messageTitleOne = "退货订单已发货"
		messageTitleTwo = "退货订单号:"
		messageBody = "您的退货订单已发货"
	case "completed":
		messageTitleOne = "退货订单已完成"
		messageTitleTwo = "退货订单号:"
		messageBody = "您的退货订单已完成"
	case "canceled":
		messageTitleOne = "退货订单已取消"
		messageTitleTwo = "退货订单号:"
		messageBody = "您的退货订单已取消"
	default:
		return nil
	}

	return CreateMessage(userID, "return_order", messageTitleOne, messageTitleTwo+returnOrderID, messageBody, returnOrderID, "")
}

// GetMessageCategories 查询消息分类和该分类下的最后一条消息
func GetMessageCategories(userID int) ([]map[string]interface{}, error) {
	// 查询用户的消息分类
	var messageTypes []string
	if err := db.DB.Table("messages_data").Where("user_id = ?", userID).Distinct("message_type").Pluck("message_type", &messageTypes).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(messageTypes))

	// 对每个分类查询最后一条消息
	for _, messageType := range messageTypes {
		var lastMessage models.Message
		if err := db.DB.Table("messages_data").Where("user_id = ? AND message_type = ?", userID, messageType).Order("created_at DESC").First(&lastMessage).Error; err != nil {
			continue
		}

		categoryInfo := map[string]interface{}{
			"message_type":      messageType,
			"last_message":      lastMessage.MessageBody,
			"last_message_time": lastMessage.CreatedAt,
			"related_num":       lastMessage.RelatedNum,
			"display_img":       lastMessage.DisplayImg,
		}

		result = append(result, categoryInfo)
	}

	return result, nil
}

// GetMessagesByType 根据分类和用户ID查询消息
func GetMessagesByType(userID int, messageType string, page, pageSize int) ([]models.Message, int64, error) {
	var messages []models.Message
	query := db.DB.Table("messages_data").Where("user_id = ? AND message_type = ?", userID, messageType)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}
