package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const activityImageCacheTTL = 10 * time.Minute

type activityImageListCache struct {
	Items []models.ActivityImage `json:"items"`
	Total int64                  `json:"total"`
}

// CreateActivityImage 创建活动图片记录
func CreateActivityImage(img *models.ActivityImage) error {
	if err := db.DB.Select("status", "image", "category", "notes", "commodities").Create(img).Error; err != nil {
		return err
	}
	InvalidateActivityImageCache()
	return nil
}

// GetActivityImageByID 根据ID获取活动图片
func GetActivityImageByID(id int) (*models.ActivityImage, error) {
	if cached, ok := getCachedActivityImageByID(id); ok {
		return cached, nil
	}

	var activityImg models.ActivityImage
	err := db.DB.Where("id = ?", id).First(&activityImg).Error
	if err != nil {
		return nil, err
	}
	setCachedActivityImageByID(&activityImg)
	return &activityImg, nil
}

// UpdateActivityImage 更新活动图片信息
func UpdateActivityImage(img *models.ActivityImage) error {
	if err := db.DB.Save(img).Error; err != nil {
		return err
	}
	InvalidateActivityImageCache()
	return nil
}

// GetOnlineActivityImageCount 获取已上线活动图片数量
func GetOnlineActivityImageCount() (int64, error) {
	var count int64
	err := db.DB.Model(&models.ActivityImage{}).Where("status = ?", "online").Count(&count).Error
	return count, err
}

// GetMaxOnlineOrder 获取当前最大的上线顺序值
func GetMaxOnlineOrder() (int, error) {
	var maxOrder int
	err := db.DB.Model(&models.ActivityImage{}).Where("status = ?", "online").Select("COALESCE(MAX(`order`), 0)").Scan(&maxOrder).Error
	return maxOrder, err
}

// UpdateActivityImageOnline 活动图片上线，包含上线数量限制检查
func UpdateActivityImageOnline(id int) (*models.ActivityImage, error) {
	// 验证活动图是否存在
	activityImg, err := GetActivityImageByID(id)
	if err != nil {
		return nil, err
	}

	// 检查已上线活动图数量
	onlineCount, err := GetOnlineActivityImageCount()
	if err != nil {
		return nil, err
	}

	// 如果当前活动图不是已上线状态，且已上线数量已达5，则不允许上线
	if activityImg.Status != "online" && onlineCount >= 5 {
		return nil, gorm.ErrInvalidData
	}

	// 保存原始状态用于判断
	originalStatus := activityImg.Status

	// 更新活动图状态和上线时间
	activityImg.Status = "online"
	now := time.Now()
	activityImg.OnlineTime = &now
	activityImg.OfflineTime = nil // 设置为nil避免'0000-00-00'问题

	// 当图片从非上线状态变为上线状态时，自动设置order为当前最大order+1，排到最后
	if originalStatus != "online" {
		// 获取当前最大的顺序值
		maxOrder, err := GetMaxOnlineOrder()
		if err != nil {
			return nil, err
		}
		newOrder := maxOrder + 1
		activityImg.Order = &newOrder
	}

	// 保存更新
	if err := UpdateActivityImage(activityImg); err != nil {
		return nil, err
	}

	return activityImg, nil
}

