package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

func InitOpenInventoryRoutes(router *gin.Engine) {
	openInventoryController := &controllers.OpenInventoryController{}

	openInventoryGroup := router.Group("/open_inventory/")
	{
		openInventoryGroup.POST("availability", openInventoryController.QueryAvailability)
		openInventoryGroup.POST("balances", openInventoryController.QueryBalances)
		openInventoryGroup.POST("movements", openInventoryController.QueryMovements)
		openInventoryGroup.POST("query", openInventoryController.QueryInventory)
		openInventoryGroup.POST("adjust", openInventoryController.AdjustInventory)
		openInventoryGroup.POST("transfer", openInventoryController.TransferInventory)
		openInventoryGroup.POST("stock_check", openInventoryController.StockCheckInventory)
	}
}
