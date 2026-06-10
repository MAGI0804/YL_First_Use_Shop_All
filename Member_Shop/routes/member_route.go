package routes

import (
	"Member_shop/controllers"
	"Member_shop/middleware"

	"github.com/gin-gonic/gin"
)

func InitMemberRoutes(router *gin.Engine) {
	memberController := &controllers.MemberController{}

	memberGroup := router.Group("/member/")
	memberGroup.Use(middleware.BackendAuthMiddleware())
	{
		memberGroup.POST("create", memberController.CreateMember)
		memberGroup.POST("update", memberController.UpdateMember)
		memberGroup.POST("list", memberController.ListMembers)
		memberGroup.POST("detail", memberController.MemberDetail)
		memberGroup.GET("import/template", memberController.DownloadImportTemplate)
		memberGroup.POST("import/match", memberController.MatchImportFile)
		memberGroup.POST("import/confirm", memberController.ConfirmImport)
		memberGroup.POST("tag/list", memberController.ListTags)
		memberGroup.POST("tag/create", memberController.CreateTag)
		memberGroup.POST("tag/set_member_tags", memberController.SetMemberTags)
		memberGroup.POST("tag/add_member_tags", memberController.SetMemberTags)
		memberGroup.POST("cart/query", memberController.QueryCart)
		memberGroup.POST("cart/add", memberController.AddCartItem)
		memberGroup.POST("cart/update_quantity", memberController.UpdateCartItemQuantity)
		memberGroup.POST("cart/delete", memberController.DeleteCartItems)
	}
}
