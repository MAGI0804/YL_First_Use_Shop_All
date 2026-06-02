package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnalyticsController 负责数据分析相关 HTTP 接口。
// 控制器只做参数绑定和响应封装，具体统计口径统一放在 service/method 中，方便后续复用和测试。
type AnalyticsController struct{}

// SalesSummary 返回销售汇总数据。
func (ac *AnalyticsController) SalesSummary(c *gin.Context) {
	filter, ok := ac.bindFilter(c)
	if !ok {
		return
	}

	data, err := method.SalesSummary(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	resp := map[string]any{"summary": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

// UserSummary 返回用户增长、下单和偏好数据。
func (ac *AnalyticsController) UserSummary(c *gin.Context) {
	filter, ok := ac.bindFilter(c)
	if !ok {
		return
	}

	data, err := method.UserSummary(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	resp := map[string]any{"summary": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

// ProductSummary 返回商品销量、库存和评价数据。
func (ac *AnalyticsController) ProductSummary(c *gin.Context) {
	filter, ok := ac.bindFilter(c)
	if !ok {
		return
	}

	data, err := method.ProductSummary(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	resp := map[string]any{"summary": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

// TrafficSummary 是流量分析预留接口。
// 当前项目还没有页面访问埋点表，因此明确返回 not_implemented，避免把请求日志误当业务流量。
func (ac *AnalyticsController) TrafficSummary(c *gin.Context) {
	filter, ok := ac.bindFilter(c)
	if !ok {
		return
	}

	data, err := method.TrafficSummary(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	resp := map[string]any{"summary": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

// Export 返回结构化聚合数据。
// 第一版先返回 JSON 结构，后续如果需要 Excel/CSV，可在这一层增加文件输出。
func (ac *AnalyticsController) Export(c *gin.Context) {
	filter, ok := ac.bindFilter(c)
	if !ok {
		return
	}

	data, err := method.ExportAnalytics(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	resp := map[string]any{"export": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

// bindFilter 统一绑定数据分析筛选条件，保证五个接口使用同一套入参。
func (ac *AnalyticsController) bindFilter(c *gin.Context) (method.AnalyticsFilter, bool) {
	var req requestbody.AnalyticsFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return method.AnalyticsFilter{}, false
	}

	filter := method.AnalyticsFilter{
		BeginTime:             req.BeginTime,
		EndTime:               req.EndTime,
		Shopname:              req.Shopname,
		Category:              req.Category,
		StyleCode:             req.StyleCode,
		OperatorID:            req.OperatorID,
		LowInventoryThreshold: req.LowInventoryThreshold,
		SlowSalesThreshold:    req.SlowSalesThreshold,
		Limit:                 req.Limit,
	}
	return filter, true
}
