package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/utils"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCommodity(cd *models.Commodity) error {
	if err := db.DB.Create(cd).Error; err != nil {
		log.Printf("创建失败: %v", err)
		return err
	}
	return nil
}

// 工具函数：从商品列表中提取商品ID
func ExtractCommodityIDs(commodities []models.Commodity) []string {
	ids := make([]string, 0, len(commodities))
	for _, commodity := range commodities {
		ids = append(ids, commodity.CommodityID)
	}
	return ids
}

// 工具函数：将商品对象转换为map
func ConvertCommodityToMap(commodity models.Commodity, c *gin.Context) map[string]interface{} {
	result := make(map[string]interface{})
	result["commodity_id"] = commodity.CommodityID
	result["name"] = commodity.Name
	result["style_code"] = commodity.StyleCode
	result["category"] = commodity.Category
	result["category_detail"] = commodity.CategoryDetail
	result["price"] = commodity.Price
	result["size"] = commodity.Size
	result["color"] = commodity.Color
	result["height"] = commodity.Height
	result["spec_code"] = commodity.SpecCode
	result["notes"] = commodity.Notes
	result["created_at"] = commodity.CreatedAt.Format("2006-01-02 15:04:05")

	// 处理图片URL
	if commodity.Image != "" {
		// 获取请求的协议，考虑反向代理环境
		proto := utils.GetRequestProto(c)
		baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
		result["image"] = utils.BuildFullImageURL(baseURL, commodity.Image, "media")
	} else {
		result["image"] = ""
	}

	return result
}

// 工具函数：将商品列表转换为map数组
func ConvertCommoditiesToMap(commodities []models.Commodity, c *gin.Context) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(commodities))
	for _, commodity := range commodities {
		result = append(result, ConvertCommodityToMap(commodity, c))
	}
	return result
}

// 工具函数：将图片列表转换为map数组
func ConvertImagesToMap(images []models.CommodityImage, c *gin.Context) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(images))
	// 获取请求的协议，考虑反向代理环境
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	for _, image := range images {
		imgMap := make(map[string]interface{})
		imgMap["id"] = image.ID
		// 使用BuildFullImageURL函数构建完整URL
		imgMap["url"] = utils.BuildFullImageURL(baseURL, image.Image, "media")
		imgMap["is_main"] = image.IsMain
		result = append(result, imgMap)
	}

	return result
}

type styleCodePageRow struct {
	StyleCode string `gorm:"column:style_code"`
}

func applyCommoditySituationFilters(query *gorm.DB, category interface{}, status string, labelOne, labelTwo, labelThree, labelFour, labelSeven []string) (*gorm.DB, bool) {
	hasFilters := false
	situationQuery := db.DB.Model(&models.CommoditySituation{}).Select("commodity_id")

	if status != "" {
		hasFilters = true
		situationQuery = situationQuery.Where("status = ?", status)
	}
	if category != nil {
		if categoryList, ok := category.([]interface{}); ok {
			stringList := make([]string, 0, len(categoryList))
			for _, cat := range categoryList {
				if strCat, ok := cat.(string); ok {
					stringList = append(stringList, strCat)
				}
			}
			if len(stringList) == 0 {
				return query, hasFilters
			}
			hasFilters = true
			situationQuery = situationQuery.Where("category IN ?", stringList)
		} else if strCat, ok := category.(string); ok && strCat != "" {
			hasFilters = true
			situationQuery = situationQuery.Where("category = ?", strCat)
		}
	}
	if len(labelOne) > 0 {
		hasFilters = true
		situationQuery = situationQuery.Where("label_one IN ?", labelOne)
	}
	if len(labelTwo) > 0 {
		hasFilters = true
		situationQuery = situationQuery.Where("label_two IN ?", labelTwo)
	}
	if len(labelThree) > 0 {
		hasFilters = true
		situationQuery = situationQuery.Where("label_three IN ?", labelThree)
	}
	if len(labelFour) > 0 {
		hasFilters = true
		situationQuery = situationQuery.Where("label_four IN ?", labelFour)
	}
	if len(labelSeven) > 0 {
		hasFilters = true
		situationQuery = situationQuery.Where("label_seven IN ?", labelSeven)
	}
	if !hasFilters {
		return query, false
	}
	return query.Where("commodity_id IN (?)", situationQuery), true
}

