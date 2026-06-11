package controllers

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"Member_shop/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ReviewController 评价管理控制器
// 负责处理评价相关的HTTP请求，包括创建、查询、审核、回复、统计等功能
type ReviewController struct{}

// CreateReview 处理创建评价请求
// 用户针对已收货的商品提交评价，包含评分、内容、图片、标签等信息
func (rc *ReviewController) CreateReview(c *gin.Context) {
	var req requestbody.ReviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	review, err := method.CreateReview(method.ReviewCreateInput{
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		SubOrderID:  req.SubOrderID,
		CommodityID: req.CommodityID,
		StyleCode:   req.StyleCode,
		Rating:      req.Rating,
		Content:     req.Content,
		Images:      req.Images,
		Tags:        req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"review": review}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// QueryByProduct 处理商品评价查询请求（前台）
// 根据商品ID或款式编码查询已通过审核的评价列表
func (rc *ReviewController) QueryByProduct(c *gin.Context) {
	var req requestbody.ReviewProductQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	reviews, total, page, pageSize, err := method.QueryReviewsByProduct(method.ReviewProductQueryInput{
		CommodityID: req.CommodityID,
		StyleCode:   req.StyleCode,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// QueryBackend 处理后台评价查询请求
// 支持多种筛选条件查询所有状态的评价，用于后台审核管理
func (rc *ReviewController) QueryBackend(c *gin.Context) {
	var req requestbody.ReviewBackendQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	reviews, total, page, pageSize, err := method.QueryReviewsForBackend(method.ReviewBackendQueryInput{
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		SubOrderID:  req.SubOrderID,
		CommodityID: req.CommodityID,
		StyleCode:   req.StyleCode,
		Status:      req.Status,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// QueryMine 处理用户本人评价查询请求
func (rc *ReviewController) QueryMine(c *gin.Context) {
	var req requestbody.ReviewMineQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	reviews, total, page, pageSize, err := method.QueryMyReviews(method.ReviewMineQueryInput{
		UserID:   req.UserID,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      reviews,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// UpdateReview 处理用户修改待审核评价请求
func (rc *ReviewController) UpdateReview(c *gin.Context) {
	var req requestbody.ReviewUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	review, err := method.UpdatePendingReview(method.ReviewUpdateInput{
		ReviewID: req.ReviewID,
		UserID:   req.UserID,
		Rating:   req.Rating,
		Content:  req.Content,
		Images:   req.Images,
		Tags:     req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"review": review}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// DeleteReview 处理用户软删除待审核评价请求
func (rc *ReviewController) DeleteReview(c *gin.Context) {
	var req requestbody.ReviewDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	review, err := method.DeletePendingReview(req.ReviewID, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"review": review}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// UploadReviewImage 处理评价图片上传请求
func (rc *ReviewController) UploadReviewImage(c *gin.Context) {
	userID, err := strconv.Atoi(strings.TrimSpace(c.PostForm("user_id")))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("user_id is required"))
		return
	}
	var user models.User
	if err := db.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("user not found"))
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("image is required"))
		return
	}
	if err := method.ValidateReviewImageUpload(file); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	savePath, _, err := utils.SaveFileWithPerms(c, file, "reviews", "review_")
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	imagePath := "/media/" + filepath.ToSlash(savePath)
	data := map[string]any{
		"url":  imagePath,
		"path": savePath,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// AuditReview 处理评价审核请求
// 后台管理员审核用户提交的评价，可通过或拒绝评价
func (rc *ReviewController) AuditReview(c *gin.Context) {
	var req requestbody.ReviewAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	review, err := method.AuditReview(req.ReviewID, req.Status, req.AuditRemark)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"review": review}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// ReplyReview 处理评价回复请求
// 运营或客服对用户评价进行回复
func (rc *ReviewController) ReplyReview(c *gin.Context) {
	var req requestbody.ReviewReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	reply, err := method.ReplyReview(req.ReviewID, req.OperatorID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"reply": reply}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// ReviewStatistics 处理评价统计请求
// 获取商品或款式的评价统计数据，包括总数、平均评分、好评率、评分分布等
func (rc *ReviewController) ReviewStatistics(c *gin.Context) {
	var req requestbody.ReviewStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	stats, err := method.GetReviewStatistics(req.CommodityID, req.StyleCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"statistics": stats}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}
