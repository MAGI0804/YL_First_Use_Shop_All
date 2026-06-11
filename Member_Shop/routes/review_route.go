package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitReviewRoutes 初始化评价相关路由
// 路由分组：/review/
// 包含：创建评价、查询商品评价、后台查询、审核评价、回复评价、统计评价
func InitReviewRoutes(router *gin.Engine) {
	reviewController := &controllers.ReviewController{}

	reviewGroup := router.Group("/review/")
	{
		reviewGroup.POST("create", reviewController.CreateReview)             // 创建评价
		reviewGroup.POST("query_by_product", reviewController.QueryByProduct) // 查询商品评价（前台）
		reviewGroup.POST("query_backend", reviewController.QueryBackend)      // 后台评价查询
		reviewGroup.POST("query_mine", reviewController.QueryMine)            // 我的评价列表
		reviewGroup.POST("update", reviewController.UpdateReview)             // 修改待审核评价
		reviewGroup.POST("delete", reviewController.DeleteReview)             // 软删除待审核评价
		reviewGroup.POST("audit", reviewController.AuditReview)               // 审核评价
		reviewGroup.POST("reply", reviewController.ReplyReview)               // 回复评价
		reviewGroup.POST("statistics", reviewController.ReviewStatistics)     // 评价统计
	}
}