func applyStyleStatusFilter(query *gorm.DB, status string) *gorm.DB {
	styleStatus := status
	if styleStatus == "" {
		styleStatus = "online"
	}
	styleQuery := db.DB.Model(&models.StyleCodeSituation{}).
		Select("style_code").
		Where("status = ?", styleStatus)
	return query.Where("style_code IN (?)", styleQuery)
}

func paginateStyleCodeCommodities(query *gorm.DB, page, pageSize int) ([]models.Commodity, int64, int64, error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Distinct("style_code").Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	offset := (page - 1) * pageSize
	var rows []styleCodePageRow
	if err := query.Session(&gorm.Session{}).
		Select("style_code, MAX(created_at) AS latest_created_at").
		Group("style_code").
		Order("latest_created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, 0, err
	}
	if len(rows) == 0 {
		return []models.Commodity{}, total, (total + int64(pageSize) - 1) / int64(pageSize), nil
	}

	styleCodes := make([]string, 0, len(rows))
	styleCodeOrder := make(map[string]int, len(rows))
	for i, row := range rows {
		styleCodes = append(styleCodes, row.StyleCode)
		styleCodeOrder[row.StyleCode] = i
	}

	var candidates []models.Commodity
	if err := query.Session(&gorm.Session{}).
		Where("style_code IN ?", styleCodes).
		Order("created_at DESC").
		Find(&candidates).Error; err != nil {
		return nil, 0, 0, err
	}

	selected := make([]models.Commodity, len(styleCodes))
	seen := make(map[string]bool, len(styleCodes))
	for _, commodity := range candidates {
		if seen[commodity.StyleCode] {
			continue
		}
		if idx, ok := styleCodeOrder[commodity.StyleCode]; ok {
			selected[idx] = commodity
			seen[commodity.StyleCode] = true
		}
	}

	commodities := make([]models.Commodity, 0, len(selected))
	for _, commodity := range selected {
		if commodity.StyleCode != "" {
			commodities = append(commodities, commodity)
		}
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return commodities, total, totalPages, nil
}

func loadStyleSituations(commodities []models.Commodity) map[string]models.StyleCodeSituation {
	styleCodes := make([]string, 0, len(commodities))
	seen := make(map[string]bool, len(commodities))
	for _, commodity := range commodities {
		if commodity.StyleCode == "" || seen[commodity.StyleCode] {
			continue
		}
		seen[commodity.StyleCode] = true
		styleCodes = append(styleCodes, commodity.StyleCode)
	}
	if len(styleCodes) == 0 {
		return map[string]models.StyleCodeSituation{}
	}

	var situations []models.StyleCodeSituation
	if err := db.DB.Where("style_code IN ?", styleCodes).Find(&situations).Error; err != nil {
		log.Printf("load style situations failed: %v", err)
		return map[string]models.StyleCodeSituation{}
	}

	result := make(map[string]models.StyleCodeSituation, len(situations))
	for _, situation := range situations {
		result[situation.StyleCode] = situation
	}
	return result
}

func loadCommodityImages(commodities []models.Commodity) map[string][]models.CommodityImage {
	commodityIDs := make([]string, 0, len(commodities))
	for _, commodity := range commodities {
		if commodity.CommodityID != "" {
			commodityIDs = append(commodityIDs, commodity.CommodityID)
		}
	}
	if len(commodityIDs) == 0 {
		return map[string][]models.CommodityImage{}
	}

	var images []models.CommodityImage
	if err := db.DB.Where("commodity_id IN ?", commodityIDs).Find(&images).Error; err != nil {
		log.Printf("load commodity images failed: %v", err)
		return map[string][]models.CommodityImage{}
	}

	result := make(map[string][]models.CommodityImage, len(commodityIDs))
	for _, image := range images {
		result[image.CommodityID] = append(result[image.CommodityID], image)
	}
	return result
}

func buildCommodityListResult(commodities []models.Commodity, demand string, c *gin.Context, includeStyleStatus bool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(commodities))
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	styleSituations := map[string]models.StyleCodeSituation{}
	if includeStyleStatus {
		styleSituations = loadStyleSituations(commodities)
	}
	commodityImages := loadCommodityImages(commodities)

	for _, commodity := range commodities {
		images := commodityImages[commodity.CommodityID]
		goodsData := make(map[string]interface{})
		if demand == "style_code" || demand == "goods" {
			goodsData["promo_image_url"] = buildPromoImageURL(commodity, images, baseURL)
			goodsData["price"] = commodity.Price
			goodsData["name"] = commodity.Name
			goodsData["style_code"] = commodity.StyleCode
			if includeStyleStatus {
				setStyleSituationFields(goodsData, styleSituations[commodity.StyleCode])
				goodsData["created_at"] = commodity.CreatedAt.Format("2006-01-02 15:04:05")
			}
		} else {
			goodsData["commodity_id"] = commodity.CommodityID
			goodsData["name"] = commodity.Name
			goodsData["style"] = commodity.StyleCode
			goodsData["category"] = commodity.Category
			goodsData["price"] = commodity.Price
			goodsData["promo_image_url"] = ""
			if commodity.PromoImage != "" {
				goodsData["promo_image_url"] = utils.BuildFullImageURL(baseURL, commodity.PromoImage, "media")
			}
			goodsData["created_at"] = commodity.CreatedAt.Format("2006-01-02 15:04:05")
			if includeStyleStatus {
				setStyleSituationFields(goodsData, styleSituations[commodity.StyleCode])
			}
			setCommodityImageFields(goodsData, images, baseURL)
		}
		result = append(result, goodsData)
	}

	return result
}

