package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"log"
)

// 创建默认的StyleCodeSituation记录
func CreateCommoditySituation(CommodityID string) error {
	styleCodeSituation := models.CommoditySituation{
		CommodityID: CommodityID,
		Status:      "pending", // 默认状态为待审核
	}
	if err := db.DB.Create(&styleCodeSituation).Error; err != nil {
		log.Printf("创建环境失败: %v", err)
		return err
	}
	return nil
}
