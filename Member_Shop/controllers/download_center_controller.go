package controllers

import (
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DownloadCenterController struct{}

func (dcc *DownloadCenterController) CreateTask(c *gin.Context) {
	backendUser := downloadCenterBackendUser(c)
	if backendUser == nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr("backend user missing"))
		return
	}

	var req requestbody.CreateDownloadTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	task, err := method.CreateDownloadTask(method.CreateDownloadTaskInput{
		TemplateCode: req.TemplateCode,
		Filters:      req.Filters,
		FileFormat:   req.FileFormat,
		RequestedBy:  int(backendUser.ID),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"task": task}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (dcc *DownloadCenterController) ListTasks(c *gin.Context) {
	backendUser := downloadCenterBackendUser(c)
	if backendUser == nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr("backend user missing"))
		return
	}

	var req requestbody.QueryDownloadTasksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	tasks, total, page, pageSize, err := method.QueryDownloadTasks(method.QueryDownloadTasksInput{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Status:       req.Status,
		BusinessType: req.BusinessType,
		TemplateCode: req.TemplateCode,
		RequestedBy:  int(backendUser.ID),
		IsAdmin:      method.IsBackendAdmin(backendUser),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"list":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (dcc *DownloadCenterController) TaskDetail(c *gin.Context) {
	backendUser := downloadCenterBackendUser(c)
	if backendUser == nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr("backend user missing"))
		return
	}

	task, err := method.GetDownloadTask(c.Param("task_id"), int(backendUser.ID), method.IsBackendAdmin(backendUser))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"task": task}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (dcc *DownloadCenterController) DownloadFile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, msg.ErrResponseStr("download file not implemented"))
}

func (dcc *DownloadCenterController) RetryTask(c *gin.Context) {
	backendUser := downloadCenterBackendUser(c)
	if backendUser == nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr("backend user missing"))
		return
	}

	task, err := method.RetryDownloadTask(c.Param("task_id"), int(backendUser.ID), method.IsBackendAdmin(backendUser))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"task": task}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (dcc *DownloadCenterController) ListTemplates(c *gin.Context) {
	templates, err := method.ListEnabledDownloadTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{"list": templates}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func downloadCenterBackendUser(c *gin.Context) *models.BackendUser {
	userValue, ok := c.Get("backendUser")
	if !ok {
		return nil
	}
	user, ok := userValue.(*models.BackendUser)
	if !ok {
		return nil
	}
	return user
}
