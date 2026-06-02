package middleware

import (
	"Member_shop/service/method"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func OrderMessageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("Message Middleware - Request Path: %s, Method: %s\n", c.Request.URL.Path, c.Request.Method)

		c.Next()

		if isOrderRelatedRequest(c) {
			fmt.Printf("Processing order related message\n")
			handleOrderMessage(c)
		}

		if isReturnOrderRelatedRequest(c) {
			fmt.Printf("Processing return order related message\n")
			handleReturnOrderMessage(c)
		}

		fmt.Printf("Message middleware processing completed\n")
	}
}

func isOrderRelatedRequest(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	orderPaths := []string{
		"/order/add_order",
		"/order/change_status",
		"/order/cancel",
		"/order/deliver",
		"/order/order_receive",
		"/order/request_return",
	}

	for _, orderPath := range orderPaths {
		if strings.HasSuffix(path, orderPath) && method == http.MethodPost {
			return true
		}
	}

	return false
}

func isReturnOrderRelatedRequest(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	fmt.Printf("Checking return order related request - Path: %s, Method: %s\n", path, method)

	returnOrderPaths := []string{
		"/return_order/create",
		"/return_order/deliver",
		"/return_order/receive",
		"/return_order/cancel",
		"/return_order/approve",
	}

	for _, returnOrderPath := range returnOrderPaths {
		if strings.HasSuffix(path, returnOrderPath) && method == http.MethodPost {
			fmt.Printf("Identified return order related request: %s\n", returnOrderPath)
			return true
		}
	}

	fmt.Printf("Not identified as return order related request\n")
	return false
}

func handleOrderMessage(c *gin.Context) {
	path := c.Request.URL.Path

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	switch {
	case strings.HasSuffix(path, "/order/add_order"):
		handleOrderCreated(c)
	case strings.HasSuffix(path, "/order/change_status"):
		handleOrderStatusChange(c)
	case strings.HasSuffix(path, "/order/cancel"):
		handleOrderCanceled(c)
	case strings.HasSuffix(path, "/order/deliver"):
		handleOrderShipped(c)
	case strings.HasSuffix(path, "/order/order_receive"):
		handleOrderDelivered(c)
	case strings.HasSuffix(path, "/order/request_return"):
		handleOrderReturnRequested(c)
	}
}

func handleReturnOrderMessage(c *gin.Context) {
	path := c.Request.URL.Path

	userID := getUserIDFromContext(c)
	fmt.Printf("Processing return order message - Path: %s, UserID: %d\n", path, userID)

	if userID <= 0 {
		userID = 1
		fmt.Printf("Invalid userID, using default: %d\n", userID)
	}

	switch {
	case strings.HasSuffix(path, "/return_order/create"):
		fmt.Printf("Calling handleReturnOrderCreated for return order creation\n")
		handleReturnOrderCreated(c, userID)
	case strings.HasSuffix(path, "/return_order/deliver"):
		fmt.Printf("Calling handleReturnOrderShipped for return order delivery\n")
		handleReturnOrderShipped(c, userID)
	case strings.HasSuffix(path, "/return_order/receive"):
		fmt.Printf("Calling handleReturnOrderCompleted for return order received\n")
		handleReturnOrderCompleted(c, userID)
	case strings.HasSuffix(path, "/return_order/cancel"):
		fmt.Printf("Calling handleReturnOrderCanceled for return order cancellation\n")
		handleReturnOrderCanceled(c, userID)
	case strings.HasSuffix(path, "/return_order/approve"):
		fmt.Printf("Calling handleReturnOrderApproved for return order approval result\n")
		handleReturnOrderApprovedOrRejected(c, userID)
	default:
		fmt.Printf("No matching return order handler function\n")
	}
}

func getUserIDFromContext(c *gin.Context) int {
	if userID, exists := c.Get("created_order_user_id"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from created_order_user_id: %d\n", id)
			return id
		}
	}
	if userID, exists := c.Get("created_return_order_user_id"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from created_return_order_user_id: %d\n", id)
			return id
		}
	}
	if userID, exists := c.Get("order_user_id"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from order_user_id: %d\n", id)
			return id
		}
	}
	if userID, exists := c.Get("return_order_user_id"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from return_order_user_id: %d\n", id)
			return id
		}
	}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from user_id: %d\n", id)
			return id
		}
	}
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(int); ok {
			fmt.Printf("Got userID from userID: %d\n", id)
			return id
		}
		if idStr, ok := userID.(string); ok {
			if id, err := strconv.Atoi(idStr); err == nil {
				fmt.Printf("Got userID from userID string conversion: %d\n", id)
				return id
			}
		}
	}
	fmt.Printf("Failed to get userID, checking all keys in context:\n")
	for key, value := range c.Keys {
		fmt.Printf("Context key: %s, value: %v, type: %T\n", key, value, value)
	}

	fmt.Printf("UserID not found, returning 0\n")
	return 0
}

