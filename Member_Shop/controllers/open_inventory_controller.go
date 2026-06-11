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
