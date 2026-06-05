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
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderController struct{}

func (oc *OrderController) QueryOrdersByUserID(c *gin.Context) {
	var req requestbody.QueryOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求体格式错误", err))
		return
	}

	if !method.ValidateShopName(req.Shopname) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	if req.PageSize > 50 {
		req.PageSize = 50
	}

	if !method.ValidateOrderStatus(req.Status, method.ValidOrderStatuses) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("订单状态无效"))
		return
	}

	result, err := method.QueryOrdersByUserID(req.UserID, req.Status, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询订单失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"data":      result.Orders,
		"page":      req.Page,
		"page_size": req.PageSize,
		"total":     result.Total,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderList(c *gin.Context) {
	var req requestbody.OrderListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	if !method.ValidateShopName(req.Shopname) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	if req.PageSize > 50 {
		req.PageSize = 50
	}

	orders, total, err := method.GetOrderListFiltered(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询订单列表失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"code":      200,
		"data":      orders,
		"page":      req.Page,
		"page_size": req.PageSize,
		"total":     total,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderDetail(c *gin.Context) {
	var req requestbody.OrderDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求体格式错误", err))
		return
	}

	order, err := method.GetOrderDetail(req.OrderID, req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		return
	}

	detailMap := method.ConvertOrderToMap(*order)
	data := map[string]any{
		"status":  "success",
		"message": "查询订单信息成功",
		"data":    detailMap,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ChangeStatus(c *gin.Context) {
	var req requestbody.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	if !method.ValidateChangeStatus(req.Status) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的订单状态"))
		return
	}

	result, err := method.ChangeOrderStatus(req.OrderID, req.Status, req.ExpressCompany, req.ExpressNumber, req.LogisticsProcess)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		return
	}

	var productList []interface{}
	if result.Order.ProductList != "" {
		_ = json.Unmarshal([]byte(result.Order.ProductList), &productList)
	}

	orderTimeCN := result.Order.OrderTime
	formattedTime := orderTimeCN.Format("2006-01-02 15:04:05")

	data := map[string]any{
		"code":    200,
		"message": "订单状态更新成功",
		"data": map[string]interface{}{
			"order_id":          result.Order.OrderID,
			"status":            result.Order.Status,
			"old_status":        result.OldStatus,
			"receiver_name":     result.Order.ReceiverName,
			"receiver_phone":    result.Order.ReceiverPhone,
			"province":          result.Order.Province,
			"city":              result.Order.City,
			"county":            result.Order.County,
			"detailed_address":  result.Order.DetailedAddress,
			"order_amount":      result.Order.OrderAmount,
			"product_list":      productList,
			"order_time":        formattedTime,
			"express_company":   result.Order.ExpressCompany,
			"express_number":    result.Order.ExpressNumber,
			"shipped_time":      result.Order.ShippedTime,
			"delivered_time":    result.Order.DeliveredTime,
			"canceled_time":     result.Order.CanceledTime,
			"processing_time":   result.Order.ProcessingTime,
			"logistics_process": []interface{}{},
		},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderCreate(c *gin.Context) {
	var req requestbody.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	receiverPhoneStr, err := method.ParsePhoneNumber(req.ReceiverPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("手机号格式不正确"))
		return
	}

	if len(req.ProductList) == 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("product_list不能为空列表"))
		return
	}

	order, err := method.CreateOrder(req.UserID, req.ReceiverName, receiverPhoneStr, req.Province, req.City, req.County, req.DetailedAddress, req.OrderAmount, req.ProductList, req.ExpressCompany, req.ExpressNumber, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("创建订单失败"))
		return
	}

	// 将订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("created_order_id", order.OrderID)
	c.Set("created_order_user_id", req.UserID)

	// 同步到聚水潭
	token, err := jushuitan.GetToken()
	if err != nil {
		log.Printf("获取聚水潭token失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("同步到聚水潭失败"))
		return
	}

	items := make([]jushuitan.OrderItem, 0, len(req.ProductList))
	log.Printf("ProductList 长度: %d", len(req.ProductList))
	for i, item := range req.ProductList {
		log.Printf("处理商品[%d]: %+v", i, item)
		var orderItem jushuitan.OrderItem
		if productMap, ok := item.(map[string]interface{}); ok {
			orderItem = jushuitan.OrderItem{
				SkuID:     getStringValue(productMap, "commodity_id"),
				ShopSkuID: getStringValue(productMap, "commodity_id"),
				Name:      getStringValue(productMap, "product_name"),
				Qty:       getIntValue(productMap, "qty", 1),
			}
			if amount, ok := productMap["price"].(float64); ok {
				orderItem.Amount = amount
				orderItem.BasePrice = amount
			}
		} else if commodityID, ok := item.(string); ok {
			commodity, err := method.GetCommodityInfoByID(commodityID)
			if err != nil {
				log.Printf("查询商品信息失败: %s, err: %v", commodityID, err)
				continue
			}
			orderItem = jushuitan.OrderItem{
				SkuID:     commodityID,
				ShopSkuID: commodityID,
				Name:      commodity.Name,
				Qty:       1,
				Amount:    commodity.Price,
				BasePrice: commodity.Price,
			}
		}
		if orderItem.SkuID != "" {
			orderItem.OuterOiID = fmt.Sprintf("%d", i+1)
			orderItem.BatchID = "1"
			orderItem.ProducedDate = time.Now().Format("2006-01-02")
			log.Printf("构建 OrderItem: %+v", orderItem)
			items = append(items, orderItem)
		}
	}
	log.Printf("最终 items 长度: %d, items: %+v", len(items), items)

	// 解析子订单号
	var subOrderIDs []string
	if order.SubOrderIDs != "" {
		if err := json.Unmarshal([]byte(order.SubOrderIDs), &subOrderIDs); err != nil {
			log.Printf("解析子订单号失败: %v", err)
		}
	}
	log.Printf("子订单号列表: %+v", subOrderIDs)

	// 将子订单号设置到 items 的 outer_oi_id
	for i := range items {
		if i < len(subOrderIDs) {
			parts := strings.Split(subOrderIDs[i], ":")
			if len(parts) > 0 {
				items[i].OuterOiID = parts[0]
			}
		}
	}
	log.Printf("最终 items: %+v", items)

	buyerID := strconv.Itoa(req.UserID)
	paymentMethod := "其它"
	paymentTime := order.OrderTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")

	orderData := jushuitan.OrderData{
		ShopID:           10395227,
		SoID:             order.OrderID,
		OrderDate:        order.OrderTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		ShopStatus:       "WAIT_SELLER_SEND_GOODS",
		ShopBuyerID:      buyerID,
		ReceiverState:    req.Province,
		ReceiverCity:     req.City,
		ReceiverDistrict: req.County,
		ReceiverAddress:  fmt.Sprintf("%s_%s_%s_%s", req.Province, req.City, req.County, req.DetailedAddress),
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    receiverPhoneStr,
		ReceiverZip:      "200000",
		PayAmount:        req.OrderAmount,
		Freight:          0,
		Remark:           req.Remark,
		BuyerMessage:     req.Remark,
		ShopModified:     order.OrderTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		Items:            items,
		Pay: jushuitan.PayInfo{
			OuterPayID:    order.OrderID,
			PayDate:       paymentTime,
			Payment:       paymentMethod,
			SellerAccount: "youlankids",
			BuyerAccount:  buyerID,
			Amount:        req.OrderAmount,
		},
	}

	resp, err := jushuitan.SendOrder(token, orderData)
	if err != nil {
		log.Printf("上传订单到聚水潭失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("同步到聚水潭失败"))
		return
	}
	log.Printf("上传订单到聚水潭成功: %s", resp)

	responseData := map[string]interface{}{
		"order_id":         order.OrderID,
		"user_id":          req.UserID,
		"receiver_name":    req.ReceiverName,
		"receiver_phone":   receiverPhoneStr,
		"express_company":  req.ExpressCompany,
		"express_number":   req.ExpressNumber,
		"remark":           req.Remark,
		"sub_order_ids":    order.SubOrderIDs,
		"status":           order.Status,
		"pay_status":       order.PayStatus,
		"order_amount":     order.OrderAmount,
		"final_pay_amount": order.FinalPayAmount,
		"discount_amount":  order.DiscountAmount,
		"discount_reason":  order.DiscountReason,
	}
	data := map[string]any{
		"status":   "success",
		"message":  "订单创建成功",
		"order_id": order.OrderID,
		"data":     responseData,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) BackendCreateOrder(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.BackendCreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}
	order, err := method.CreateBackendOrder(req, operator, requestMeta(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	data := method.ConvertOrderToMap(*order)
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

func (oc *OrderController) OrderCancel(c *gin.Context) {
	log.Printf("OrderCancel 控制器开始执行")

	var req requestbody.OrderCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("OrderCancel 请求绑定失败, err: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	log.Printf("OrderCancel 请求绑定成功, orderID: %s, userID: %d", req.OrderID, req.UserID)

	// 获取订单详细信息
	order, err := method.GetOrderByID(req.OrderID)
	if err != nil {
		log.Printf("OrderCancel 获取订单失败: %v", err)
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		return
	}

	// 解析订单的商品列表
	var productList []interface{}
	if order.ProductList != "" {
		if err := json.Unmarshal([]byte(order.ProductList), &productList); err != nil {
			log.Printf("OrderCancel 解析商品列表失败: %v", err)
		}
	}
	log.Printf("OrderCancel 商品列表: %+v", productList)

	// 构建聚水潭订单数据
	items := make([]jushuitan.OrderItem, 0, len(productList))
	for i, item := range productList {
		log.Printf("处理商品[%d]: %+v", i, item)
		var orderItem jushuitan.OrderItem
		if productMap, ok := item.(map[string]interface{}); ok {
			orderItem = jushuitan.OrderItem{
				SkuID:     getStringValue(productMap, "commodity_id"),
				ShopSkuID: getStringValue(productMap, "commodity_id"),
				Name:      getStringValue(productMap, "product_name"),
				Qty:       getIntValue(productMap, "qty", 1),
			}
			if amount, ok := productMap["price"].(float64); ok {
				orderItem.Amount = amount
				orderItem.BasePrice = amount
			}
		} else if commodityID, ok := item.(string); ok {
			commodity, err := method.GetCommodityInfoByID(commodityID)
			if err != nil {
				log.Printf("查询商品信息失败: %s, err: %v", commodityID, err)
				continue
			}
			orderItem = jushuitan.OrderItem{
				SkuID:     commodityID,
				ShopSkuID: commodityID,
				Name:      commodity.Name,
				Qty:       1,
				Amount:    commodity.Price,
				BasePrice: commodity.Price,
			}
		}
		if orderItem.SkuID != "" {
			orderItem.OuterOiID = fmt.Sprintf("%d", i+1)
			orderItem.BatchID = "1"
			orderItem.ProducedDate = time.Now().Format("2006-01-02")
			log.Printf("构建 OrderItem: %+v", orderItem)
			items = append(items, orderItem)
		}
	}
	log.Printf("最终 items 长度: %d, items: %+v", len(items), items)

	// 解析子订单号并设置到 items 的 outer_oi_id
	var subOrderIDs []string
	if order.SubOrderIDs != "" {
		if err := json.Unmarshal([]byte(order.SubOrderIDs), &subOrderIDs); err != nil {
			log.Printf("解析子订单号失败: %v", err)
		}
	}
	log.Printf("子订单号列表: %+v", subOrderIDs)

	for i := range items {
		if i < len(subOrderIDs) {
			parts := strings.Split(subOrderIDs[i], ":")
			if len(parts) > 0 {
				items[i].OuterOiID = parts[0]
			}
		}
	}
	log.Printf("最终 items: %+v", items)

	// 获取聚水潭token
	token, err := jushuitan.GetToken()
	if err != nil {
		log.Printf("获取聚水潭token失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("同步到聚水潭失败"))
		return
	}

	// 构建订单数据，状态改为 TRADE_CLOSED
	buyerID := strconv.Itoa(order.UserID)
	paymentMethod := "其它"
	paymentTime := order.OrderTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")

	orderData := jushuitan.OrderData{
		ShopID:           10395227,
		SoID:             order.OrderID,
		OrderDate:        order.OrderTime.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		ShopStatus:       "TRADE_CLOSED",
		ShopBuyerID:      buyerID,
		ReceiverState:    order.Province,
		ReceiverCity:     order.City,
		ReceiverDistrict: order.County,
		ReceiverAddress:  fmt.Sprintf("%s_%s_%s_%s", order.Province, order.City, order.County, order.DetailedAddress),
		ReceiverName:     order.ReceiverName,
		ReceiverPhone:    order.ReceiverPhone,
		ReceiverZip:      "200000",
		PayAmount:        order.OrderAmount,
		Freight:          0,
		Remark:           order.Remarks,
		BuyerMessage:     order.Remarks,
		ShopModified:     time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"),
		Items:            items,
		Pay: jushuitan.PayInfo{
			OuterPayID:    order.OrderID,
			PayDate:       paymentTime,
			Payment:       paymentMethod,
			SellerAccount: "youlankids",
			BuyerAccount:  buyerID,
			Amount:        order.OrderAmount,
		},
	}

	// 发送到聚水潭
	resp, err := jushuitan.SendOrder(token, orderData)
	if err != nil {
		log.Printf("上传取消订单到聚水潭失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("同步到聚水潭失败"))
		return
	}
	log.Printf("上传取消订单到聚水潭成功: %s", resp)

	// 将订单ID和用户ID存储到上下文中，供中间件使用（在操作之前设置）
	c.Set("order_id", req.OrderID)
	c.Set("order_user_id", req.UserID)

	// 取消订单
	log.Printf("OrderCancel 调用 CancelOrder, orderID: %s", req.OrderID)
	err = method.CancelOrder(req.OrderID)
	if err != nil {
		log.Printf("OrderCancel CancelOrder 返回错误: %v", err)
		if err.Error() == "订单状态不允许取消" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		}
		return
	}

	log.Printf("OrderCancel 取消成功")

	data := map[string]any{
		"status":  "success",
		"message": "订单取消成功",
		"data":    map[string]string{"order_id": req.OrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) UpdatePaymentAmount(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.UpdatePaymentAmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	before, _ := method.GetOrderByID(req.OrderID)
	order, err := method.UpdatePaymentAmount(req.OrderID, req.FinalPayAmount, req.DiscountReason, int(operator.ID))
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	beforeData := map[string]any{}
	if before != nil {
		beforeData = map[string]any{
			"final_pay_amount": before.FinalPayAmount,
			"discount_amount":  before.DiscountAmount,
			"discount_reason":  before.DiscountReason,
		}
	}
	if err := method.RecordBackendOperation(method.BackendOperationLogInput{
		Operator:   operator,
		Action:     method.ActionOrderPaymentAmountUpdate,
		Module:     method.OperationModuleOrder,
		TargetType: "order",
		TargetID:   order.OrderID,
		UserID:     order.UserID,
		OrderID:    order.OrderID,
		BeforeData: beforeData,
		AfterData: map[string]any{
			"final_pay_amount": order.FinalPayAmount,
			"discount_amount":  order.DiscountAmount,
			"discount_reason":  order.DiscountReason,
		},
		RequestID: requestMeta(c).RequestID,
		ClientIP:  requestMeta(c).ClientIP,
		UserAgent: requestMeta(c).UserAgent,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
		return
	}

	data := map[string]any{
		"order_id":          order.OrderID,
		"order_amount":      order.OrderAmount,
		"final_pay_amount":  order.FinalPayAmount,
		"discount_amount":   order.DiscountAmount,
		"discount_reason":   order.DiscountReason,
		"price_adjusted_by": order.PriceAdjustedBy,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ConfirmPayment(c *gin.Context) {
	operator, ok := requireBackendOperator(c)
	if !ok {
		return
	}
	var req requestbody.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("invalid request", err))
		return
	}

	before, _ := method.GetOrderByID(req.OrderID)
	if err := method.ConfirmOrderPayment(req.OrderID, int(operator.ID), req.PaymentRemark); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		return
	}
	after, _ := method.GetOrderByID(req.OrderID)
	if after != nil {
		beforeData := map[string]any{}
		if before != nil {
			beforeData = map[string]any{
				"pay_status":       before.PayStatus,
				"payment_time":     before.PaymentTime,
				"total_paid_delta": 0,
			}
		}
		if err := method.RecordBackendOperation(method.BackendOperationLogInput{
			Operator:   operator,
			Action:     method.ActionOrderPaymentConfirm,
			Module:     method.OperationModuleOrder,
			TargetType: "order",
			TargetID:   after.OrderID,
			UserID:     after.UserID,
			OrderID:    after.OrderID,
			BeforeData: beforeData,
			AfterData: map[string]any{
				"pay_status":          after.PayStatus,
				"payment_time":        after.PaymentTime,
				"payment_operator_id": after.PaymentOperatorID,
				"payment_remark":      after.PaymentRemark,
				"total_paid_delta":    after.FinalPayAmount,
			},
			RequestID: requestMeta(c).RequestID,
			ClientIP:  requestMeta(c).ClientIP,
			UserAgent: requestMeta(c).UserAgent,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr(err.Error()))
			return
		}
	}

	data := map[string]any{
		"order_id":    req.OrderID,
		"pay_status":  "paid",
		"operator_id": operator.ID,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderPay(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("订单ID不能为空"))
		return
	}

	err := method.PayOrder(orderID)
	if err != nil {
		if err.Error() == "order already paid" || err.Error() == "order status does not allow payment" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
			return
		}
		if err.Error() == "订单状态不允许支付" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		}
		return
	}

	data := map[string]any{
		"status":  "success",
		"message": "订单支付成功",
		"data":    map[string]string{"order_id": orderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderDeliver(c *gin.Context) {
	var req requestbody.OrderDeliverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.DeliverOrder(req.OrderID, req.ExpressCompany, req.ExpressNumber)
	if err != nil {
		if err.Error() == "订单状态不允许发货" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		}
		return
	}

	// 将订单ID存储到上下文中，供中间件使用
	c.Set("order_id", req.OrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "订单发货成功",
		"data":    map[string]string{"order_id": req.OrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderReceive(c *gin.Context) {
	var req requestbody.OrderReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ReceiveOrder(req.OrderID)
	if err != nil {
		if err.Error() == "订单状态不允许签收" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		}
		return
	}

	// 将订单ID存储到上下文中，供中间件使用
	c.Set("order_id", req.OrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "订单签收成功",
		"data":    map[string]string{"order_id": req.OrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderRequestReturn(c *gin.Context) {
	var req requestbody.OrderRequestReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	if req.Type == "" {
		req.Type = "return"
	}
	order, err := method.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		return
	}

	// 检查请求中的订单状态与实际订单状态是否一致
	if order.Status != req.OrderStatus {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("订单状态不一致"))
		return
	}

	allowedStatus := map[string]bool{
		"delivered": true,
		"shipped":   true,
	}

	// 判断订单状态是否不在允许的集合中
	if !allowedStatus[order.Status] {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("订单状态不允许申请退换货"))
		return
	}
	result, err := method.RequestReturn(req.UserID, req.OrderID, req.OrderStatus, req.Type, req.Reason, req.SpecificReasons, req.BuyerProvince, req.BuyerCity, req.BuyerCounty, req.BuyerAddress, req.BuyerPhone, req.ProductIDs, order.ProductList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("申请退换货失败: "+err.Error()))
		return
	}

	// 将订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("order_id", req.OrderID)
	c.Set("order_user_id", req.UserID)
	c.Set("created_return_order_id", result.ReturnID)

	typeLabel := map[string]string{
		"return":      "退货",
		"exchange":    "换货",
		"refund":      "仅退款",
		"replacement": "补发",
		"reissue":     "补发",
	}[req.Type]
	if typeLabel == "" {
		typeLabel = "售后"
	}
	data := map[string]any{
		"status":  "success",
		"message": typeLabel + "申请提交成功",
		"data": map[string]interface{}{
			"order_id":        req.OrderID,
			"type":            req.Type,
			"return_order_id": result.ReturnID,
		},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) SyncLogisticsInfo(c *gin.Context) {
	var req requestbody.JushuitanLogisticQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	oc.queryJushuitanLogistic(c, req)
}

func (oc *OrderController) QueryJushuitanLogistic(c *gin.Context) {
	var req requestbody.JushuitanLogisticQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	oc.queryJushuitanLogistic(c, req)
}

func (oc *OrderController) queryJushuitanLogistic(c *gin.Context, req requestbody.JushuitanLogisticQueryRequest) {
	soIDs := append([]string{}, req.SoIDs...)
	if req.OrderID != "" && len(soIDs) == 0 {
		soIDs = []string{req.OrderID}
	}
	if len(soIDs) == 0 && (req.ModifiedBegin == "" || req.ModifiedEnd == "") {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("order_id/so_ids或modified_begin+modified_end不能为空"))
		return
	}

	token, err := jushuitan.GetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取聚水潭token失败: "+err.Error()))
		return
	}

	resp, rawResp, err := jushuitan.QueryLogistic(token, jushuitan.LogisticQueryRequest{
		ShopID:        req.ShopID,
		PageIndex:     req.PageIndex,
		PageSize:      req.PageSize,
		ModifiedBegin: req.ModifiedBegin,
		ModifiedEnd:   req.ModifiedEnd,
		DateType:      req.DateType,
		SoIDs:         soIDs,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, msg.ErrResponseStr(err.Error()))
		return
	}

	applyResults := make([]*method.JushuitanLogisticApplyResult, 0, len(resp.Data.Orders))
	for _, logisticOrder := range resp.Data.Orders {
		items := make([]method.JushuitanLogisticItemInput, 0, len(logisticOrder.Items))
		for _, item := range logisticOrder.Items {
			items = append(items, method.JushuitanLogisticItemInput{
				SubOrderID: item.OuterOiID,
				Qty:        item.Qty,
				SkuID:      item.SkuID,
			})
		}
		result, applyErr := method.ApplyJushuitanLogisticOrder(method.JushuitanLogisticOrderInput{
			OrderID:          logisticOrder.SoID,
			ExpressCompany:   logisticOrder.LogisticsCompany,
			ExpressNumber:    logisticOrder.LID,
			SendDate:         logisticOrder.SendDate,
			LogisticsProcess: logisticOrder,
			Items:            items,
		})
		if applyErr != nil {
			log.Printf("应用聚水潭发货信息失败, so_id=%s: %v", logisticOrder.SoID, applyErr)
			continue
		}
		applyResults = append(applyResults, result)
	}

	data := map[string]any{
		"raw_response":      rawResp,
		"page_index":        resp.Data.PageIndex,
		"page_size":         resp.Data.PageSize,
		"data_count":        resp.Data.DataCount,
		"page_count":        resp.Data.PageCount,
		"has_next":          resp.Data.HasNext,
		"orders":            resp.Data.Orders,
		"applied_orders":    applyResults,
		"applied_order_num": len(applyResults),
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ChangeReceivingData(c *gin.Context) {
	var req requestbody.ChangeReceivingDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ChangeReceivingData(req.OrderID, req.ReceiverName, req.ReceiverPhone, req.Province, req.City, req.County, req.DetailedAddress)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("修改收货信息失败"))
		}
		return
	}

	order, _ := method.GetOrderByID(req.OrderID)
	if order.Status == "shipped" || order.Status == "delivered" || order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("订单状态不允许修改收货信息"))
		return
	}

	data := map[string]any{
		"status":  "success",
		"message": "收货信息修改成功",
		"data":    map[string]string{"order_id": req.OrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ReturnOrderDeliver(c *gin.Context) {
	var req requestbody.ReturnOrderDeliverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ReturnOrderDeliver(req.ReturnOrderID, req.ExpressCompany, req.ExpressNumber)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许发货" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("退货订单发货失败"))
		}
		return
	}

	// 将退货订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "退货订单发货成功",
		"data": map[string]string{
			"return_order_id": req.ReturnOrderID,
			"express_company": req.ExpressCompany,
			"express_number":  req.ExpressNumber,
		},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ReturnOrderReceive(c *gin.Context) {
	var req requestbody.ReturnOrderReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ReturnOrderReceive(req.ReturnOrderID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许签收" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("退货订单签收失败"))
		}
		return
	}

	// 将退货订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "退货订单签收成功",
		"data":    map[string]string{"return_order_id": req.ReturnOrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ReturnOrderCancel(c *gin.Context) {
	var req requestbody.ReturnOrderCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ReturnOrderCancel(req.ReturnOrderID, req.Reason)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许取消" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("退货订单取消失败"))
		}
		return
	}

	// 将退货订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "退货订单取消成功",
		"data":    map[string]string{"return_order_id": req.ReturnOrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ReturnOrderUpdateBuyerInfo(c *gin.Context) {
	var req requestbody.ReturnOrderUpdateBuyerInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ReturnOrderUpdateBuyerInfo(req.ReturnOrderID, req.BuyerProvince, req.BuyerCity, req.BuyerCounty, req.BuyerAddress, req.BuyerPhone)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("退货订单不存在"))
		} else if err.Error() == "退货订单状态不允许修改买家信息" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("修改买家信息失败"))
		}
		return
	}

	// 将退货订单ID和用户ID存储到上下文中，供中间件使用
	c.Set("return_order_id", req.ReturnOrderID)
	c.Set("order_user_id", req.UserID)

	data := map[string]any{
		"status":  "success",
		"message": "买家信息修改成功",
		"data":    map[string]string{"return_order_id": req.ReturnOrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) BatchOrdersQuery(c *gin.Context) {
	var req requestbody.OrderListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	if !method.ValidateShopName(req.Shopname) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	if req.PageSize > 50 {
		req.PageSize = 50
	}

	result, err := method.BatchOrdersQuery(req.UserID, req.Status, req.BeginTime, req.EndTime, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	data := map[string]any{
		"code":      200,
		"data":      result.Orders,
		"page":      req.Page,
		"page_size": req.PageSize,
		"total":     result.Total,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) SearchOrdersByProductName(c *gin.Context) {
	var req requestbody.OrderSearchByProductNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	if !method.ValidateShopName(req.Shopname) {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	if req.PageSize > 50 {
		req.PageSize = 50
	}

	result, err := method.SearchOrdersByProductName(req.UserID, req.ProductName, req.Status, req.BeginTime, req.EndTime, req.Tid, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	data := map[string]any{
		"code":      200,
		"data":      result.Orders,
		"page":      req.Page,
		"page_size": req.PageSize,
		"total":     result.Total,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) SubOrderDetail(c *gin.Context) {
	var req requestbody.SubOrderDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求参数错误", err))
		return
	}

	subOrders, err := method.GetSubOrdersByOrderID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询子订单失败"))
		return
	}

	data := map[string]any{
		"status":     "success",
		"message":    "查询子订单成功",
		"sub_orders": method.ConvertSubOrdersToMap(subOrders),
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) ChangeSubOrderStatus(c *gin.Context) {
	var req requestbody.ChangeSubOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	err := method.ChangeSubOrderStatus(req.SubOrderID, req.Status)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("子订单不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("修改子订单状态失败"))
		}
		return
	}

	data := map[string]any{
		"status":  "success",
		"message": "子订单状态更新成功",
		"data": map[string]string{
			"sub_order_id": req.SubOrderID,
			"status":       req.Status,
		},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) SubOrderCancel(c *gin.Context) {
	var req requestbody.SubOrderCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	c.Set("sub_order_id", req.SubOrderID)
	c.Set("user_id", req.UserID)

	err := method.CancelSubOrder(req.SubOrderID, req.Reason)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("子订单不存在"))
		} else if err.Error() == "子订单状态不允许取消" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("取消子订单失败"))
		}
		return
	}

	data := map[string]any{
		"status":  "success",
		"message": "子订单取消成功",
		"data":    map[string]string{"sub_order_id": req.SubOrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) SubOrderReturn(c *gin.Context) {
	var req requestbody.SubOrderReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	c.Set("sub_order_id", req.SubOrderID)
	c.Set("user_id", req.UserID)

	err := method.ReturnSubOrder(req.SubOrderID, req.Reason, req.SpecificReasons, req.BuyerProvince, req.BuyerCity, req.BuyerCounty, req.BuyerAddress, req.BuyerPhone)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("子订单不存在"))
		} else if err.Error() == "子订单状态不允许退货" {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("申请退货失败"))
		}
		return
	}

	data := map[string]any{
		"status":  "success",
		"message": "子订单退货申请成功",
		"data":    map[string]string{"sub_order_id": req.SubOrderID},
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) OrderUpdate(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的订单ID"))
		return
	}

	order, err := method.GetOrderByID(strconv.Itoa(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("订单不存在"))
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的JSON格式"))
		return
	}

	if status, ok := updateData["status"].(string); ok && status != "" {
		order.Status = status
	}

	if paymentMethod, ok := updateData["payment_method"].(string); ok && paymentMethod != "" {
		order.PaymentMethod = paymentMethod
	}

	if paymentTime, ok := updateData["payment_time"].(string); ok && paymentTime != "" {
		parsedTime, timeErr := time.Parse("2006-01-02 15:04:05", paymentTime)
		if timeErr == nil {
			order.PaymentTime = parsedTime
		}
	}

	if deliveryMethod, ok := updateData["delivery_method"].(string); ok && deliveryMethod != "" {
		order.DeliveryMethod = deliveryMethod
	}

	if expressCompany, ok := updateData["express_company"].(string); ok && expressCompany != "" {
		order.ExpressCompany = expressCompany
	}

	if expressNumber, ok := updateData["express_number"].(string); ok && expressNumber != "" {
		order.ExpressNumber = expressNumber
	}

	if err := method.UpdateOrder(order, []string{"status", "payment_method", "payment_time", "delivery_method", "express_company", "express_number"}); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	data := map[string]any{
		"code":    200,
		"message": "订单更新成功",
		"data":    method.ConvertOrderToMap(*order),
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("success", &data))
}

func (oc *OrderController) JushuitanShipInfo(c *gin.Context) {
	var req requestbody.JushuitanShipInfoRequest

	log.Printf("========== JushuitanShipInfo 请求开始 ==========")
	log.Printf("请求URL: %s", c.Request.URL.String())
	log.Printf("请求方法: %s", c.Request.Method)
	log.Printf("Content-Type: %s", c.ContentType())
	log.Printf("ClientIP: %s", c.ClientIP())
	log.Printf("Query参数: %s", c.Request.URL.RawQuery)

	// 尝试从 form 表单中获取 biz 参数
	bizData := c.PostForm("biz")
	log.Printf("biz参数: %s", bizData)

	// 如果biz参数为空，尝试从请求体读取（JSON格式）
	if bizData == "" {
		// 检查Content-Type是否为application/json
		if strings.Contains(c.ContentType(), "application/json") {
			log.Printf("JushuitanShipInfo 尝试从JSON请求体读取数据")
			buf := make([]byte, 4096)
			n, _ := c.Request.Body.Read(buf)
			requestBody := string(buf[:n])
			log.Printf("JushuitanShipInfo 请求Body: %s", requestBody)

			// 如果请求体不为空，直接使用请求体作为biz参数
			if requestBody != "" {
				bizData = requestBody
				log.Printf("JushuitanShipInfo 从JSON请求体获取到数据")
			} else {
				log.Printf("JushuitanShipInfo 数据不存在")
				c.JSON(http.StatusBadRequest, gin.H{"code": "-1", "msg": "执行失败"})
				return
			}
		} else {
			// 对于非JSON请求，检查biz参数是否存在
			if _, exists := c.Request.PostForm["biz"]; !exists {
				log.Printf("JushuitanShipInfo biz参数不存在")
				log.Printf("JushuitanShipInfo 所有PostForm参数: %v", c.Request.PostForm)
				log.Printf("JushuitanShipInfo RawURL: %s", c.Request.URL.RawQuery)
				c.JSON(http.StatusBadRequest, gin.H{"code": "-1", "msg": "执行失败"})
				return
			}
			// 如果biz参数为空字符串，视为空对象{}
			bizData = "{}"
			log.Printf("JushuitanShipInfo biz参数为空字符串，视为空对象{}")
		}
	}

	log.Printf("所有PostForm参数: %v", c.Request.PostForm)
	log.Printf("所有Form参数: %v", c.Request.Form)
	log.Printf("请求头: %v", c.Request.Header)

	if err := json.Unmarshal([]byte(bizData), &req); err != nil {
		log.Printf("JushuitanShipInfo biz参数解析失败: %v", err)
		log.Printf("JushuitanShipInfo bizData内容: %s", bizData)
		c.JSON(http.StatusBadRequest, gin.H{"code": "-1", "msg": "执行失败"})
		return
	}

	log.Printf("JushuitanShipInfo 解析后的req: %+v", req)

	clientIP := c.ClientIP()
	requestURL := c.Request.URL.String()

	log.Printf("收到聚水潭发货信息: so_id=%s, o_id=%d, l_id=%s, lc_id=%s, is_send_all=%v",
		req.SoID, req.OID, req.LID, req.LCID, req.IsSendAll)

	order, err := method.GetOrderByID(req.SoID)
	if err != nil {
		log.Printf("根据so_id查询订单失败: %v", err)
	}

	if order != nil {
		if req.LID != "" {
			order.ExpressNumber = req.LID
		}
		if req.LogisticsCompany != "" {
			order.ExpressCompany = req.LogisticsCompany
		}
		if req.IsSendAll {
			order.Status = "shipped"
		}
		if err := method.UpdateOrder(order, []string{"status", "express_company", "express_number"}); err != nil {
			log.Printf("更新订单信息失败: %v", err)
		}
	}

	for _, item := range req.Items {
		if item.OuterOiID != "" {
			err := method.UpdateSubOrderShipInfo(item.OuterOiID, req.LID, req.LogisticsCompany, req.SendDate)
			if err != nil {
				log.Printf("更新子订单发货信息失败, sub_order_id=%s: %v", item.OuterOiID, err)
			}
		}
	}

	responseResult := `code=0&msg=执行成功`
	// 构建原始请求数据用于保存
	requestData := fmt.Sprintf("biz=%s&action_code=%s&sign=%s&timestamp=%s",
		c.PostForm("biz"), c.PostForm("action_code"), c.PostForm("sign"), c.PostForm("timestamp"))
	rawData := models.JushuitanPushRawData{
		RequestURL:  requestURL,
		RequestIP:   clientIP,
		RequestTime: time.Now(),
		Response:    responseResult,
		RawData:     requestData,
		Remarks:     fmt.Sprintf("订单号: %s, 子订单数: %d", req.SoID, len(req.Items)),
	}
	if err := db.DB.Create(&rawData).Error; err != nil {
		log.Printf("保存聚水潭原始数据失败: %v", err)
	}

	log.Printf("发货信息处理完成")
	c.JSON(http.StatusOK, gin.H{"code": "0", "msg": "执行成功"})
}
