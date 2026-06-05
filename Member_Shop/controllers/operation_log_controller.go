package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct{}

func (oc *OperationLogController) Query(c *gin.Context) {
	var req requestbody.OperationLogQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	logs, total, err := method.QueryBackendOperationLogs(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{
		"items":     logs,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}