func buildPromoImageURL(commodity models.Commodity, images []models.CommodityImage, baseURL string) string {
	if commodity.PromoImage != "" {
		return utils.BuildFullImageURL(baseURL, commodity.PromoImage, "media")
	}
	if commodity.Image != "" {
		return utils.BuildFullImageURL(baseURL, commodity.Image, "media")
	}
	for _, image := range images {
		if image.IsMain {
			return utils.BuildFullImageURL(baseURL, image.Image, "media")
		}
	}
	return ""
}

func setStyleSituationFields(goodsData map[string]interface{}, situation models.StyleCodeSituation) {
	goodsData["online_status"] = situation.Status
	if situation.OnlineTime != nil {
		goodsData["online_time"] = situation.OnlineTime.Format("2006-01-02 15:04:05")
	} else {
		goodsData["online_time"] = ""
	}
}

func setCommodityImageFields(goodsData map[string]interface{}, images []models.CommodityImage, baseURL string) {
	imageURLs := make([]map[string]interface{}, 0, len(images))
	var mainImage map[string]interface{}
	otherImages := make([]map[string]interface{}, 0)

	for _, img := range images {
		imgInfo := make(map[string]interface{})
		imgInfo["id"] = img.ID
		imgInfo["url"] = utils.BuildFullImageURL(baseURL, img.Image, "media")
		imgInfo["is_main"] = img.IsMain
		imgInfo["created_at"] = img.CreatedAt.Format("2006-01-02 15:04:05")
		imageURLs = append(imageURLs, imgInfo)
		if img.IsMain {
			mainImage = imgInfo
		} else {
			otherImages = append(otherImages, imgInfo)
		}
	}

	goodsData["images"] = imageURLs
	goodsData["main_image"] = mainImage
	goodsData["other_images"] = otherImages
}

// GetCommodityListByPage 获取商品列表（带分页）
func GetCommodityListByPage(category, keyword string, pageNum, pageSize int) ([]models.Commodity, int64, error) {
	var commodities []models.Commodity
	var totalCount int64

	query := db.DB.Model(&models.Commodity{})

	// 添加分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 添加关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 执行分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&commodities).Error; err != nil {
		return nil, 0, err
	}

	return commodities, totalCount, nil
}

