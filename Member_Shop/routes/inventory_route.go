package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitInventoryRoutes 初始化库存相关路由
func InitInventoryRoutes(router *gin.Engine) {
	// 创建库存控制器实例
	inventoryController := &controllers.InventoryController{}

	// 为所有库存相关路由添加"/inventory/"前缀
	inventoryGroup := router.Group("/inventory/")
	{
		// 查询库存 - 查询商品库存信息
		inventoryGroup.POST("query", inventoryController.QueryInventory)
		// 调整库存 - 手动调整库存数量
		inventoryGroup.POST("adjust", inventoryController.AdjustInventory)
		// 查询库存日志 - 查询库存变更历史记录
		inventoryGroup.POST("logs", inventoryController.QueryInventoryLogs)
		// 查询库存预警 - 查询库存低于预警阈值的商品
		inventoryGroup.POST("warnings", inventoryController.QueryInventoryWarnings)
		// 库存调拨 - 在源仓和目标仓之间记录调拨出入库
		inventoryGroup.POST("transfer", inventoryController.TransferInventory)
		// 库存盘点 - 按实盘数量修正库存并记录差异
		inventoryGroup.POST("stock_check", inventoryController.StockCheckInventory)
		// 同步聚水潭库存 - 从聚水潭同步库存数据
		inventoryGroup.POST("sync_jushuitan", inventoryController.SyncJushuitanInventory)
		// 查询聚水潭库存 - 只返回ERP库存，不应用本地库存
		inventoryGroup.POST("query_jushuitan", inventoryController.QueryJushuitanInventory)
		// 接收聚水潭库存同步消息 business_sku_syn
		inventoryGroup.POST("jushuitan_sku_sync", inventoryController.JushuitanSkuSync)
	}
}
