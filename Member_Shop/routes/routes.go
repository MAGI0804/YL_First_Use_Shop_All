package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine) {
	// 创建控制器实例
	accessTokenController := &controllers.AccessTokenController{}

	router.POST("access_token/get_token", accessTokenController.GetToken)
	router.POST("access_token/get_ips", accessTokenController.GetIPs)

	router.GET("api/test/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server is running"})
	})

	// 健康检查路由
	router.GET("api/health/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "页面不存在"})
	})

	// 405 路由
	router.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{"error": "请求方法不允许"})
	})
	// 初始化商品相关路由
	InitCommodityRoutes(router)
	InitInventoryRoutes(router)
	InitReviewRoutes(router)
	InitAnalyticsRoutes(router)

	// 初始化用户相关路由
	InitUserRoutes(router)

	// 初始化订单相关路由
	InitOrderRoutes(router)
	// 初始化运营用户相关路由
	InitOperationUserRoutes(router)

	// 初始化活动相关路由
	InitActivityRoutes(router)

	// 初始化购物车相关路由
	InitCartRoutes(router)

	// 初始化地址相关路由
	InitAddressRoutes(router)

	// 初始化退货订单相关路由
	InitReturnOrderRoutes(router)

	// 初始化消息相关路由
	InitMessageRoutes(router)
}