// UpdateActivityImageOffline 活动图片下线，包含后续图片顺序调整
func UpdateActivityImageOffline(id int) error {
	// 开始事务
	tx := db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 验证活动图是否存在
	var activityImg models.ActivityImage
	if err := tx.Where("id = ?", id).First(&activityImg).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 保存当前图片的order值，用于后续调整
	currentOrder := activityImg.Order

	// 更新活动图状态为下线，并清空order
	activityImg.Status = "offline"
	offlineNow := time.Now()
	activityImg.OfflineTime = &offlineNow
	activityImg.Order = nil

	if err := tx.Save(&activityImg).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果当前图片有order值，则将后续图片的order往前移一位
	if currentOrder != nil {
		// 更新所有order大于当前图片order的图片，order减1
		if err := tx.Model(&models.ActivityImage{}).
			Where("status = ? AND `order` > ?", "online", *currentOrder).
			Update("`order`", gorm.Expr("`order` - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	InvalidateActivityImageCache()
	return nil
}

// BatchUpdateActivityImageOrders 批量更新活动图片顺序
func BatchUpdateActivityImageOrders(orders []struct {
	ID    int
	Order int
}) error {
	// 开始事务
	tx := db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, order := range orders {
		// 验证图片是否存在
		var activityImg models.ActivityImage
		if err := tx.Where("id = ?", order.ID).First(&activityImg).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新顺序
		orderValue := order.Order
		activityImg.Order = &orderValue
		if err := tx.Save(&activityImg).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	InvalidateActivityImageCache()
	return nil
}

// QueryActivityImages 分页查询活动图片，支持状态过滤、时间范围过滤和是否有活动详情过滤
func QueryActivityImages(page, pageSize int, status string, startTime, endTime string, hasActivityDetail *bool) ([]models.ActivityImage, int64, error) {
	if cached, ok := getCachedActivityImageList(page, pageSize, status, startTime, endTime, hasActivityDetail); ok {
		return cached.Items, cached.Total, nil
	}

	var activityImages []models.ActivityImage
	var total int64

	// 构建查询
	query := db.DB.Model(&models.ActivityImage{})

	// 处理状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 处理时间范围过滤
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 处理是否有活动详情过滤
	if hasActivityDetail != nil {
		query = query.Where("has_activity_detail = ?", *hasActivityDetail)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询数据
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&activityImages).Error; err != nil {
		return nil, 0, err
	}

	setCachedActivityImageList(page, pageSize, status, startTime, endTime, hasActivityDetail, activityImages, total)
	return activityImages, total, nil
}

// InvalidateActivityImageCache clears activity image cache after any write.
func InvalidateActivityImageCache() {
	if db.Rds == nil {
		return
	}
	iter := db.Rds.Scan(db.RedisCtx, 0, "cache:activity:*", 100).Iterator()
	for iter.Next(db.RedisCtx) {
		_ = db.Rds.Del(db.RedisCtx, iter.Val()).Err()
	}
}

func getCachedActivityImageByID(id int) (*models.ActivityImage, bool) {
	if db.Rds == nil {
		return nil, false
	}
	value, err := db.Rds.Get(db.RedisCtx, fmt.Sprintf("cache:activity:detail:%d", id)).Result()
	if err == redis.Nil || err != nil {
		return nil, false
	}
	var activityImg models.ActivityImage
	if err := json.Unmarshal([]byte(value), &activityImg); err != nil {
		return nil, false
	}
	return &activityImg, true
}

func setCachedActivityImageByID(img *models.ActivityImage) {
	if db.Rds == nil || img == nil {
		return
	}
	payload, err := json.Marshal(img)
	if err != nil {
		return
	}
	_ = db.Rds.Set(db.RedisCtx, fmt.Sprintf("cache:activity:detail:%d", img.ID), payload, activityImageCacheTTL).Err()
}

func getCachedActivityImageList(page, pageSize int, status, startTime, endTime string, hasActivityDetail *bool) (*activityImageListCache, bool) {
	if db.Rds == nil {
		return nil, false
	}
	value, err := db.Rds.Get(db.RedisCtx, activityImageListCacheKey(page, pageSize, status, startTime, endTime, hasActivityDetail)).Result()
	if err == redis.Nil || err != nil {
		return nil, false
	}
	var cached activityImageListCache
	if err := json.Unmarshal([]byte(value), &cached); err != nil {
		return nil, false
	}
	return &cached, true
}

func setCachedActivityImageList(page, pageSize int, status, startTime, endTime string, hasActivityDetail *bool, items []models.ActivityImage, total int64) {
	if db.Rds == nil {
		return
	}
	payload, err := json.Marshal(activityImageListCache{Items: items, Total: total})
	if err != nil {
		return
	}
	_ = db.Rds.Set(db.RedisCtx, activityImageListCacheKey(page, pageSize, status, startTime, endTime, hasActivityDetail), payload, activityImageCacheTTL).Err()
}

func activityImageListCacheKey(page, pageSize int, status, startTime, endTime string, hasActivityDetail *bool) string {
	detail := "all"
	if hasActivityDetail != nil {
		detail = fmt.Sprintf("%t", *hasActivityDetail)
	}
	return fmt.Sprintf("cache:activity:list:p%d:s%d:status:%s:start:%s:end:%s:detail:%s", page, pageSize, status, startTime, endTime, detail)
}
