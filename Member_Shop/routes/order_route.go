package routes

import (
	"Member_shop/controllers"
	"Member_shop/middleware"

	"github.com/gin-gonic/gin"
)

// InitOrderRoutes 初始化订单相关路由 - 与Django版本order.urls完全匹配
func InitOrderRoutes(router *gin.Engine) {
	// 初始化订单控制器
	orderController := &controllers.OrderController{}

	// 添加前置URL前缀
	orderGroup := router.Group("/order/")
	{
		// 订单相关路由 - 与Django版本order.urls完全匹配
		orderGroup.POST("add_order", orderController.OrderCreate) // 创建订单
		orderGroup.POST("backend_create_order", middleware.BackendAuthMiddleware(), orderController.BackendCreateOrder)
		orderGroup.POST("query_order_data", orderController.OrderDetail)              // 查询订单信息
		orderGroup.POST("change_receiving_data", orderController.ChangeReceivingData) // 修改收货信息
		orderGroup.POST("change_status", orderController.ChangeStatus)                // 修改订单状态
		orderGroup.POST("orders_query", orderController.OrderList)                    // 查询订单列表
		orderGroup.POST("batch_orders_query", orderController.OrderList)              // 批量查询订单
		orderGroup.POST("update_express_info", orderController.OrderDeliver)          // 更新物流信息
		orderGroup.POST("backend_deliver", middleware.BackendAuthMiddleware(), orderController.OrderDeliver)
		orderGroup.POST("sync_logistics_info", orderController.SyncLogisticsInfo)           // 从聚水潭同步物流信息
		orderGroup.POST("query_jushuitan_logistic", orderController.QueryJushuitanLogistic) // 查询聚水潭发货信息

		// 额外的订单操作路由
		orderGroup.POST("update/:id", orderController.OrderUpdate) // 更新订单
		orderGroup.POST("cancel", orderController.OrderCancel)     // 取消订单
		orderGroup.POST("pay", orderController.OrderPay)           // 支付订单
		orderGroup.POST("deliver", orderController.OrderDeliver)   // 发货
		// 根据用户ID查询订单
		orderGroup.POST("query_by_user_id", orderController.QueryOrdersByUserID)
		// 申请退换货
		orderGroup.POST("request_return", orderController.OrderRequestReturn)
		orderGroup.POST("order_receive", orderController.OrderReceive)
		orderGroup.POST("backend_receive", middleware.BackendAuthMiddleware(), orderController.OrderReceive)
		// 根据商品名称搜索订单
		orderGroup.POST("search_by_product_name", orderController.SearchOrdersByProductName)
		orderGroup.POST("query_sub_order_data", orderController.SubOrderDetail)
		orderGroup.POST("change_sub_order_status", orderController.ChangeSubOrderStatus)
		orderGroup.POST("cancel_sub_order", orderController.SubOrderCancel)
		orderGroup.POST("return_sub_order", orderController.SubOrderReturn)
		orderGroup.POST("update_payment_amount", middleware.BackendAuthMiddleware(), orderController.UpdatePaymentAmount)
		orderGroup.POST("confirm_payment", middleware.BackendAuthMiddleware(), orderController.ConfirmPayment)
		orderGroup.POST("jushuitan_ship_info", orderController.JushuitanShipInfo)
	}
}
