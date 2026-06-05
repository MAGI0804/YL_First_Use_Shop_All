package routes

import (
	"Member_shop/controllers"
	"Member_shop/middleware"

	"github.com/gin-gonic/gin"
)

func InitOperationLogRoutes(router *gin.Engine) {
	operationLogController := &controllers.OperationLogController{}

	operationLogGroup := router.Group("/operation_log/")
	operationLogGroup.Use(middleware.BackendAuthMiddleware())
	{
		operationLogGroup.POST("query", operationLogController.Query)
	}
}
