package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OpenInventoryController struct{}

func (oc *OpenInventoryController) QueryInventory(c *gin.Context) {
	var req requestbody.OpenInventoryQueryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	result, err := method.QueryOpenInventory(method.OpenInventoryQueryInput{
		CommodityID:   req.CommodityID,
		StyleCode:     req.StyleCode,
		WarehouseCode: req.WarehouseCode,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"commodity_id":   result.CommodityID,
		"style_code":     result.StyleCode,
		"warehouse_code": result.WarehouseCode,
		"summary":        result.Summary,
		"items":          result.Items,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OpenInventoryController) AdjustInventory(c *gin.Context) {
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

	data := map[string]any{
		"commodity_id":          req.CommodityID,
		"change_qty":            req.ChangeQty,
		"warehouse_code":        req.WarehouseCode,
		"inventory_change_type": method.InventoryChangeManualAdjust,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OpenInventoryController) TransferInventory(c *gin.Context) {
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

func (oc *OpenInventoryController) StockCheckInventory(c *gin.Context) {
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
