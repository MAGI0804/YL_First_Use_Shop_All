package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitCommodityRoutes 初始化商品相关路由 - 与Django版本commodity.urls完全匹配
func InitCommodityRoutes(router *gin.Engine) {
	// 初始化商品控制器
	commodityController := &controllers.CommodityController{}

	// 添加前置URL前缀
	commodityGroup := router.Group("/commodity/")
	{
		// 商品相关路由 - 与Django版本commodity.urls完全匹配
		commodityGroup.POST("get_all_categories", commodityController.GetAllCategories)                          //获取所有类别
		commodityGroup.POST("get_all_labels", commodityController.GetAllLabels)                                  //获取所有标签
		commodityGroup.POST("search_style_codes", commodityController.SearchStyleCodes)                          //根据类别搜索商品
		commodityGroup.POST("add_goods", commodityController.AddGoods)                                           //添加商品
		commodityGroup.POST("delete_goods", commodityController.DeleteGoods)                                     //删除商品
		commodityGroup.POST("search_commodity_data", commodityController.SearchCommodityData)                    //查询商品信息
		commodityGroup.POST("goods_query", commodityController.GoodsQuery)                                       //批量查询商品
		commodityGroup.POST("goods_query_wx", commodityController.GoodsQueryWX)                                   //批量查询商品（小程序专用）
		commodityGroup.POST("change_commodity_data", commodityController.ChangeCommodityData)                    //修改商品信息
		commodityGroup.POST("change_commodity_status_online", commodityController.ChangeCommodityStatusOnline)   //商品上线
		commodityGroup.POST("change_commodity_status_offline", commodityController.ChangeCommodityStatusOffline) //商品下线
		commodityGroup.POST("get_commodity_status", commodityController.GetCommodityStatus)                      //获取商品状态
		commodityGroup.POST("search_products_by_name", commodityController.SearchProductsByName)                 //根据名称搜索商品
		commodityGroup.POST("batch_get_products_by_ids", commodityController.BatchGetProductsByIDs)              //批量获取商品信息
		commodityGroup.POST("stylecode_status_online", commodityController.ChangeStyleCodeStatusOnline)          //编码上线
		commodityGroup.POST("stylecode_status_offline", commodityController.ChangeStyleCodeStatusOffline)        //编码下线
		commodityGroup.POST("stylecode_commodities", commodityController.GetCommoditiesByStyleCode)              //根据款式代码获取商品列表
		commodityGroup.POST("update_style_code_info", commodityController.UpdateStyleCodeInfo)                   //更新款式信息
	}
}
