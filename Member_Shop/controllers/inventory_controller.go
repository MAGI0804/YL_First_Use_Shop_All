package controllers

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/jushuitan"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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

// SyncJushuitanInventory 从聚水潭查询库存并应用到本地商品库存。
func (ic *InventoryController) SyncJushuitanInventory(c *gin.Context) {
	var req requestbody.InventorySyncJushuitanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	req.Apply = true
	ic.queryJushuitanInventory(c, req)
}

func (ic *InventoryController) QueryJushuitanInventory(c *gin.Context) {
	var req requestbody.InventorySyncJushuitanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	req.Apply = false
	ic.queryJushuitanInventory(c, req)
}

func (ic *InventoryController) queryJushuitanInventory(c *gin.Context, req requestbody.InventorySyncJushuitanRequest) {
	skuIDs := normalizedInventorySkuIDs(req)
	if len(skuIDs) == 0 && (strings.TrimSpace(req.ModifiedBegin) == "" || strings.TrimSpace(req.ModifiedEnd) == "") {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("sku_ids/commodity_ids或modified_begin+modified_end不能为空"))
		return
	}

	token, err := jushuitan.GetTokenProd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取聚水潭生产token失败: "+err.Error()))
		return
	}

	pageIndex := req.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	resp, err := jushuitan.QueryInventoryWithRequest(token, jushuitan.InventoryQueryRequest{
		PageIndex:     pageIndex,
		PageSize:      pageSize,
		SkuIDs:        strings.Join(skuIDs, ","),
		ModifiedBegin: req.ModifiedBegin,
		ModifiedEnd:   req.ModifiedEnd,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, msg.ErrResponseStr(err.Error()))
		return
	}

	applyResults := []method.JushuitanInventorySyncResult{}
	if req.Apply {
		for _, item := range resp.Data.Inventorys {
			result, applyErr := method.ApplyJushuitanInventorySync(method.JushuitanInventorySyncInput{
				SkuID:      item.SkuID,
				IID:        item.IID,
				Name:       item.Name,
				Qty:        item.Qty,
				VirtualQty: item.VirtualQty,
				OrderLock:  item.OrderLock,
				PickLock:   item.PickLock,
				Modified:   item.Modified,
			})
			if applyErr != nil {
				log.Printf("应用聚水潭库存查询结果失败, sku_id=%s: %v", item.SkuID, applyErr)
				continue
			}
			applyResults = append(applyResults, *result)
		}
	}

	data := map[string]any{
		"jushuitan_inventory": resp,
		"applied":             req.Apply,
		"apply_results":       applyResults,
		"applied_count":       len(applyResults),
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (ic *InventoryController) JushuitanSkuSync(c *gin.Context) {
	req, rawData, err := parseJushuitanInventorySkuSyncRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "-1", "msg": "执行失败"})
		return
	}

	items := req.Items
	if len(items) == 0 {
		var single requestbody.JushuitanInventorySkuItem
		if err := json.Unmarshal([]byte(rawData), &single); err == nil && firstInventorySkuID(single) != "" {
			items = []requestbody.JushuitanInventorySkuItem{single}
		}
	}

	results := make([]method.JushuitanInventorySyncResult, 0, len(items))
	for _, item := range items {
		result, applyErr := method.ApplyJushuitanInventorySync(method.JushuitanInventorySyncInput{
			SkuID:         firstInventorySkuID(item),
			IID:           firstNonEmptyInventoryString(item.IID, item.IIDSnake),
			Name:          item.Name,
			Qty:           item.Qty,
			VirtualQty:    firstNonZeroInventoryInt(item.VirtualQty, item.VirtualQtyRaw),
			OrderLock:     firstNonZeroInventoryInt(item.OrderLock, item.OrderLockRaw),
			PickLock:      firstNonZeroInventoryInt(item.PickLock, item.PickLockRaw),
			Modified:      item.Modified,
			WarehouseCode: firstNonEmptyInventoryString(fmt.Sprint(item.WmsCoID), fmt.Sprint(item.CoID)),
		})
		if applyErr != nil {
			log.Printf("应用聚水潭库存推送失败, sku_id=%s: %v", firstInventorySkuID(item), applyErr)
			continue
		}
		results = append(results, *result)
	}

	responseResult := `code=0&msg=执行成功`
	rawRecord := models.JushuitanPushRawData{
		RequestURL:  c.Request.URL.String(),
		RequestIP:   c.ClientIP(),
		RequestTime: time.Now(),
		Response:    responseResult,
		RawData:     rawData,
		Remarks:     fmt.Sprintf("库存同步: msg_type=%s, item_count=%d, applied_count=%d", req.MsgType, len(items), len(results)),
	}
	if err := db.DB.Create(&rawRecord).Error; err != nil {
		log.Printf("保存聚水潭库存推送原始数据失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"code": "0", "msg": "执行成功", "data": gin.H{"applied_count": len(results)}})
}

func parseJushuitanInventorySkuSyncRequest(c *gin.Context) (requestbody.JushuitanInventorySkuSyncRequest, string, error) {
	var req requestbody.JushuitanInventorySkuSyncRequest
	bizData := c.PostForm("biz")
	if strings.TrimSpace(bizData) == "" {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return req, "", err
		}
		bizData = string(body)
	}
	if strings.TrimSpace(bizData) == "" {
		return req, "", fmt.Errorf("biz不能为空")
	}
	if err := json.Unmarshal([]byte(bizData), &req); err != nil {
		return req, bizData, err
	}
	return req, bizData, nil
}

func normalizedInventorySkuIDs(req requestbody.InventorySyncJushuitanRequest) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range append(req.SkuIDs, req.CommodityIDs...) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func firstInventorySkuID(item requestbody.JushuitanInventorySkuItem) string {
	return firstNonEmptyInventoryString(item.SkuID, item.SkuIDSnake)
}

func firstNonEmptyInventoryString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && strings.TrimSpace(value) != "0" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonZeroInventoryInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