// SearchStyleCodes 搜索款式编码
func SearchStyleCodes(keyword string, category interface{}, page, pageSize int) ([]models.StyleCodeData, int64, error) {
	var styleCodeDataList []models.StyleCodeData
	var totalCount int64

	offset := (page - 1) * pageSize

	// 调试：先看看数据库里所有含关键词的款式，不限制价格
	if keyword != "" {
		var debugList []models.StyleCodeData
		debugQuery := db.DB.Model(&models.StyleCodeData{})
		debugQuery = debugQuery.Where("style_code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		debugQuery.Find(&debugList)
		log.Printf("调试：找到 %d 条含关键词 '%s' 的记录（未限制价格）", len(debugList), keyword)
		for _, item := range debugList {
			log.Printf("  - style_code: %s, name: %s, price: %f", item.StyleCode, item.Name, item.Price)
		}
	}

	// count 查询 - 暂时移除价格过滤以便调试
	countQuery := db.DB.Model(&models.StyleCodeData{})
	// countQuery = countQuery.Where("price > ?", 0)

	if category != nil && category != "" && category != "全部" {
		if categoryList, ok := category.([]interface{}); ok {
			stringList := make([]string, 0, len(categoryList))
			for _, cat := range categoryList {
				if strCat, ok := cat.(string); ok {
					stringList = append(stringList, strCat)
				}
			}
			countQuery = countQuery.Where("category IN ?", stringList)
		} else if strCat, ok := category.(string); ok {
			countQuery = countQuery.Where("category = ?", strCat)
		}
	}

	if keyword != "" {
		countQuery = countQuery.Where("style_code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	log.Printf("调试：count 结果: %d", totalCount)

	// find 查询 - 全新的查询对象，也暂时移除价格过滤
	findQuery := db.DB.Model(&models.StyleCodeData{})
	// findQuery = findQuery.Where("price > ?", 0)

	if category != nil && category != "" && category != "全部" {
		if categoryList, ok := category.([]interface{}); ok {
			stringList := make([]string, 0, len(categoryList))
			for _, cat := range categoryList {
				if strCat, ok := cat.(string); ok {
					stringList = append(stringList, strCat)
				}
			}
			findQuery = findQuery.Where("category IN ?", stringList)
		} else if strCat, ok := category.(string); ok {
			findQuery = findQuery.Where("category = ?", strCat)
		}
	}

	if keyword != "" {
		findQuery = findQuery.Where("style_code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := findQuery.Order("style_code").Offset(offset).Limit(pageSize).Find(&styleCodeDataList).Error; err != nil {
		return nil, 0, err
	}
	log.Printf("调试：find 结果数: %d", len(styleCodeDataList))

	return styleCodeDataList, totalCount, nil
}

// DeleteGoods 删除商品
func DeleteGoods(commodityID string) error {
	// 开始事务
	begin := db.DB.Begin()
	if begin.Error != nil {
		return begin.Error
	}

	// 查询商品
	var commodity models.Commodity
	if err := begin.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		begin.Rollback()
		return err
	}

	// 删除商品图片
	var commodityImages []models.CommodityImage
	begin.Where("commodity_id = ?", commodityID).Find(&commodityImages)
	// 实际项目中应该删除文件

	// 删除数据库记录
	if err := begin.Where("commodity_id = ?", commodityID).Delete(&models.CommodityImage{}).Error; err != nil {
		begin.Rollback()
		return err
	}

	if err := begin.Where("commodity_id = ?", commodityID).Delete(&models.CommoditySituation{}).Error; err != nil {
		begin.Rollback()
		return err
	}

	if err := begin.Delete(&commodity).Error; err != nil {
		begin.Rollback()
		return err
	}

	// 提交事务
	if err := begin.Commit().Error; err != nil {
		begin.Rollback()
		return err
	}

	return nil
}

// GetCommodityNameByID 根据商品ID查询商品名称
func GetCommodityNameByID(commodityID string) (string, error) {
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		return "", err
	}
	return commodity.Name, nil
}

// GetCommodityInfoByID 根据商品ID查询商品完整信息
func GetCommodityInfoByID(commodityID string) (*models.Commodity, error) {
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		return nil, err
	}
	return &commodity, nil
}

// GetCommodityData 获取商品数据
func GetCommodityData(commodityID string, dataList []string, c *gin.Context) (map[string]interface{}, error) {
	// 查询商品
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		return nil, err
	}

	// 构建响应数据
	result := make(map[string]interface{})
	// 获取请求的协议，考虑反向代理环境
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	// 如果指定了字段列表，只返回指定的字段
	if len(dataList) > 0 {
		for _, field := range dataList {
			switch field {
			case "commodity_id":
				result[field] = commodity.CommodityID
			case "name":
				result[field] = commodity.Name
			case "style_code":
				result[field] = commodity.StyleCode
			case "category":
				result[field] = commodity.Category
			case "price":
				result[field] = commodity.Price
			case "size":
				result[field] = commodity.Size
			case "color":
				result[field] = commodity.Color
			case "image":
				if commodity.Image != "" {
					result[field] = utils.BuildFullImageURL(baseURL, commodity.Image, "media")
				} else {
					result[field] = ""
				}
			case "promo_image":
				if commodity.PromoImage != "" {
					result[field] = utils.BuildFullImageURL(baseURL, commodity.PromoImage, "media")
				} else {
					result[field] = ""
				}
			case "created_at":
				result[field] = commodity.CreatedAt.Format("2006-01-02 15:04:05")
			default:
				// 忽略未定义的字段
			}
		}
	}

	// 查询商品图片
	var commodityImages []models.CommodityImage
	if err := db.DB.Where("commodity_id = ?", commodityID).Find(&commodityImages).Error; err != nil {
		log.Printf("获取商品图片失败: %v", err)
	}

	// 构建图片信息
	images := make([]map[string]interface{}, 0, len(commodityImages))
	var mainImage map[string]interface{}
	otherImages := make([]map[string]interface{}, 0)

	for _, img := range commodityImages {
		imgInfo := make(map[string]interface{})
		imgInfo["id"] = img.ID
		imgInfo["url"] = utils.BuildFullImageURL(baseURL, img.Image, "media")
		imgInfo["is_main"] = img.IsMain
		imgInfo["created_at"] = img.CreatedAt.Format("2006-01-02 15:04:05")

		images = append(images, imgInfo)

		if img.IsMain {
			mainImage = imgInfo
		} else {
			otherImages = append(otherImages, imgInfo)
		}
	}

	// 添加图片信息
	result["images"] = images
	result["main_image"] = mainImage
	result["other_images"] = otherImages

	return result, nil
}

