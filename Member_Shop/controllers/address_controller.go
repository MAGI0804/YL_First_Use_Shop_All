package controllers

import (
	"net/http"
	"strconv"

	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"

	"github.com/gin-gonic/gin"
)

type AddressController struct{}

func (ac *AddressController) AddAddress(c *gin.Context) {
	var req requestbody.AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	addressID, err := method.AddAddress(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("用户不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("新增地址失败"))
		}
		return
	}

	data := map[string]any{
		"address_id": addressID,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("新增地址成功", &data))
}

func (ac *AddressController) DeleteAddress(c *gin.Context) {
	var req requestbody.DeleteAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	err := method.DeleteAddress(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("地址不存在或不属于该用户"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("删除地址失败"))
		}
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("删除地址成功"))
}

func (ac *AddressController) UpdateAddress(c *gin.Context) {
	var req requestbody.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	err := method.UpdateAddress(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("地址不存在或不属于该用户"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新地址失败"))
		}
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("更新地址成功"))
}

func (ac *AddressController) SetDefaultAddress(c *gin.Context) {
	var req requestbody.SetDefaultAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	err := method.SetDefaultAddress(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("地址不存在或不属于该用户"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("设置默认地址失败"))
		}
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("设置默认地址成功"))
}

func (ac *AddressController) GetAddresses(c *gin.Context) {
	var req requestbody.GetAddressesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少必填字段", err))
		return
	}

	addresses, err := method.GetAddresses(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取地址失败"))
		return
	}

	data := map[string]any{
		"addresses": addresses,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("获取地址成功", &data))
}

func (ac *AddressController) GetAddressByID(c *gin.Context) {
	var requestMap map[string]interface{}
	if err := c.ShouldBindJSON(&requestMap); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求格式错误", err))
		return
	}

	addressID, ok := requestMap["address_id"]
	if !ok {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("缺少address_id字段"))
		return
	}

	var addressIDInt int
	switch v := addressID.(type) {
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("address_id格式错误，需为整数"))
			return
		}
		addressIDInt = id
	case float64:
		addressIDInt = int(v)
	case int:
		addressIDInt = v
	default:
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("address_id格式错误"))
		return
	}

	userID, ok := requestMap["user_id"]
	if !ok {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("缺少user_id字段"))
		return
	}

	var userIDInt int
	switch v := userID.(type) {
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("user_id格式错误，需为整数"))
			return
		}
		userIDInt = id
	case float64:
		userIDInt = int(v)
	case int:
		userIDInt = v
	default:
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("user_id格式错误"))
		return
	}

	req := requestbody.GetAddressByIDRequest{
		AddressID: addressIDInt,
		UserID:    userIDInt,
	}

	address, err := method.GetAddressByID(req)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("地址不存在或不属于该用户"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取地址失败"))
		}
		return
	}

	data := map[string]any{
		"address": address,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("获取地址成功", &data))
}
