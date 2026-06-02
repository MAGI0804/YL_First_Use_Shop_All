package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitAnalyticsRoutes 注册数据分析接口。
// 路由统一挂在 /analytics/ 下，方便前端按模块调用和后续加权限中间件。
func InitAnalyticsRoutes(router *gin.Engine) {
	analyticsController := &controllers.AnalyticsController{}

	analyticsGroup := router.Group("/analytics/")
	{
		analyticsGroup.POST("sales_summary", analyticsController.SalesSummary)     // 销售统计
		analyticsGroup.POST("user_summary", analyticsController.UserSummary)       // 用户分析
		analyticsGroup.POST("product_summary", analyticsController.ProductSummary) // 商品分析
		analyticsGroup.POST("traffic_summary", analyticsController.TrafficSummary) // 流量分析预留
		analyticsGroup.POST("export", analyticsController.Export)                  // 结构化导出
	}
}
