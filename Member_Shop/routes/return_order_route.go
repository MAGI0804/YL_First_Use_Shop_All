package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitReturnOrderRoutes 初始化退货订单相关路由
func InitReturnOrderRoutes(router *gin.Engine) {
	// 初始化退货订单控制器
	returnOrderController := &controllers.ReturnOrderController{}

	// 添加前置URL前缀
	returnOrderGroup := router.Group("/return_order/")
	{
		// 创建退货订单路由
		returnOrderGroup.POST("create", returnOrderController.CreateReturnOrder)
		// 退货订单发货路由
		returnOrderGroup.POST("deliver", returnOrderController.ReturnOrderDeliver)
		// 退货订单签收路由
		returnOrderGroup.POST("receive", returnOrderController.ReturnOrderReceive)
		// 退货订单取消路由
		returnOrderGroup.POST("cancel", returnOrderController.ReturnOrderCancel)
		// 退货订单修改买家信息路由
		returnOrderGroup.POST("update_buyer_info", returnOrderController.ReturnOrderUpdateBuyerInfo)
		// 退货订单审核路由
		returnOrderGroup.POST("approve", returnOrderController.ReturnOrderApprove)
		// 退货订单查询路由
		returnOrderGroup.POST("query", returnOrderController.QueryReturnOrder)
		// 退货订单详情路由
		returnOrderGroup.POST("detail", returnOrderController.GetReturnOrderDetail)
	}
}
