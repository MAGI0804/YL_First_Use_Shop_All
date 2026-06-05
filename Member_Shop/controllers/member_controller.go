package controllers

import (
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MemberController struct{}

func (mc *MemberController) CreateMember(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	member, err := method.CreateMember(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{"member": member}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) UpdateMember(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	member, err := method.UpdateMember(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{"member": member}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) ListMembers(c *gin.Context) {
	var req requestbody.MemberListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	members, total, err := method.QueryMembers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{
		"items":     members,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) MemberDetail(c *gin.Context) {
	var req requestbody.MemberDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	detail, err := method.GetMemberDetail(req)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("member not found"))
		return
	}
	data := map[string]any{"detail": detail}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) ListTags(c *gin.Context) {
	var req requestbody.MemberTagListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	tags, total, err := method.QueryMemberTags(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{
		"items":     tags,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) CreateTag(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberTagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	tag, err := method.CreateMemberTag(req, operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{"tag": tag}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) SetMemberTags(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberTagSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	tags, err := method.SetMemberTags(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := map[string]any{"tags": tags}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (mc *MemberController) QueryCart(c *gin.Context) {
	var req requestbody.MemberCartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	data, err := method.QueryMemberCart(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	resp := map[string]any{"cart": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

func (mc *MemberController) AddCartItem(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberCartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	data, err := method.AddMemberCartItem(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	resp := map[string]any{"cart": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

func (mc *MemberController) UpdateCartItemQuantity(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberCartUpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	data, err := method.UpdateMemberCartItemQuantity(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	resp := map[string]any{"cart": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

func (mc *MemberController) DeleteCartItems(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.MemberCartDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}
	data, err := method.DeleteMemberCartItems(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	resp := map[string]any{"cart": data}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &resp))
}

func requireBackendOperator(c *gin.Context) (method.BackendOperatorSnapshot, bool) {
	backendUser, err := method.BackendOperatorFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, msg.ErrResponseStr(err.Error()))
		return method.BackendOperatorSnapshot{}, false
	}
	return method.BuildBackendOperatorSnapshot(backendUser), true
}

func requestMeta(c *gin.Context) method.OperationRequestMeta {
	return method.OperationRequestMeta{
		RequestID: c.GetHeader("X-Request-ID"),
		ClientIP:  c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
}
