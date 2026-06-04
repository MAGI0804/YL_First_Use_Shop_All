package controllers

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/jushuitan"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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

// ReturnOrderStatistics returns after-sales summary metrics for operations.
func (roc *ReturnOrderController) ReturnOrderStatistics(c *gin.Context) {
	var req requestbody.ReturnOrderStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	statistics, err := method.ReturnOrderStatistics(req.BeginTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		c.Abort()
		return
	}

	info := map[string]any{
		"statistics": statistics,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &info))
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

func (roc *ReturnOrderController) PushReturnOrderToJushuitan(c *gin.Context) {
	var req requestbody.ReturnOrderPushJushuitanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	if err := method.PushReturnOrderToJushuitan(req.ReturnOrderID); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("推送聚水潭售后失败", err))
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, msg.SuccessResponseStr("推送聚水潭售后成功"))
}

func (roc *ReturnOrderController) JushuitanAfterSalePush(c *gin.Context) {
	req, rawData, err := parseJushuitanAfterSalePushRequest(c)
	if err != nil {
		log.Printf("解析聚水潭售后推送失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "-1", "msg": "执行失败"})
		return
	}

	status := firstNonEmpty(req.Status, req.ShopStatus, req.RefundStatus)
	applyErr := method.ApplyJushuitanAfterSaleUpdate(method.JushuitanAfterSaleUpdateInput{
		ReturnID:             req.ReturnOrderID,
		JushuitanAfterSaleID: req.JushuitanAfterSaleID,
		OrderID:              req.OrderID,
		Status:               status,
		Response:             rawData,
	})

	responseResult := `code=0&msg=执行成功`
	if applyErr != nil {
		responseResult = fmt.Sprintf("code=-1&msg=%s", applyErr.Error())
	}
	saveJushuitanAfterSaleRawData(c, rawData, responseResult, req)

	if applyErr != nil {
		log.Printf("处理聚水潭售后推送失败: %v", applyErr)
		c.JSON(http.StatusOK, gin.H{"code": "-1", "msg": applyErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "0", "msg": "执行成功"})
}

func (roc *ReturnOrderController) QueryJushuitanAfterSaleReceived(c *gin.Context) {
	var req requestbody.JushuitanAfterSaleReceivedQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		c.Abort()
		return
	}

	token, err := jushuitan.GetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("获取聚水潭token失败", err))
		c.Abort()
		return
	}
	resp, err := jushuitan.QueryAfterSaleReceived(token, jushuitan.AfterSaleReceivedQuery{
		PageIndex:     req.PageIndex,
		PageSize:      req.PageSize,
		ModifiedBegin: req.ModifiedBegin,
		ModifiedEnd:   req.ModifiedEnd,
		SoID:          req.OrderID,
		OuterASID:     req.ReturnOrderID,
		ASID:          req.ASID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("查询聚水潭实际收货失败", err))
		c.Abort()
		return
	}
	appliedCount, applyErr := method.ApplyJushuitanAfterSaleReceivedResponse(resp)
	if applyErr != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("回写聚水潭实际收货失败", applyErr))
		c.Abort()
		return
	}

	info := map[string]any{
		"response":      resp,
		"applied_count": appliedCount,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("查询聚水潭实际收货成功", &info))
}

func parseJushuitanAfterSalePushRequest(c *gin.Context) (requestbody.JushuitanAfterSalePushRequest, string, error) {
	var req requestbody.JushuitanAfterSalePushRequest
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return req, "", err
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		raw := string(body)
		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal(body, &wrapper); err == nil {
			if biz, ok := wrapper["biz"]; ok {
				var bizStr string
				if err := json.Unmarshal(biz, &bizStr); err == nil {
					raw = bizStr
				} else {
					raw = string(biz)
				}
			}
		}
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			return req, raw, err
		}
		return req, raw, nil
	}

	if err := c.Request.ParseForm(); err != nil {
		return req, "", err
	}
	raw := c.PostForm("biz")
	if raw == "" {
		return req, "", fmt.Errorf("biz参数不能为空")
	}
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return req, raw, err
	}
	return req, raw, nil
}

func saveJushuitanAfterSaleRawData(c *gin.Context, rawData, response string, req requestbody.JushuitanAfterSalePushRequest) {
	record := models.JushuitanPushRawData{
		RequestURL:  c.Request.URL.String(),
		RequestIP:   c.ClientIP(),
		RequestTime: time.Now(),
		Response:    response,
		RawData:     rawData,
		Remarks:     fmt.Sprintf("售后单: %s, 聚水潭售后单: %s, 订单号: %s", req.ReturnOrderID, req.JushuitanAfterSaleID, req.OrderID),
	}
	if err := db.DB.Create(&record).Error; err != nil {
		log.Printf("保存聚水潭售后原始数据失败: %v", err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
