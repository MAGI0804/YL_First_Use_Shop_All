package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"log"
)

// 创建用户信息
func CreateUserData(us *models.UserData) error {
	if err := db.DB.Create(us).Error; err != nil {
		log.Printf("创建环境失败: %v", err)
		return err
	}
	return nil
}

// 查询用户信息
func SelectUserData(UserId int) (map[string]any, error) {
	var user models.UserData
	err := db.DB.Where("user_id=?", UserId).Find(&user).First(&user).Error
	if err != nil {
		return nil, err
	}
	info := map[string]any{
		"user_id":     UserId,
		"data_type":   user.DataType,
		"data_value":  user.DataValue,
		"create_time": user.CreateTime,
	}
	return info, nil
}
