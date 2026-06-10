package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateBackendOrder(req requestbody.BackendCreateOrderRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*models.Order, error) {
	return CreateBackendOrderWithAfterCreate(req, operator, requestMeta, nil)
}

func CreateBackendOrderWithAfterCreate(req requestbody.BackendCreateOrderRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta, afterCreate func(*models.Order) error) (*models.Order, error) {
	var createdOrder models.Order
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		member, err := resolveMemberTx(tx, req.MemberID, req.MemberNo, req.Mobile, req.UserID)
		if err != nil {
			return err
		}
		if member.UserID <= 0 {
			return fmt.Errorf("member has no linked user_id")
		}
		if len(req.Items) == 0 {
			return fmt.Errorf("items are required")
		}

		var cart models.Cart
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", member.UserID).First(&cart).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("cart not found")
			}
			return err
		}
		if cart.CartItems == nil {
			cart.CartItems = make(models.CartItemsMap)
		}
		cartBefore := cloneCartItems(cart.CartItems)

		productList := make([]interface{}, 0, len(req.Items))
		productNames := make([]string, 0, len(req.Items))
		orderAmount := req.OrderAmount
		var calculatedAmount float64
		for _, item := range req.Items {
			code := strings.TrimSpace(item.CommodityCode)
			if code == "" || item.Quantity <= 0 {
				return fmt.Errorf("invalid order item")
			}
			cartItem, exists := cart.CartItems[code]
			if !exists || cartItem.Quantity < item.Quantity {
				return fmt.Errorf("cart quantity not enough for %s", code)
			}
			var commodity models.Commodity
			if err := tx.Where("commodity_id = ?", code).First(&commodity).Error; err != nil {
				return err
			}
			price := item.Price
			if price <= 0 {
				price = commodity.Price
			}
			lineAmount := price * float64(item.Quantity)
			calculatedAmount += lineAmount
			productNames = append(productNames, commodity.Name)
			productList = append(productList, map[string]interface{}{
				"commodity_id": code,
				"product_name": commodity.Name,
				"name":         commodity.Name,
				"qty":          item.Quantity,
				"price":        price,
				"sub_amount":   lineAmount,
			})

			cartItem.Quantity -= item.Quantity
			if cartItem.Quantity <= 0 {
				delete(cart.CartItems, code)
			} else {
				cart.CartItems[code] = cartItem
			}
		}
		if orderAmount <= 0 {
			orderAmount = calculatedAmount
		}
		productListJSON, err := json.Marshal(productList)
		if err != nil {
			return err
		}
		productNamesJSON, err := json.Marshal(productNames)
		if err != nil {
			return err
		}

		order := models.Order{
			OrderID:            GenerateOrderNo(),
			UserID:             member.UserID,
			ReceiverName:       strings.TrimSpace(req.ReceiverName),
			ReceiverPhone:      strings.TrimSpace(req.ReceiverPhone),
			Province:           strings.TrimSpace(req.Province),
			City:               strings.TrimSpace(req.City),
			County:             strings.TrimSpace(req.County),
			DetailedAddress:    strings.TrimSpace(req.DetailedAddress),
			OrderAmount:        orderAmount,
			FinalPayAmount:     orderAmount,
			ProductList:        string(productListJSON),
			ProdoctNameList:    string(productNamesJSON),
			ExpressCompany:     strings.TrimSpace(req.ExpressCompany),
			ExpressNumber:      strings.TrimSpace(req.ExpressNumber),
			Status:             "pending",
			PayStatus:          "unpaid",
			OrderTime:          time.Now(),
			Remarks:            strings.TrimSpace(req.Remark),
			CreatedByBackend:   true,
			BackendOperatorID:  operator.ID,
			BackendOrderRemark: strings.TrimSpace(req.BackendRemark),
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if orderAmount > 0 {
			if err := tx.Model(&models.Member{}).Where("id = ?", member.ID).
				Update("total_order_amount", gorm.Expr("total_order_amount + ?", orderAmount)).Error; err != nil {
				return err
			}
		}

		inventoryItems, err := ParseOrderInventoryItems(productList)
		if err != nil {
			return err
		}
		var subOrderIDs []string
		var createdSubOrders []models.SubOrder
		for index, item := range productList {
			productMap := item.(map[string]interface{})
			productInfoJSON, _ := json.Marshal(productMap)
			subOrder, err := createSubOrderTx(
				tx,
				order.OrderID,
				productMap["product_name"].(string),
				string(productInfoJSON),
				productMap["sub_amount"].(float64),
				inventoryItems[index].CommodityID,
				inventoryItems[index].Qty,
			)
			if err != nil {
				return err
			}
			createdSubOrders = append(createdSubOrders, *subOrder)
			subOrderIDs = append(subOrderIDs, subOrder.SubOrderID+":pending")
		}
		if err := DeductInventoryForOrder(tx, order.OrderID, createdSubOrders); err != nil {
			return err
		}
		subOrderIDsJSON, _ := json.Marshal(subOrderIDs)
		order.SubOrderIDs = string(subOrderIDsJSON)
		if err := tx.Model(&order).Update("sub_order_ids", order.SubOrderIDs).Error; err != nil {
			return err
		}
		if err := tx.Save(&cart).Error; err != nil {
			return err
		}
		cartAfter := cloneCartItems(cart.CartItems)
		if err := recordBackendOperation(tx, BackendOperationLogInput{
			Operator:   operator,
			Action:     ActionOrderBackendCreate,
			Module:     OperationModuleOrder,
			TargetType: "order",
			TargetID:   order.OrderID,
			MemberID:   member.ID,
			UserID:     member.UserID,
			OrderID:    order.OrderID,
			AfterData:  orderOperationSnapshot(order),
			RequestID:  requestMeta.RequestID,
			ClientIP:   requestMeta.ClientIP,
			UserAgent:  requestMeta.UserAgent,
			Remark:     req.BackendRemark,
		}); err != nil {
			return err
		}
		if err := recordBackendOperation(tx, BackendOperationLogInput{
			Operator:   operator,
			Action:     ActionMemberCartClearAfterOrder,
			Module:     OperationModuleCart,
			TargetType: "cart",
			TargetID:   strconv.Itoa(member.UserID),
			MemberID:   member.ID,
			UserID:     member.UserID,
			OrderID:    order.OrderID,
			BeforeData: cartBefore,
			AfterData:  cartAfter,
			RequestID:  requestMeta.RequestID,
			ClientIP:   requestMeta.ClientIP,
			UserAgent:  requestMeta.UserAgent,
		}); err != nil {
			return err
		}
		if afterCreate != nil {
			if err := afterCreate(&order); err != nil {
				return err
			}
		}
		createdOrder = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdOrder, nil
}

func resolveMemberTx(tx *gorm.DB, memberID uint, memberNo, mobile string, userID int) (*models.Member, error) {
	query := tx.Model(&models.Member{})
	if memberID > 0 {
		query = query.Where("id = ?", memberID)
	} else if strings.TrimSpace(memberNo) != "" {
		query = query.Where("member_no = ?", strings.TrimSpace(memberNo))
	} else if strings.TrimSpace(mobile) != "" {
		query = query.Where("mobile = ?", strings.TrimSpace(mobile))
	} else if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else {
		return nil, fmt.Errorf("missing member query condition")
	}
	var member models.Member
	if err := query.First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func orderOperationSnapshot(order models.Order) map[string]any {
	return map[string]any{
		"order_id":             order.OrderID,
		"user_id":              order.UserID,
		"order_amount":         order.OrderAmount,
		"final_pay_amount":     order.FinalPayAmount,
		"status":               order.Status,
		"pay_status":           order.PayStatus,
		"created_by_backend":   order.CreatedByBackend,
		"backend_operator_id":  order.BackendOperatorID,
		"backend_order_remark": order.BackendOrderRemark,
	}
}