// GetCommodityList 获取商品列表
func GetCommodityList(demand, styleCode string, category interface{}, status string, labelOne, labelTwo, labelThree, labelFour, labelSeven []string, beginTime, endTime string, page, pageSize int, c *gin.Context) ([]map[string]interface{}, int64, int64, error) {
	var commoditiesPage []models.Commodity
	var total int64
	var totalPages int64

	query := db.DB.Model(&models.Commodity{}).Where("price > ?", 0).Where("inventory > ?", 0)
	query, _ = applyCommoditySituationFilters(query, category, status, labelOne, labelTwo, labelThree, labelFour, labelSeven)

	if beginTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", beginTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", beginTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour)
				query = query.Where("created_at >= ?", t)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			query = query.Where("created_at >= ?", t)
		}
	}

	if endTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02", endTime, time.Local)
			if err == nil {
				t = t.Add(-8 * time.Hour).Add(24 * time.Hour)
				query = query.Where("created_at < ?", t)
			}
		} else {
			t = t.Add(-8 * time.Hour)
			query = query.Where("created_at < ?", t)
		}
	}

	if demand == "style_code" || demand == "goods" {
		query = applyStyleStatusFilter(query, status)
		if styleCode != "" {
			query = query.Where("style_code = ?", styleCode)
		}
	}

	if demand == "style_code" {
		pageData, totalCount, pageCount, err := paginateStyleCodeCommodities(query, page, pageSize)
		if err != nil {
			return nil, 0, 0, err
		}
		commoditiesPage = pageData
		total = totalCount
		totalPages = pageCount
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, 0, err
		}
		offset := (page - 1) * pageSize
		if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&commoditiesPage).Error; err != nil {
			return nil, 0, 0, err
		}
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	return buildCommodityListResult(commoditiesPage, demand, c, true), total, totalPages, nil
}

