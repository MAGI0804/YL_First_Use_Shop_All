package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"Member_shop/utils"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type MessageController struct{}

// GetMessageCategories 查询消息分类和该分类下的最后一条消息
func (mc *MessageController) GetMessageCategories(c *gin.Context) {
	var req requestbody.MessageCategoryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	categories, err := method.GetMessageCategories(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询消息分类失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"code": 200,
		"data": categories,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// GetMessagesByType 根据分类和用户ID查询消息
func (mc *MessageController) GetMessagesByType(c *gin.Context) {
	var req requestbody.MessageQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	if req.PageSize > 50 {
		req.PageSize = 50
	}

	messages, total, err := method.GetMessagesByType(req.UserID, req.MessageType, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询消息失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"code":      200,
		"data":      messages,
		"page":      req.Page,
		"page_size": req.PageSize,
		"total":     total,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// CreateMessage 自定义添加消息（支持文件上传）
func (mc *MessageController) CreateMessage(c *gin.Context) {
	var req requestbody.MessageCreateRequest

	// 绑定普通表单字段
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	// 处理文件上传
	file, err := c.FormFile("file")
	if err == nil {
		// 保存文件
		dst := utils.MediaPath(file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("文件上传失败: "+err.Error()))
			return
		}
		if err := os.Chmod(dst, 0755); err != nil {
			log.Printf("设置文件权限失败: %v", err)
		}
		// 设置文件路径到DisplayImg字段
		req.DisplayImg = "/media/" + file.Filename
	}

	err = method.CreateMessage(req.UserID, req.MessageType, req.MessageTitleOne, req.MessageTitleTwo, req.MessageBody, req.RelatedNum, req.DisplayImg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("创建消息失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"code":    200,
		"message": "消息创建成功",
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}
