package controllers

import (
	"net/http"

	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"

	"github.com/gin-gonic/gin"
)

type CartController struct{}

func (cc *CartController) AddToCart(c *gin.Context) {
	var req requestbody.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	data, err := method.AddToCart(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("用户不存在"))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("商品不存在"))
		}
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("商品添加到购物车成功", data))
}

func (cc *CartController) BatchDeleteFromCart(c *gin.Context) {
	var req requestbody.BatchDeleteFromCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	deletedCount, notExistCodes, err := method.BatchDeleteFromCart(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("删除购物车商品失败"))
		return
	}

	data := map[string]any{
		"deleted_count": deletedCount,
	}
	if len(notExistCodes) > 0 {
		data["not_exist_codes"] = notExistCodes
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("删除成功", &data))
}

func (cc *CartController) QueryCartItems(c *gin.Context) {
	var req requestbody.QueryCartItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	cartItems, totalQuantity, err := method.QueryCartItems(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询购物车失败"))
		return
	}

	data := map[string]any{
		"cart_items":     cartItems,
		"items_count":    len(cartItems),
		"total_quantity": totalQuantity,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

func (cc *CartController) UpdateCartItemQuantity(c *gin.Context) {
	var req requestbody.UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	commodityCode, quantity, err := method.UpdateCartItemQuantity(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车不存在"))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车商品不存在"))
		}
		return
	}

	data := map[string]any{
		"commodity_code": commodityCode,
		"quantity":       quantity,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("更新成功", &data))
}

func (cc *CartController) IncreaseCartItemQuantity(c *gin.Context) {
	var req requestbody.IncreaseCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	commodityCode, quantity, err := method.IncreaseCartItemQuantity(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车不存在"))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车商品不存在"))
		}
		return
	}

	data := map[string]any{
		"commodity_code": commodityCode,
		"quantity":       quantity,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("增加成功", &data))
}

func (cc *CartController) DecreaseCartItemQuantity(c *gin.Context) {
	var req requestbody.DecreaseCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	commodityCode, quantity, err := method.DecreaseCartItemQuantity(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车不存在"))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("购物车商品不存在"))
		}
		return
	}

	data := map[string]any{
		"commodity_code": commodityCode,
		"quantity":       quantity,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("减少成功", &data))
}

func (cc *CartController) ClearCart(c *gin.Context) {
	var req requestbody.ClearCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	clearedCount, err := method.ClearCart(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("清空购物车失败"))
		return
	}

	data := map[string]any{
		"cleared_count": clearedCount,
		"total_items":   0,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("购物车已清空", &data))
}
