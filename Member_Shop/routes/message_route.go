package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitMessageRoutes initializes message related routes
func InitMessageRoutes(router *gin.Engine) {
	// 初始化消息控制器
	messageController := &controllers.MessageController{}

	// 添加前置URL前缀
	messageGroup := router.Group("/message/")
	{
		// Query message categories and the last message in each category
		messageGroup.POST("categories", messageController.GetMessageCategories)
		// 根据分类和用户ID查询消息
		messageGroup.POST("query", messageController.GetMessagesByType)
		// Custom message creation
		messageGroup.POST("create", messageController.CreateMessage)
	}
}