func handleOrderCreated(c *gin.Context) {
	orderID := getOrderIDFromResponse(c)
	if orderID == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	messageBody := fmt.Sprintf("Your order has been created successfully, creation time: %s", time.Now().Format("2006-01-02 15:04:05"))
	err := method.CreateMessage(userID, "Order", "Order Created", "Order ID: "+orderID, messageBody, orderID, "")
	if err != nil {
		fmt.Printf("Failed to create order message: %v\n", err)
	}
}

func handleOrderStatusChange(c *gin.Context) {
	orderID := getOrderIDFromRequest(c)
	status := getOrderStatusFromRequest(c)
	if orderID == "" || status == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	var messageTitleOne, messageBody string
	switch status {
	case "shipped":
		messageTitleOne = "Order Shipped"
		messageBody = "Your order has been shipped"
	case "delivered":
		messageTitleOne = "Order Received"
		messageBody = "Your order has been received"
	default:
		return
	}

	err := method.CreateMessage(userID, "Order", messageTitleOne, "Order ID: "+orderID, messageBody, orderID, "")
	if err != nil {
		fmt.Printf("Failed to create order status change message: %v\n", err)
	}
}

func handleOrderCanceled(c *gin.Context) {
	orderID := getOrderIDFromRequest(c)
	if orderID == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	err := method.CreateMessage(userID, "Order", "Order Cancelled", "Order ID: "+orderID, "Your order has been cancelled", orderID, "")
	if err != nil {
		fmt.Printf("Failed to create order cancellation message: %v\n", err)
	}
}

func handleOrderShipped(c *gin.Context) {
	orderID := getOrderIDFromRequest(c)
	if orderID == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	err := method.CreateMessage(userID, "Order", "Order Shipped", "Order ID: "+orderID, "Your order has been shipped", orderID, "")
	if err != nil {
		fmt.Printf("Failed to create order shipping message: %v\n", err)
	}
}

func handleOrderDelivered(c *gin.Context) {
	orderID := getOrderIDFromRequest(c)
	if orderID == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	err := method.CreateMessage(userID, "Order", "Order Received", "Order ID: "+orderID, "Your order has been received", orderID, "")
	if err != nil {
		fmt.Printf("Failed to create order delivery message: %v\n", err)
	}
}

func handleOrderReturnRequested(c *gin.Context) {
	orderID := getOrderIDFromRequest(c)
	if orderID == "" {
		return
	}

	userID := getUserIDFromContext(c)
	if userID <= 0 {
		return
	}

	err := method.CreateMessage(userID, "Order", "After-sales Requested", "Order ID: "+orderID, "You have successfully requested after-sales service", orderID, "")
	if err != nil {
		fmt.Printf("Failed to create after-sales request message: %v\n", err)
	}

	returnOrderID := getReturnOrderIDFromResponse(c)
	if returnOrderID != "" {
		err := method.CreateMessage(userID, "return_order", "Return Order Created", "Return Order ID: "+returnOrderID, "Your return order has been created successfully", returnOrderID, "")
		if err != nil {
			fmt.Printf("Failed to create return order created message: %v\n", err)
		}
	}
}

func handleReturnOrderCreated(c *gin.Context, userID int) {
	returnOrderID := getReturnOrderIDFromResponse(c)
	if returnOrderID == "" {
		return
	}

	// 售后创建成功后单独发一条售后消息，和订单申请消息拆开，方便前端按 return_order 分类展示。
	err := method.CreateMessage(userID, "return_order", "Return Order Created", "Return Order ID: "+returnOrderID, "Your return order has been created successfully", returnOrderID, "")
	if err != nil {
		fmt.Printf("Failed to create return order created message: %v\n", err)
	}
}

func handleReturnOrderShipped(c *gin.Context, userID int) {
	returnOrderID := getReturnOrderIDFromRequest(c)
	fmt.Printf("handleReturnOrderShipped - Got return order ID: %s\n", returnOrderID)

	if returnOrderID == "" {
		fmt.Printf("Return order ID is empty, skipping message creation\n")
		return
	}

	fmt.Printf("handleReturnOrderShipped - Using userID: %d\n", userID)

	fmt.Printf("Return order shipping message - UserID: %d, Return Order ID: %s\n", userID, returnOrderID)
	err := method.CreateMessage(userID, "return_order", "Return Order Shipped", "Return Order ID: "+returnOrderID, "Your return order has been shipped", returnOrderID, "")
	if err != nil {
		fmt.Printf("Failed to create return order shipping message: %v\n", err)
	} else {
		fmt.Printf("Return order shipping message created successfully\n")
	}
}

