package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"log"
)

func CreateStyleCodeData(sc *models.StyleCodeData) error {
	if err := db.DB.Create(sc).Error; err != nil {
		log.Printf("创建失败: %v", err)
		return err
	}
	return nil
}
