package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OpenInventoryController struct{}

func (oc *OpenInventoryController) QueryAvailability(c *gin.Context) {
	var req requestbody.OpenInventoryAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("items不能为空"))
		return
	}

	items := make([]method.OpenInventoryBalanceView, 0, len(req.Items))
	for _, item := range req.Items {
		result, err := method.QueryOpenInventory(method.OpenInventoryQueryInput{
			CommodityID:   item.CommodityID,
			WarehouseCode: item.WarehouseCode,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
			return
		}
		items = append(items, result.Items...)
	}

	data := map[string]any{
		"items": items,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OpenInventoryController) QueryBalances(c *gin.Context) {
	var req requestbody.OpenInventoryBalancesRequest
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

	items := result.Items
	if req.LowAvailableThreshold > 0 {
		filtered := make([]method.OpenInventoryBalanceView, 0, len(items))
		for _, item := range items {
			if item.AvailableQty <= req.LowAvailableThreshold {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	summary := method.OpenInventorySummary{}
	for _, item := range items {
		summary.TotalOnHandQty += item.OnHandQty
		summary.TotalLockedQty += item.LockedQty
		summary.TotalAvailableQty += item.AvailableQty
	}
	total := len(items)
	page, pageSize := normalizeOpenInventoryPage(req.Page, req.PageSize)
	start := (page - 1) * pageSize
	if start >= total {
		items = []method.OpenInventoryBalanceView{}
	} else {
		end := start + pageSize
		if end > total {
			end = total
		}
		items = items[start:end]
	}

	data := map[string]any{
		"commodity_id":            result.CommodityID,
		"style_code":              result.StyleCode,
		"warehouse_code":          result.WarehouseCode,
		"low_available_threshold": req.LowAvailableThreshold,
		"summary":                 summary,
		"items":                   items,
		"total":                   total,
		"page":                    page,
		"page_size":               pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OpenInventoryController) QueryMovements(c *gin.Context) {
	var req requestbody.OpenInventoryMovementsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	movements, total, page, pageSize, err := method.QueryOpenInventoryMovements(method.OpenInventoryMovementQueryInput{
		CommodityID:   req.CommodityID,
		StyleCode:     req.StyleCode,
		WarehouseCode: req.WarehouseCode,
		MovementType:  req.MovementType,
		BizType:       req.BizType,
		BizID:         req.BizID,
		BizItemID:     req.BizItemID,
		Page:          req.Page,
		PageSize:      req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"data":      movements,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func normalizeOpenInventoryPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

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
