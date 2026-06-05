package routes

import (
	"Member_shop/controllers"
	"Member_shop/middleware"

	"github.com/gin-gonic/gin"
)

// InitDownloadCenterRoutes registers web-admin download center APIs.
func InitDownloadCenterRoutes(router *gin.Engine) {
	downloadCenterController := &controllers.DownloadCenterController{}

	downloadCenterGroup := router.Group("/download_center")
	downloadCenterGroup.Use(middleware.BackendAuthMiddleware())
	{
		downloadCenterGroup.POST("/tasks", downloadCenterController.CreateTask)
		downloadCenterGroup.GET("/tasks", downloadCenterController.ListTasks)
		downloadCenterGroup.GET("/tasks/:task_id", downloadCenterController.TaskDetail)
		downloadCenterGroup.GET("/tasks/:task_id/file", downloadCenterController.DownloadFile)
		downloadCenterGroup.POST("/tasks/:task_id/retry", downloadCenterController.RetryTask)
		downloadCenterGroup.GET("/templates", downloadCenterController.ListTemplates)
	}
}
