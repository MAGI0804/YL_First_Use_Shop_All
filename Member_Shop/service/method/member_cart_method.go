package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MemberCartResult struct {
	Member        models.Member    `json:"member"`
	CartItems     []map[string]any `json:"cart_items"`
	ItemsCount    int              `json:"items_count"`
	TotalQuantity int64            `json:"total_quantity"`
}

func QueryMemberCart(req requestbody.MemberCartQueryRequest) (*MemberCartResult, error) {
	member, err := ResolveMember(req.MemberID, req.MemberNo, req.Mobile, req.UserID)
	if err != nil {
		return nil, err
	}
	if member.UserID <= 0 {
		return nil, fmt.Errorf("member has no linked user_id")
	}
	items, totalQuantity, err := QueryCartItems(member.UserID)
	if err != nil {
		return nil, err
	}
	return &MemberCartResult{
		Member:        *member,
		CartItems:     items,
		ItemsCount:    len(items),
		TotalQuantity: totalQuantity,
	}, nil
}

func AddMemberCartItem(req requestbody.MemberCartAddRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*MemberCartResult, error) {
	member, err := ResolveMember(req.MemberID, req.MemberNo, req.Mobile, req.UserID)
	if err != nil {
		return nil, err
	}
	if member.UserID <= 0 {
		return nil, fmt.Errorf("member has no linked user_id")
	}
	before, after, err := mutateCart(member.UserID, func(tx *gorm.DB, cart *models.Cart) error {
		var commodity models.Commodity
		if err := resolveCartCommodityTx(tx, req.CommodityCode, &commodity); err != nil {
			return err
		}
		commodityCode := commodity.CommodityID
		current := cart.CartItems[commodityCode]
		current.Quantity += req.Quantity
		current.AddedTime = time.Now().Format("2006-01-02 15:04:05")
		cart.CartItems[commodityCode] = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := RecordBackendOperation(BackendOperationLogInput{
		Operator:   operator,
		Action:     ActionMemberCartAdd,
		Module:     OperationModuleCart,
		TargetType: "cart",
		TargetID:   strconv.Itoa(member.UserID),
		MemberID:   member.ID,
		UserID:     member.UserID,
		BeforeData: before,
		AfterData:  after,
		RequestID:  requestMeta.RequestID,
		ClientIP:   requestMeta.ClientIP,
		UserAgent:  requestMeta.UserAgent,
	}); err != nil {
		return nil, err
	}
	return QueryMemberCart(requestbody.MemberCartQueryRequest{UserID: member.UserID})
}

func UpdateMemberCartItemQuantity(req requestbody.MemberCartUpdateQuantityRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*MemberCartResult, error) {
	member, err := ResolveMember(req.MemberID, req.MemberNo, req.Mobile, req.UserID)
	if err != nil {
		return nil, err
	}
	if member.UserID <= 0 {
		return nil, fmt.Errorf("member has no linked user_id")
	}
	before, after, err := mutateCart(member.UserID, func(tx *gorm.DB, cart *models.Cart) error {
		if _, exists := cart.CartItems[req.CommodityCode]; !exists {
			return fmt.Errorf("cart item not found")
		}
		if req.Quantity == 0 {
			delete(cart.CartItems, req.CommodityCode)
			return nil
		}
		item := cart.CartItems[req.CommodityCode]
		item.Quantity = req.Quantity
		item.AddedTime = time.Now().Format("2006-01-02 15:04:05")
		cart.CartItems[req.CommodityCode] = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := RecordBackendOperation(BackendOperationLogInput{
		Operator:   operator,
		Action:     ActionMemberCartUpdateQty,
		Module:     OperationModuleCart,
		TargetType: "cart",
		TargetID:   strconv.Itoa(member.UserID),
		MemberID:   member.ID,
		UserID:     member.UserID,
		BeforeData: before,
		AfterData:  after,
		RequestID:  requestMeta.RequestID,
		ClientIP:   requestMeta.ClientIP,
		UserAgent:  requestMeta.UserAgent,
	}); err != nil {
		return nil, err
	}
	return QueryMemberCart(requestbody.MemberCartQueryRequest{UserID: member.UserID})
}

func DeleteMemberCartItems(req requestbody.MemberCartDeleteRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*MemberCartResult, error) {
	member, err := ResolveMember(req.MemberID, req.MemberNo, req.Mobile, req.UserID)
	if err != nil {
		return nil, err
	}
	if member.UserID <= 0 {
		return nil, fmt.Errorf("member has no linked user_id")
	}
	before, after, err := mutateCart(member.UserID, func(tx *gorm.DB, cart *models.Cart) error {
		if len(req.CommodityCodes) == 0 {
			cart.CartItems = make(models.CartItemsMap)
			return nil
		}
		for _, code := range req.CommodityCodes {
			delete(cart.CartItems, strings.TrimSpace(code))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := RecordBackendOperation(BackendOperationLogInput{
		Operator:   operator,
		Action:     ActionMemberCartDelete,
		Module:     OperationModuleCart,
		TargetType: "cart",
		TargetID:   strconv.Itoa(member.UserID),
		MemberID:   member.ID,
		UserID:     member.UserID,
		BeforeData: before,
		AfterData:  after,
		RequestID:  requestMeta.RequestID,
		ClientIP:   requestMeta.ClientIP,
		UserAgent:  requestMeta.UserAgent,
	}); err != nil {
		return nil, err
	}
	return QueryMemberCart(requestbody.MemberCartQueryRequest{UserID: member.UserID})
}

func mutateCart(userID int, mutate func(tx *gorm.DB, cart *models.Cart) error) (models.CartItemsMap, models.CartItemsMap, error) {
	var before models.CartItemsMap
	var after models.CartItemsMap
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&cart).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cart = models.Cart{UserID: userID, CartItems: make(models.CartItemsMap)}
			} else {
				return err
			}
		}
		if cart.CartItems == nil {
			cart.CartItems = make(models.CartItemsMap)
		}
		before = cloneCartItems(cart.CartItems)
		if err := mutate(tx, &cart); err != nil {
			return err
		}
		after = cloneCartItems(cart.CartItems)
		return tx.Save(&cart).Error
	})
	return before, after, err
}

func cloneCartItems(items models.CartItemsMap) models.CartItemsMap {
	cloned := make(models.CartItemsMap, len(items))
	for code, item := range items {
		cloned[code] = item
	}
	return cloned
}
