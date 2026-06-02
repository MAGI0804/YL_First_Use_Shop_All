package method

import (
	"Member_shop/db"
	"Member_shop/models"
)

// SearchStyleCode 模糊搜索商品编码
func SearchStyleCode(styleCode string, page, pageSize int) ([]models.StyleCodeData, int64, error) {
	var styleCodes []models.StyleCodeData
	var total int64

	// 构建查询
	query := db.DB.Model(&models.StyleCodeData{}).Where("price > ?", 0)

	// 如果提供了styleCode参数，添加模糊搜索条件
	if styleCode != "" {
		query = query.Where("style_code LIKE ? OR name LIKE ?", "%"+styleCode+"%", "%"+styleCode+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&styleCodes).Error; err != nil {
		return nil, 0, err
	}

	return styleCodes, total, nil
}