// GetCommodityListWX 商品查询（小程序专用，不包含新增的时间筛选和字段）
func GetCommodityListWX(demand, styleCode string, category interface{}, status string, labelOne, labelTwo, labelThree, labelFour, labelSeven []string, page, pageSize int, c *gin.Context) ([]map[string]interface{}, int64, int64, error) {
	var commoditiesPage []models.Commodity
	var total int64
	var totalPages int64

	query := db.DB.Model(&models.Commodity{}).Where("price > ?", 0).Where("inventory > ?", 0)
	query, _ = applyCommoditySituationFilters(query, category, status, labelOne, labelTwo, labelThree, labelFour, labelSeven)

	if demand == "style_code" || demand == "goods" {
		query = applyStyleStatusFilter(query, status)
		if styleCode != "" {
			query = query.Where("style_code = ?", styleCode)
		}
	}

	if demand == "style_code" {
		pageData, totalCount, pageCount, err := paginateStyleCodeCommodities(query, page, pageSize)
		if err != nil {
			return nil, 0, 0, err
		}
		commoditiesPage = pageData
		total = totalCount
		totalPages = pageCount
	} else {
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, 0, err
		}
		offset := (page - 1) * pageSize
		if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&commoditiesPage).Error; err != nil {
			return nil, 0, 0, err
		}
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	return buildCommodityListResult(commoditiesPage, demand, c, false), total, totalPages, nil
}

// UpdateCommodity 更新商品信息
func UpdateCommodity(commodityID string, updateFields map[string]interface{}) ([]string, error) {
	// 查询商品
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		return nil, err
	}

	// 更新允许修改的字段
	updatedFields := make([]string, 0)

	for field, value := range updateFields {
		switch field {
		case "name":
			if strValue, ok := value.(string); ok {
				commodity.Name = strValue
				updatedFields = append(updatedFields, field)
			}
		case "category":
			if strValue, ok := value.(string); ok {
				commodity.Category = strValue
				updatedFields = append(updatedFields, field)
			}
		case "price":
			if floatValue, ok := value.(float64); ok && floatValue > 0 {
				commodity.Price = floatValue
				updatedFields = append(updatedFields, field)
			}
		case "size":
			if strValue, ok := value.(string); ok {
				commodity.Size = strValue
				updatedFields = append(updatedFields, field)
			}
		case "color":
			if strValue, ok := value.(string); ok {
				commodity.Color = strValue
				updatedFields = append(updatedFields, field)
			}
		case "notes":
			if strValue, ok := value.(string); ok {
				commodity.Notes = strValue
				updatedFields = append(updatedFields, field)
			}
		case "style_code":
			if strValue, ok := value.(string); ok {
				commodity.StyleCode = strValue
				updatedFields = append(updatedFields, field)
			}
			// 可以添加更多可更新的字段
		}
	}

	// 保存更新
	if err := db.DB.Save(&commodity).Error; err != nil {
		return nil, err
	}

	return updatedFields, nil
}

// UpdateCommodityStatusOnline 商品上线
func UpdateCommodityStatusOnline(commodityID string) (string, error) {
	// 查询商品状态
	var commoditySituation models.CommoditySituation
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commoditySituation).Error; err != nil {
		return "", err
	}

	// 更新状态
	commoditySituation.Status = "online"
	commoditySituation.OnlineTime = time.Now()

	if err := db.DB.Save(&commoditySituation).Error; err != nil {
		return "", err
	}

	// 格式化上线时间
	formattedTime := commoditySituation.OnlineTime.Format("2006-01-02 15:04:05")

	return formattedTime, nil
}

// UpdateCommodityStatusOffline 商品下线
func UpdateCommodityStatusOffline(commodityID string) (string, error) {
	// 查询商品状态
	var commoditySituation models.CommoditySituation
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commoditySituation).Error; err != nil {
		return "", err
	}

	// 更新状态
	time_now := time.Now()
	commoditySituation.Status = "offline"
	commoditySituation.OfflineTime = &time_now

	if err := db.DB.Save(&commoditySituation).Error; err != nil {
		return "", err
	}

	// 格式化下线时间
	formattedTime := commoditySituation.OfflineTime.Format("2006-01-02 15:04:05")

	return formattedTime, nil
}