func handleReturnOrderCompleted(c *gin.Context, userID int) {
	returnOrderID := getReturnOrderIDFromRequest(c)
	fmt.Printf("handleReturnOrderCompleted - Got return order ID: %s, UserID: %d\n", returnOrderID, userID)

	if returnOrderID == "" {
		fmt.Printf("Return order ID is empty, skipping message creation\n")
		return
	}

	fmt.Printf("Creating return order completed message - UserID: %d, Return Order ID: %s\n", userID, returnOrderID)
	err := method.CreateMessage(userID, "return_order", "Return Order Completed", "Return Order ID: "+returnOrderID, "Your return order has been completed", returnOrderID, "")
	if err != nil {
		fmt.Printf("Failed to create return order completed message: %v\n", err)
	} else {
		fmt.Printf("Return order completed message created successfully\n")
	}
}

func handleReturnOrderCanceled(c *gin.Context, userID int) {
	returnOrderID := getReturnOrderIDFromRequest(c)
	fmt.Printf("handleReturnOrderCanceled - Got return order ID: %s, UserID: %d\n", returnOrderID, userID)

	if returnOrderID == "" {
		fmt.Printf("Return order ID is empty, skipping message creation\n")
		return
	}

	fmt.Printf("Creating return order cancellation message - UserID: %d, Return Order ID: %s\n", userID, returnOrderID)
	err := method.CreateMessage(userID, "return_order", "Return Order Cancelled", "Return Order ID: "+returnOrderID, "Your return order has been cancelled", returnOrderID, "")
	if err != nil {
		fmt.Printf("Failed to create return order cancellation message: %v\n", err)
	} else {
		fmt.Printf("Return order cancellation message created successfully\n")
	}
}

func handleReturnOrderApprovedOrRejected(c *gin.Context, userID int) {
	returnOrderID := getReturnOrderIDFromRequest(c)
	fmt.Printf("handleReturnOrderApproved - Got return order ID: %s, UserID: %d\n", returnOrderID, userID)

	if returnOrderID == "" {
		fmt.Printf("Return order ID is empty, skipping message creation\n")
		return
	}

	approveStatus := getReturnOrderApproveStatusFromContext(c)
	title := "Return Order Approved"
	body := "Your return order has been approved"
	if approveStatus == "rejected" {
		title = "Return Order Rejected"
		body = "Your return order has been rejected"
		if remark := getReturnOrderApproveRemarkFromContext(c); remark != "" {
			body += ", reason: " + remark
		}
	}

	fmt.Printf("Creating return order approval/rejection message - UserID: %d, Return Order ID: %s\n", userID, returnOrderID)
	err := method.CreateMessage(userID, "return_order", title, "Return Order ID: "+returnOrderID, body, returnOrderID, "")
	if err != nil {
		fmt.Printf("Failed to create return order approval/rejection message: %v\n", err)
	} else {
		fmt.Printf("Return order approval/rejection message created successfully\n")
	}
}

func getReturnOrderApproveStatusFromContext(c *gin.Context) string {
	if approveStatus, exists := c.Get("return_order_approve_status"); exists {
		if status, ok := approveStatus.(string); ok {
			return status
		}
	}
	return ""
}

func getReturnOrderApproveRemarkFromContext(c *gin.Context) string {
	if approveRemark, exists := c.Get("return_order_approve_remark"); exists {
		if remark, ok := approveRemark.(string); ok {
			return remark
		}
	}
	return ""
}

func getOrderIDFromRequest(c *gin.Context) string {
	if orderID, exists := c.Get("order_id"); exists {
		if id, ok := orderID.(string); ok {
			return id
		}
	}

	return ""
}

func getOrderIDFromResponse(c *gin.Context) string {
	if orderID, exists := c.Get("created_order_id"); exists {
		if id, ok := orderID.(string); ok {
			return id
		}
	}
	return ""
}

func getOrderStatusFromRequest(c *gin.Context) string {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.Status != "" {
		return req.Status
	}

	return ""
}

func getReturnOrderIDFromRequest(c *gin.Context) string {
	if returnOrderID, exists := c.Get("return_order_id"); exists {
		if id, ok := returnOrderID.(string); ok {
			fmt.Printf("Got return order ID from context: %s\n", id)
			return id
		}
	}

	var req struct {
		ReturnOrderID string `json:"return_order_id"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.ReturnOrderID != "" {
		fmt.Printf("Got return order ID from request body: %s\n", req.ReturnOrderID)
		return req.ReturnOrderID
	}

	fmt.Printf("Return order ID not found, returning empty string\n")
	return ""
}

func getReturnOrderIDFromResponse(c *gin.Context) string {
	if returnOrderID, exists := c.Get("created_return_order_id"); exists {
		if id, ok := returnOrderID.(string); ok {
			return id
		}
	}
	return ""
}
