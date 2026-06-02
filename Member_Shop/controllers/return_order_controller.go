package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReturnOrderController 退货订单控制器
type ReturnOrderController struct{}

// CreateReturnOrder 创建退货订单
func (roc *ReturnOrderController) CreateReturnOrder(c *gin.Context) {
	var req requestbody.ReturnOrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 将用户ID设置到上下文中，用于消息中间件
	c.Set("created_return_order_user_id", req.UserID)

	returnType := req.Type
	if returnType == "" {
		returnType = req.ReturnType
	}
	reason := req.Reason
	if reason == "" {
		reason = req.ReturnReason
	}

	result, err := method.CreateReturnOrderFromInput(method.ReturnOrderCreateInput{
		UserID:          req.UserID,
		OrderID:         req.OrderID,
		SubOrderID:      req.SubOrderID,
		Type:            returnType,
		Reason:          reason,
		SpecificReasons: req.SpecificReasons,
		ProductIDs:      req.ProductIDs,
		BuyerProvince:   req.BuyerProvince,
		BuyerCity:       req.BuyerCity,
		BuyerCounty:     req.BuyerCounty,
		BuyerAddress:    req.BuyerAddress,
		BuyerPhone:      req.BuyerPhone,
		Remark:          req.Remark,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("创建退货订单失败", err))
		c.Abort()
		return
	}

	info := map[string]any{
		"return_order_id": result.ReturnID,
		"status":          result.ReturnOrder.Status,
	}

	// 将创建的退货订单ID设置到上下文中，用于消息中间件
	c.Set("created_return_order_id", result.ReturnID)

	c.JSON(http.StatusOK, msg.SuccessResponse("创建退货订单成功", &info))
}

// ReturnOrderDeliver 退货订单发货
func (roc *ReturnOrderController) ReturnOrderDeliver(c *gin.Context) {
	var req requestbody.ReturnOrderDeliverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 将退货订单ID和用户ID设置到上下文中，用于消息中间件
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("return_order_user_id", req.UserID)

	if err := method.ReturnOrderDeliver(req.ReturnOrderID, req.ExpressCompany, req.ExpressNumber); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许发货" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("退货订单发货失败", err))
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("退货订单发货成功"))
}

// ReturnOrderReceive 退货订单签收
func (roc *ReturnOrderController) ReturnOrderReceive(c *gin.Context) {
	var req requestbody.ReturnOrderReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 将退货订单ID和用户ID设置到上下文中，用于消息中间件
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("return_order_user_id", req.UserID)

	if err := method.ReturnOrderReceive(req.ReturnOrderID); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许签收" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("退货订单签收失败", err))
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("退货订单签收成功"))
}

// ReturnOrderCancel 退货订单取消
func (roc *ReturnOrderController) ReturnOrderCancel(c *gin.Context) {
	var req requestbody.ReturnOrderCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 将退货订单ID和用户ID设置到上下文中，用于消息中间件
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("return_order_user_id", req.UserID)

	if err := method.ReturnOrderCancel(req.ReturnOrderID, req.Reason); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许取消" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("退货订单取消失败", err))
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("退货订单取消成功"))
}

// ReturnOrderUpdateBuyerInfo 退货订单修改买家信息
func (roc *ReturnOrderController) ReturnOrderUpdateBuyerInfo(c *gin.Context) {
	var req requestbody.ReturnOrderUpdateBuyerInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	if err := method.ReturnOrderUpdateBuyerInfo(req.ReturnOrderID, req.BuyerProvince, req.BuyerCity, req.BuyerCounty, req.BuyerAddress, req.BuyerPhone); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许修改买家信息" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("修改买家信息失败", err))
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("修改买家信息成功"))
}

// ReturnOrderApprove 退货订单审核
func (roc *ReturnOrderController) ReturnOrderApprove(c *gin.Context) {
	var req requestbody.ReturnOrderApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 将退货订单ID和用户ID设置到上下文中，用于消息中间件
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("return_order_user_id", req.UserID)
	c.Set("return_order_approve_status", req.ApproveStatus)
	c.Set("return_order_approve_remark", req.Remark)

	if err := method.ReturnOrderApprove(req.ReturnOrderID, req.ApproveStatus, req.Remark); err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许审核" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("退货订单审核失败", err))
		}
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("退货订单审核成功"))
}

// QueryReturnOrder 查询退货订单
func (roc *ReturnOrderController) QueryReturnOrder(c *gin.Context) {
	var req requestbody.ReturnOrderQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	// 默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 查询退货订单
	returnOrders, total, err := method.GetReturnOrders(req.ReturnOrderID, req.OrderID, req.UserID, req.Status, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("查询退货订单失败", err))
		c.Abort()
		return
	}

	info := map[string]any{
		"return_orders": method.ConvertReturnOrdersToMap(returnOrders),
		"total":         total,
		"page":          req.Page,
		"page_size":     req.PageSize,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询退货订单成功", &info))
}

// GetReturnOrderDetail 获取退货订单详情
func (roc *ReturnOrderController) GetReturnOrderDetail(c *gin.Context) {
	var req requestbody.ReturnOrderDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	returnOrder, err := method.GetReturnOrderDetail(req.ReturnOrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponse("退货订单不存在", err))
		c.Abort()
		return
	}

	info := map[string]any{
		"return_order": method.ConvertReturnOrderToMap(*returnOrder),
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("获取退货订单详情成功", &info))
}
