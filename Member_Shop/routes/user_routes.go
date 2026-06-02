package routes

import (
	"Member_shop/controllers"

	"github.com/gin-gonic/gin"
)

// InitUserRoutes 初始化用户相关路由 - 与Django版本users.urls完全匹配
func InitUserRoutes(router *gin.Engine) {
	// 初始化用户控制器
	userController := &controllers.UserController{}

	// 添加前置URL前缀
	userGroup := router.Group("/ordinary_user/")
	{
		// 用户相关路由 - 与Django版本users.urls完全匹配
		userGroup.POST("add_data", userController.AddData)
		userGroup.POST("find_data", userController.QueryUserInfo)
		userGroup.POST("Modify_data", userController.UserModify)
		userGroup.POST("get_user_id", userController.UserGetID)
		userGroup.POST("wechat_login", userController.WechatLogin)
		userGroup.POST("send_register_captcha", userController.SendRegisterCaptcha)
		userGroup.POST("bind_wechat_phone", userController.BindWechatPhone)
		userGroup.POST("update_platform_info", userController.UpdatePlatformInfo)
		userGroup.POST("member_amount_summary", userController.MemberAmountSummary)

	}
}
