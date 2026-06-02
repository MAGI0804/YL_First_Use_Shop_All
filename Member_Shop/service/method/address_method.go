package method

import (
	"log"

	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
)

func AddAddress(req requestbody.AddAddressRequest) (int, error) {
	var user models.User
	if err := db.DB.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		log.Printf("用户不存在: %d", req.UserID)
		return 0, err
	}

	if req.IsDefault {
		db.DB.Model(&models.Address{}).
			Where("user_id = ? AND is_default = ?", req.UserID, true).
			Update("is_default", false)
	}

	address := models.Address{
		UserID:          req.UserID,
		ReceiverName:    req.ReceiverName,
		PhoneNumber:     req.PhoneNumber,
		Province:        req.Province,
		City:            req.City,
		County:          req.County,
		DetailedAddress: req.DetailedAddress,
		IsDefault:       req.IsDefault,
	}

	if err := db.DB.Create(&address).Error; err != nil {
		log.Printf("新增地址失败: %v", err)
		return 0, err
	}

	log.Printf("用户 %d 新增地址成功: %d", req.UserID, address.AddressID)
	return address.AddressID, nil
}

func DeleteAddress(req requestbody.DeleteAddressRequest) error {
	var address models.Address
	if err := db.DB.Where("address_id = ? AND user_id = ?", req.AddressID, req.UserID).First(&address).Error; err != nil {
		log.Printf("地址不存在或不属于该用户: user_id=%d, address_id=%d", req.UserID, req.AddressID)
		return err
	}

	if err := db.DB.Delete(&address).Error; err != nil {
		log.Printf("删除地址失败: %v", err)
		return err
	}

	log.Printf("用户 %d 删除地址成功: %d", req.UserID, req.AddressID)
	return nil
}

func UpdateAddress(req requestbody.UpdateAddressRequest) error {
	var address models.Address
	if err := db.DB.Where("address_id = ? AND user_id = ?", req.AddressID, req.UserID).First(&address).Error; err != nil {
		log.Printf("地址不存在或不属于该用户: user_id=%d, address_id=%d", req.UserID, req.AddressID)
		return err
	}

	if req.IsDefault {
		db.DB.Model(&models.Address{}).
			Where("user_id = ? AND is_default = ? AND address_id != ?", address.UserID, true, req.AddressID).
			Update("is_default", false)
	}

	hasChanges := false
	if req.Province != "" {
		address.Province = req.Province
		hasChanges = true
	}
	if req.City != "" {
		address.City = req.City
		hasChanges = true
	}
	if req.County != "" {
		address.County = req.County
		hasChanges = true
	}
	if req.DetailedAddress != "" {
		address.DetailedAddress = req.DetailedAddress
		hasChanges = true
	}
	if req.ReceiverName != "" {
		address.ReceiverName = req.ReceiverName
		hasChanges = true
	}
	if req.PhoneNumber != "" {
		address.PhoneNumber = req.PhoneNumber
		hasChanges = true
	}
	address.IsDefault = req.IsDefault
	hasChanges = true

	if hasChanges {
		if err := db.DB.Save(&address).Error; err != nil {
			log.Printf("更新地址失败: %v", err)
			return err
		}
	}

	log.Printf("用户 %d 更新地址成功: %d", req.UserID, req.AddressID)
	return nil
}

func SetDefaultAddress(req requestbody.SetDefaultAddressRequest) error {
	var address models.Address
	if err := db.DB.Where("address_id = ? AND user_id = ?", req.AddressID, req.UserID).First(&address).Error; err != nil {
		log.Printf("地址不存在或不属于该用户: user_id=%d, address_id=%d", req.UserID, req.AddressID)
		return err
	}

	db.DB.Model(&models.Address{}).
		Where("user_id = ? AND is_default = ?", req.UserID, true).
		Update("is_default", false)

	address.IsDefault = true
	if err := db.DB.Save(&address).Error; err != nil {
		log.Printf("设置默认地址失败: %v", err)
		return err
	}

	log.Printf("用户 %d 设置默认地址成功: %d", req.UserID, req.AddressID)
	return nil
}

func GetAddresses(userID int) ([]models.Address, error) {
	var addresses []models.Address
	if err := db.DB.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addresses).Error; err != nil {
		log.Printf("获取用户 %d 的地址列表失败: %v", userID, err)
		return nil, err
	}

	log.Printf("成功获取用户 %d 的地址列表，共 %d 条记录", userID, len(addresses))
	return addresses, nil
}

func GetAddressByID(req requestbody.GetAddressByIDRequest) (*models.Address, error) {
	var address models.Address
	if err := db.DB.Where("address_id = ? AND user_id = ?", req.AddressID, req.UserID).First(&address).Error; err != nil {
		log.Printf("地址不存在或不属于该用户: user_id=%d, address_id=%d", req.UserID, req.AddressID)
		return nil, err
	}

	log.Printf("成功获取用户 %d 的地址详情: %d", req.UserID, req.AddressID)
	return &address, nil
}