// GetCommoditySituation 获取商品状态
func GetCommoditySituation(commodityID string) (map[string]interface{}, error) {
	// 查询商品状态
	var commoditySituation models.CommoditySituation
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commoditySituation).Error; err != nil {
		return nil, err
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"status": commoditySituation.Status,
	}

	// 根据状态返回对应时间
	if commoditySituation.Status == "online" && !commoditySituation.OnlineTime.IsZero() {
		responseData["online_time"] = commoditySituation.OnlineTime.Format("2006-01-02 15:04:05")
		responseData["offline_time"] = ""
	} else if commoditySituation.Status == "offline" && !commoditySituation.OfflineTime.IsZero() {
		responseData["online_time"] = ""
		responseData["offline_time"] = commoditySituation.OfflineTime.Format("2006-01-02 15:04:05")
	} else {
		responseData["online_time"] = ""
		responseData["offline_time"] = ""
	}

	return responseData, nil
}

// GetCommodityDetail 获取商品详情
func GetCommodityDetail(commodityID string, c *gin.Context) (map[string]interface{}, error) {
	// 查询商品
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		return nil, err
	}

	// 查询商品图片
	var commodityImages []models.CommodityImage
	if err := db.DB.Where("commodity_id = ?", commodityID).Find(&commodityImages).Error; err != nil {
		log.Printf("获取商品图片失败: %v", err)
	}

	// 查询商品状态
	var commoditySituation models.CommoditySituation
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commoditySituation).Error; err != nil {
		// 如果没有状态记录，创建一个默认的
		commoditySituation = models.CommoditySituation{
			CommodityID: commodityID,
			Status:      "online",
			SalesVolume: 0,
			StyleCode:   commodity.StyleCode,
		}
		if err := db.DB.Create(&commoditySituation).Error; err != nil {
			log.Printf("创建商品状态记录失败: %v", err)
		}
	}

	// 准备响应数据
	detailMap := ConvertCommodityToMap(commodity, c)
	detailMap["images"] = ConvertImagesToMap(commodityImages, c)
	detailMap["status"] = commoditySituation.Status
	detailMap["sales_volume"] = commoditySituation.SalesVolume

	return detailMap, nil
}

// UpdateStyleCodeStatusOnline 更新款式代码状态为在线
func UpdateStyleCodeStatusOnline(styleCode string) error {
	// 获取当前时间
	currentTime := time.Now()

	// 查找或创建款式状态
	var styleCodeSituation models.StyleCodeSituation
	if err := db.DB.Where("style_code = ?", styleCode).First(&styleCodeSituation).Error; err != nil {
		styleCodeSituation = models.StyleCodeSituation{
			StyleCode:  styleCode,
			Status:     "online",
			OnlineTime: &currentTime,
		}
		if err := db.DB.Create(&styleCodeSituation).Error; err != nil {
			return err
		}
	} else {
		// 更新状态为在线
		styleCodeSituation.Status = "online"
		styleCodeSituation.OnlineTime = &currentTime
		if err := db.DB.Save(&styleCodeSituation).Error; err != nil {
			return err
		}
	}

	return nil
}

// SearchCommoditiesByName 根据名称搜索商品
func SearchCommoditiesByName(keyword string, page, pageSize int) ([]models.Commodity, int64, error) {
	var commodities []models.Commodity
	var totalCount int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询
	query := db.DB.Model(&models.Commodity{}).
		Where("name LIKE ?", "%"+keyword+"%")

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// 执行分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&commodities).Error; err != nil {
		return nil, 0, err
	}

	return commodities, totalCount, nil
}

// BatchGetCommodities 批量获取商品
func BatchGetCommodities(commodityIDs []string) ([]models.Commodity, error) {
	var commodities []models.Commodity
	if err := db.DB.Where("commodity_id IN (?)", commodityIDs).Find(&commodities).Error; err != nil {
		return nil, err
	}
	return commodities, nil
}
