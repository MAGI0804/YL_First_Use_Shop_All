package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InventoryController 库存管理控制器
// 负责处理库存相关的HTTP请求，包括查询、调整、日志查询等功能
type InventoryController struct{}

// QueryInventory 处理库存查询请求
// 支持按商品ID精确查询或按款式编码查询该款式下所有商品的库存汇总
func (ic *InventoryController) QueryInventory(c *gin.Context) {
	var req requestbody.InventoryQueryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	data, err := method.QueryInventory(req.CommodityID, req.StyleCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// AdjustInventory 处理库存调整请求
// 用于手动调整商品库存，可增加或减少库存数量
func (ic *InventoryController) AdjustInventory(c *gin.Context) {
	var req requestbody.InventoryAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	if err := method.AdjustInventory(method.ChangeInventoryInput{
		CommodityID:   req.CommodityID,
		ChangeQty:     req.ChangeQty,
		OperatorID:    req.OperatorID,
		Remark:        req.Remark,
		WarehouseCode: req.WarehouseCode,
	}); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	c.JSON(http.StatusOK, msg.SuccessResponseStr("success"))
}

// QueryInventoryLogs 处理库存变动日志查询请求
// 查询库存变动的历史记录，支持按商品ID、款式编码、变动类型等多种条件筛选
func (ic *InventoryController) QueryInventoryLogs(c *gin.Context) {
	var req requestbody.InventoryLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	logs, total, page, pageSize, err := method.QueryInventoryLogs(method.InventoryLogQueryInput{
		CommodityID:       req.CommodityID,
		StyleCode:         req.StyleCode,
		ChangeType:        req.ChangeType,
		RelatedOrderID:    req.RelatedOrderID,
		RelatedSubOrderID: req.RelatedSubOrderID,
		RelatedReturnID:   req.RelatedReturnID,
		Page:              req.Page,
		PageSize:          req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// QueryInventoryWarnings 处理库存预警查询请求
// 查询库存低于设定阈值的商品列表，用于及时补充库存
func (ic *InventoryController) QueryInventoryWarnings(c *gin.Context) {
	var req requestbody.InventoryWarningsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	commodities, total, threshold, page, pageSize, err := method.QueryInventoryWarnings(req.Threshold, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      commodities,
		"total":     total,
		"threshold": threshold,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (ic *InventoryController) TransferInventory(c *gin.Context) {
	var req requestbody.InventoryTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	if err := method.TransferInventory(method.InventoryTransferInput{
		CommodityID:         req.CommodityID,
		Qty:                 req.Qty,
		SourceWarehouseCode: req.SourceWarehouseCode,
		TargetWarehouseCode: req.TargetWarehouseCode,
		OperatorID:          req.OperatorID,
		Remark:              req.Remark,
	}); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"commodity_id":           req.CommodityID,
		"qty":                    req.Qty,
		"source_warehouse_code":  req.SourceWarehouseCode,
		"target_warehouse_code":  req.TargetWarehouseCode,
		"inventory_change_type":  method.InventoryChangeStockTransfer,
		"inventory_total_change": 0,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (ic *InventoryController) StockCheckInventory(c *gin.Context) {
	var req requestbody.InventoryStockCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	data, err := method.StockCheckInventory(method.InventoryStockCheckInput{
		CommodityID:   req.CommodityID,
		ActualQty:     req.ActualQty,
		WarehouseCode: req.WarehouseCode,
		OperatorID:    req.OperatorID,
		Remark:        req.Remark,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

// SyncJushuitanInventory 处理库存同步聚水潭请求
// 将指定商品的库存数据同步到聚水潭系统（当前为预留接口）
func (ic *InventoryController) SyncJushuitanInventory(c *gin.Context) {
	var req requestbody.InventorySyncJushuitanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	data := map[string]any{
		"commodity_ids": req.CommodityIDs,
		"status":        "reserved",
	}
	c.JSON(http.StatusNotImplemented, msg.SuccessResponse("sync_jushuitan route is reserved; token-based sync will be implemented in the jushuitan phase", &data))
}
