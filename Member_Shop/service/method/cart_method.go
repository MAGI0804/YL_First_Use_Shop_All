package method

import (
	"fmt"
	"log"
	"time"

	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
)

func AddToCart(req requestbody.AddToCartRequest) (*map[string]any, error) {
	var user models.User
	if err := db.DB.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		log.Printf("用户不存在: %d", req.UserID)
		return nil, err
	}

	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", req.CommodityCode).First(&commodity).Error; err != nil {
		log.Printf("商品不存在: %s", req.CommodityCode)
		return nil, err
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	var cart models.Cart
	if err := db.DB.Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		cart = models.Cart{
			UserID:    req.UserID,
			CartItems: make(models.CartItemsMap),
		}
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	if item, exists := cart.CartItems[req.CommodityCode]; exists {
		item.Quantity += quantity
		item.AddedTime = currentTime
		cart.CartItems[req.CommodityCode] = item
	} else {
		cart.CartItems[req.CommodityCode] = models.CartItemJSON{
			Quantity:  quantity,
			AddedTime: currentTime,
		}
	}

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("添加到购物车失败: %v", err)
		return nil, err
	}

	totalQuantity := 0
	for _, item := range cart.CartItems {
		totalQuantity += item.Quantity
	}

	data := map[string]any{
		"commodity_code": req.CommodityCode,
		"quantity":       cart.CartItems[req.CommodityCode].Quantity,
		"total_items":    totalQuantity,
	}

	log.Printf("用户 %d 添加商品 %s 到购物车成功", req.UserID, req.CommodityCode)
	return &data, nil
}

func BatchDeleteFromCart(req requestbody.BatchDeleteFromCartRequest) (int64, []string, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", req.UserID)
		return 0, nil, nil
	}

	var deletedCount int64
	var notExistCodes []string

	if len(req.CommodityCodes) > 0 {
		deletedCount = 0
		for _, code := range req.CommodityCodes {
			if _, exists := cart.CartItems[code]; exists {
				delete(cart.CartItems, code)
				deletedCount++
			} else {
				notExistCodes = append(notExistCodes, code)
			}
		}
	} else {
		deletedCount = int64(len(cart.CartItems))
		cart.CartItems = make(models.CartItemsMap)
	}

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("删除购物车商品失败: %v", err)
		return 0, nil, err
	}

	log.Printf("用户 %d 删除购物车商品成功，删除数量: %d", req.UserID, deletedCount)
	return deletedCount, notExistCodes, nil
}

func QueryCartItems(userID int) ([]map[string]any, int64, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", userID)
		return make([]map[string]any, 0), 0, nil
	}

	var totalQuantity int64
	var cartItems []map[string]any

	for commodityCode, item := range cart.CartItems {
		var commodity models.Commodity
		if err := db.DB.Where("commodity_id = ?", commodityCode).First(&commodity).Error; err != nil {
			continue
		}

		totalQuantity += int64(item.Quantity)

		itemData := map[string]any{
			"commodity_code": commodityCode,
			"quantity":       item.Quantity,
			"added_time":     item.AddedTime,
		}
		cartItems = append(cartItems, itemData)
	}

	log.Printf("成功获取用户 %d 的购物车，共 %d 条记录", userID, len(cartItems))
	return cartItems, totalQuantity, nil
}

func UpdateCartItemQuantity(req requestbody.UpdateCartItemQuantityRequest) (string, int, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", req.UserID)
		return "", 0, err
	}

	cartItem, exists := cart.CartItems[req.CommodityCode]
	if !exists {
		log.Printf("购物车商品不存在: user_id=%d, commodity_code=%s", req.UserID, req.CommodityCode)
		return "", 0, fmt.Errorf("cart item not found")
	}

	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", req.CommodityCode).First(&commodity).Error; err != nil {
		log.Printf("商品不存在: %s", req.CommodityCode)
		return "", 0, err
	}

	cartItem.Quantity = req.Quantity
	cart.CartItems[req.CommodityCode] = cartItem

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("更新购物车商品数量失败: %v", err)
		return "", 0, err
	}

	log.Printf("用户 %d 更新商品 %s 数量为 %d 成功", req.UserID, req.CommodityCode, req.Quantity)
	return req.CommodityCode, req.Quantity, nil
}

func IncreaseCartItemQuantity(req requestbody.IncreaseCartItemQuantityRequest) (string, int, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", req.UserID)
		return "", 0, err
	}

	cartItem, exists := cart.CartItems[req.CommodityCode]
	if !exists {
		log.Printf("购物车商品不存在: user_id=%d, commodity_code=%s", req.UserID, req.CommodityCode)
		return "", 0, fmt.Errorf("cart item not found")
	}

	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", req.CommodityCode).First(&commodity).Error; err != nil {
		log.Printf("商品不存在: %s", req.CommodityCode)
		return "", 0, err
	}

	cartItem.Quantity += 1
	cart.CartItems[req.CommodityCode] = cartItem

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("增加购物车商品数量失败: %v", err)
		return "", 0, err
	}

	log.Printf("用户 %d 增加商品 %s 数量成功，当前数量: %d", req.UserID, req.CommodityCode, cartItem.Quantity)
	return req.CommodityCode, cartItem.Quantity, nil
}

func DecreaseCartItemQuantity(req requestbody.DecreaseCartItemQuantityRequest) (string, int, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", req.UserID)
		return "", 0, err
	}

	cartItem, exists := cart.CartItems[req.CommodityCode]
	if !exists {
		log.Printf("购物车商品不存在: user_id=%d, commodity_code=%s", req.UserID, req.CommodityCode)
		return "", 0, fmt.Errorf("cart item not found")
	}

	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", req.CommodityCode).First(&commodity).Error; err != nil {
		log.Printf("商品不存在: %s", req.CommodityCode)
		return "", 0, err
	}

	var quantity int
	if cartItem.Quantity > 1 {
		cartItem.Quantity -= 1
		cart.CartItems[req.CommodityCode] = cartItem
		quantity = cartItem.Quantity
	} else {
		delete(cart.CartItems, req.CommodityCode)
		quantity = 0
	}

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("减少购物车商品数量失败: %v", err)
		return "", 0, err
	}

	log.Printf("用户 %d 减少商品 %s 数量成功", req.UserID, req.CommodityCode)
	return req.CommodityCode, quantity, nil
}

func ClearCart(userID int) (int64, error) {
	var cart models.Cart
	if err := db.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		log.Printf("购物车不存在: user_id=%d", userID)
		return 0, nil
	}

	clearedCount := int64(len(cart.CartItems))
	cart.CartItems = make(models.CartItemsMap)

	if err := db.DB.Save(&cart).Error; err != nil {
		log.Printf("清空购物车失败: %v", err)
		return 0, err
	}

	log.Printf("用户 %d 清空购物车成功，清除数量: %d", userID, clearedCount)
	return clearedCount, nil
}
