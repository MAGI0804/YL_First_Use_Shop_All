package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitActivityRoutes 初始化活动相关路由 - 与Django版本activity.urls完全匹配
func InitActivityRoutes(router *gin.Engine) {
	// 初始化活动控制器
	activityController := &controllers.ActivityController{}

	// 添加前置URL前缀
	activityGroup := router.Group("/activity/")
	{
		// 活动相关路由 - 与Django版本完全匹配
		activityGroup.POST("add_activity_img", activityController.AddActivityImg)  //新增活动图
		activityGroup.POST("update_activity_image_relations", activityController.UpdateActivityImageRelations) //更新活动图关系
		activityGroup.POST("activity_image_online", activityController.ActivityImageOnline) //上线活动图
		activityGroup.POST("activity_image_offline", activityController.ActivityImageOffline)  //下线活动图
		activityGroup.POST("batch_query_activity_images", activityController.BatchQueryActivityImages)  //获取所有活动图片
		activityGroup.POST("query_online_activity_images", activityController.QueryOnlineActivityImages)  //获取所有已上线的活动图片
		activityGroup.POST("batch_update_activity_image_order", activityController.BatchUpdateActivityImageOrder)  //批量更新活动图片顺序
		activityGroup.POST("add_promotional_pic", activityController.AddPromotionalPic)  //新增宣传图
		activityGroup.POST("update_promotional_pic_order", activityController.UpdatePromotionalPicOrder)  //调整宣传图位置
		activityGroup.POST("delete_promotional_pic", activityController.DeletePromotionalPic)  //删除宣传图
		activityGroup.POST("get_activity_image_detail", activityController.GetActivityImageDetail)  //根据活动图id查询详情
		activityGroup.POST("set_has_activity_detail", activityController.SetHasActivityDetail)  //设置活动详情状态
	}
}
