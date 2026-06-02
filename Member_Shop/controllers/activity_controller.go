package controllers

import (
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"Member_shop/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityController 活动控制器
type ActivityController struct{}

// AddActivityImg 添加活动图片 - 与Django版本完全匹配
func (ac *ActivityController) AddActivityImg(c *gin.Context) {

	// 确保解析表单数据
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		// 不是multipart/form-data格式也没关系，继续尝试获取表单数据
	}

	// 获取表单数据
	category := c.PostForm("category")
	notes := c.PostForm("notes")
	commodities := c.PostForm("commodities")

	// 处理文件上传
	file, header, err := c.Request.FormFile("image")
	var imagePath string
	if err == nil && file != nil {
		defer file.Close()
		// 保存文件到活动图片目录
		directory := "activities"
		filename := utils.GenerateUniqueFilename(header.Filename)

		// 获取当前工作目录的绝对路径
		currentDir, err := filepath.Abs(utils.MediaRoot())
		if err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取工作目录失败: "+err.Error()))
			return
		}

		// 构建完整的保存路径（使用绝对路径）
		fullDir := filepath.Join(currentDir, directory)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("创建目录失败: "+err.Error()))
			return
		}

		// 保存文件
		savePath := filepath.Join(fullDir, filename)
		if err := c.SaveUploadedFile(header, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("保存文件失败: "+err.Error()))
			return
		}
		// 设置文件权限为 0644（读权限给所有人），确保 nginx 可以访问
		if err := os.Chmod(savePath, 0644); err != nil {
			log.Printf("设置文件权限失败: %v", err)
		}
		// 确保目录也有正确的权限
		dirPath := filepath.Dir(savePath)
		if err := os.Chmod(dirPath, 0755); err != nil {
			log.Printf("设置目录权限失败: %v", err)
		}

		// 验证文件是否成功保存
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("文件保存后验证失败: 文件不存在"))
			return
		}

		// 只保存相对路径到数据库
		imagePath = filepath.Join(directory, filename)
	} else if err != nil && !strings.Contains(err.Error(), "request Content-Type isn't multipart/form-data") {
		// 如果有其他错误但不是Content-Type错误，则报错
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("文件上传失败: "+err.Error()))
		return
	}

	// 创建活动图对象，状态默认为'pending'
	activityImg := models.ActivityImage{
		Status:      "pending",
		Image:       imagePath, // 保存相对路径
		Category:    category,
		Notes:       notes,
		Commodities: commodities,
		// Order字段不设置，默认为null
	}

	// 使用Select方法只保存需要的字段，不包含order字段
	if err := method.CreateActivityImage(&activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("添加失败: "+err.Error()))
		return
	}

	// 构建完整的图片URL返回给前端
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
	fullImageURL := utils.BuildFullImageURL(baseURL, imagePath, "media")

	data := map[string]any{
		"id":    activityImg.ID,
		"image": fullImageURL,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("添加成功", &data))
}

// UpdateActivityImageRelations 更新活动图片关系 - 与Django版本完全匹配
func (ac *ActivityController) UpdateActivityImageRelations(c *gin.Context) {
	// 解析JSON请求体
	var requestData requestbody.ActivityImageRelationRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 获取请求数据
	activityID := int(requestData.ActivityID)

	// 查询活动图是否存在
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 更新分类
	if requestData.Category != "" {
		activityImg.Category = requestData.Category
	}

	// 更新款式编码（将列表转为逗号分隔的字符串）
	if len(requestData.StyleCodes) > 0 {
		// 使用strings.Join将切片转为逗号分隔的字符串
		styleCodesStr := strings.Join(requestData.StyleCodes, ",")
		activityImg.StyleCodes = styleCodesStr
	}

	// 保存更新
	if err := method.UpdateActivityImage(activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"id": activityImg.ID,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("更新成功", &data))
}

// ActivityImageOnline 活动图片上线 - 与Django版本完全匹配
func (ac *ActivityController) ActivityImageOnline(c *gin.Context) {

	// 解析JSON请求体
	var requestData requestbody.ActivityImageStatusRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 获取请求数据
	activityID := int(requestData.ActivityID)

	// 更新活动图片状态为上线
	activityImg, err := method.UpdateActivityImageOnline(activityID)
	if err != nil {
		if err == gorm.ErrInvalidData {
			c.JSON(http.StatusBadRequest, msg.ErrResponseStr("上线失败：最多只能上线5张活动图"))
			return
		}
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("上线失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"id":    activityImg.ID,
		"order": activityImg.Order,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("上线成功", &data))
}

// BatchUpdateActivityImageOrder 批量修改活动图片顺序
func (ac *ActivityController) BatchUpdateActivityImageOrder(c *gin.Context) {

	// 解析JSON请求体
	var requestData requestbody.BatchUpdateActivityImageOrderRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 验证请求数据
	if len(requestData.Images) == 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("缺少图片数据"))
		return
	}

	// 准备批量更新数据
	var orders []struct {
		ID    int
		Order int
	}
	for _, img := range requestData.Images {
		orders = append(orders, struct {
			ID    int
			Order int
		}{
			ID:    img.ID,
			Order: img.Order,
		})
	}

	// 批量更新图片顺序
	if err := method.BatchUpdateActivityImageOrders(orders); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新顺序失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("批量更新顺序成功"))
}

// ActivityImageOffline 活动图片下线 - 与Django版本完全匹配
func (ac *ActivityController) ActivityImageOffline(c *gin.Context) {
	// 解析JSON请求体
	var requestData requestbody.ActivityImageStatusRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 获取请求数据
	activityID := int(requestData.ActivityID)

	// 更新活动图片状态为下线
	if err := method.UpdateActivityImageOffline(activityID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("下线失败: "+err.Error()))
		return
	}

	data := map[string]any{
		"id": activityID,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("下线成功", &data))
}

// BatchQueryActivityImages 批量查询活动图片 - 与Django版本完全匹配
func (ac *ActivityController) BatchQueryActivityImages(c *gin.Context) {
	// 解析JSON请求体
	var requestData requestbody.BatchQueryActivityImagesRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 获取分页参数
	page := int(requestData.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(requestData.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取状态过滤参数
	status := requestData.Status

	// 获取时间范围参数
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	// 获取是否有活动详情过滤参数
	hasActivityDetail := requestData.HasActivityDetail

	// 查询数据
	activityImages, total, err := method.QueryActivityImages(page, pageSize, status, startTime, endTime, hasActivityDetail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询失败: "+err.Error()))
		return
	}

	// 格式化返回结果
	results := make([]map[string]interface{}, 0, len(activityImages))
	for _, img := range activityImages {
		// 解析commodities字段（文本格式，用逗号分隔的商品ID）
		commodityIDs := []int{}
		if img.Commodities != "" {
			ids := strings.Split(img.Commodities, ",")
			for _, idStr := range ids {
				if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
					commodityIDs = append(commodityIDs, id)
				}
			}
		}

		// 解析style_codes字段（文本格式，用逗号分隔的款式编码）
		styleCodes := []string{}
		if img.StyleCodes != "" {
			styleCodes = strings.Split(img.StyleCodes, ",")
			// 去除每个编码的空格
			for i, code := range styleCodes {
				styleCodes[i] = strings.TrimSpace(code)
			}
		}

		// 解析promotional_pics字段（JSON格式的宣传图信息）
		promotionalPics := map[string]interface{}{}
		if img.PromotionalPics != "" {
			if err := json.Unmarshal([]byte(img.PromotionalPics), &promotionalPics); err != nil {
				log.Printf("解析宣传图信息失败: %v", err)
				promotionalPics = map[string]interface{}{}
			}
		}

		// 处理日期格式
		var onlineTime, offlineTime string
		if img.OnlineTime != nil {
			onlineTime = img.OnlineTime.Format("2006-01-02 15:04:05")
		}
		if img.OfflineTime != nil {
			offlineTime = img.OfflineTime.Format("2006-01-02 15:04:05")
		}

		// 构建完整的图片URL
		// 获取请求的协议，考虑反向代理环境
		proto := utils.GetRequestProto(c)
		baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
		// 将Windows路径的反斜杠转换为正斜杠，确保URL可访问
		imagePathWithForwardSlashes := strings.ReplaceAll(img.Image, "\\", "/")
		fullImageURL := utils.BuildFullImageURL(baseURL, imagePathWithForwardSlashes, "media")

		result := map[string]interface{}{
			"id":                  img.ID,
			"image":               fullImageURL, // 添加完整的media路径
			"status":              img.Status,
			"online_time":         onlineTime,
			"offline_time":        offlineTime,
			"commodities":         commodityIDs,
			"style_codes":         styleCodes,
			"category":            img.Category, // 从模型中直接获取
			"notes":               img.Notes,    // 从模型中直接获取
			"promotional_pics":    promotionalPics,
			"has_activity_detail": img.HasActivityDetail,
			"created_at":          img.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":          img.UpdatedAt.Format("2006-01-02 15:04:05"),
			"order":               img.Order, // 从模型中直接获取
		}
		results = append(results, result)
	}

	// 返回分页数据，与Django格式匹配
	data := map[string]any{
		"items":    results,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

// QueryOnlineActivityImages 获取所有已上线的活动图片
func (ac *ActivityController) QueryOnlineActivityImages(c *gin.Context) {
	// 解析JSON请求体
	var requestData requestbody.BatchQueryActivityImagesRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	// 获取分页参数
	page := int(requestData.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(requestData.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	// 固定status为online
	status := "online"

	// 获取时间范围参数
	startTime := requestData.StartTime
	endTime := requestData.EndTime

	// 获取是否有活动详情过滤参数
	hasActivityDetail := requestData.HasActivityDetail

	// 查询数据
	activityImages, total, err := method.QueryActivityImages(page, pageSize, status, startTime, endTime, hasActivityDetail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("查询失败: "+err.Error()))
		return
	}

	// 格式化返回结果
	results := make([]map[string]interface{}, 0, len(activityImages))
	for _, img := range activityImages {
		// 解析commodities字段（文本格式，用逗号分隔的商品ID）
		commodityIDs := []int{}
		if img.Commodities != "" {
			ids := strings.Split(img.Commodities, ",")
			for _, idStr := range ids {
				if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
					commodityIDs = append(commodityIDs, id)
				}
			}
		}

		// 解析style_codes字段（文本格式，用逗号分隔的款式编码）
		styleCodes := []string{}
		if img.StyleCodes != "" {
			styleCodes = strings.Split(img.StyleCodes, ",")
			// 去除每个编码的空格
			for i, code := range styleCodes {
				styleCodes[i] = strings.TrimSpace(code)
			}
		}

		// 解析promotional_pics字段（JSON格式的宣传图信息）
		promotionalPics := map[string]interface{}{}
		if img.PromotionalPics != "" {
			if err := json.Unmarshal([]byte(img.PromotionalPics), &promotionalPics); err != nil {
				log.Printf("解析宣传图信息失败: %v", err)
				promotionalPics = map[string]interface{}{}
			}
		}

		// 处理日期格式
		var onlineTime, offlineTime string
		if img.OnlineTime != nil {
			onlineTime = img.OnlineTime.Format("2006-01-02 15:04:05")
		}
		if img.OfflineTime != nil {
			offlineTime = img.OfflineTime.Format("2006-01-02 15:04:05")
		}

		// 构建完整的图片URL
		// 获取请求的协议，考虑反向代理环境
		proto := utils.GetRequestProto(c)
		baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
		// 将Windows路径的反斜杠转换为正斜杠，确保URL可访问
		imagePathWithForwardSlashes := strings.ReplaceAll(img.Image, "\\", "/")
		fullImageURL := utils.BuildFullImageURL(baseURL, imagePathWithForwardSlashes, "media")

		result := map[string]interface{}{
			"id":                  img.ID,
			"image":               fullImageURL, // 添加完整的media路径
			"status":              img.Status,
			"online_time":         onlineTime,
			"offline_time":        offlineTime,
			"commodities":         commodityIDs,
			"style_codes":         styleCodes,
			"category":            img.Category, // 从模型中直接获取
			"notes":               img.Notes,    // 从模型中直接获取
			"promotional_pics":    promotionalPics,
			"has_activity_detail": img.HasActivityDetail,
			"created_at":          img.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":          img.UpdatedAt.Format("2006-01-02 15:04:05"),
			"order":               img.Order, // 从模型中直接获取
		}
		results = append(results, result)
	}

	// 返回分页数据，与Django格式匹配
	data := map[string]any{
		"items":    results,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

// AddPromotionalPic 新增宣传图
func (ac *ActivityController) AddPromotionalPic(c *gin.Context) {
	// 解析multipart/form-data请求
	activityIDStr := c.PostForm("activity_id")
	if activityIDStr == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("活动图id不能为空"))
		return
	}

	activityID, err := strconv.Atoi(activityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("活动图id格式无效"))
		return
	}

	// 查询活动图
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 处理上传的图片文件
	var imagePath string
	file, header, err := c.Request.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()
		// 保存文件到活动图片目录
		directory := "activities/promotional"
		filename := utils.GenerateUniqueFilename(header.Filename)

		// 获取当前工作目录的绝对路径
		currentDir, err := filepath.Abs(utils.MediaRoot())
		if err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("获取工作目录失败: "+err.Error()))
			return
		}

		// 构建完整的保存路径（使用绝对路径）
		fullDir := filepath.Join(currentDir, directory)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("创建目录失败: "+err.Error()))
			return
		}

		// 保存文件
		savePath := filepath.Join(fullDir, filename)
		if err := c.SaveUploadedFile(header, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("保存文件失败: "+err.Error()))
			return
		}
		// 设置文件权限为 0644（读权限给所有人），确保 nginx 可以访问
		if err := os.Chmod(savePath, 0644); err != nil {
			log.Printf("设置文件权限失败: %v", err)
		}
		// 确保目录也有正确的权限
		dirPath := filepath.Dir(savePath)
		if err := os.Chmod(dirPath, 0755); err != nil {
			log.Printf("设置目录权限失败: %v", err)
		}

		// 验证文件是否成功保存
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("文件保存后验证失败: 文件不存在"))
			return
		}

		// 构建完整的图片URL
		proto := utils.GetRequestProto(c)
		baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
		imagePathWithForwardSlashes := strings.ReplaceAll(filepath.Join(directory, filename), "\\", "/")
		imagePath = utils.BuildFullImageURL(baseURL, imagePathWithForwardSlashes, "media")
	} else if err != nil && !strings.Contains(err.Error(), "request Content-Type isn't multipart/form-data") {
		// 如果有其他错误但不是Content-Type错误，则报错
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("文件上传失败: "+err.Error()))
		return
	}

	if imagePath == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("请上传宣传图"))
		return
	}

	// 解析现有的宣传图信息
	var promotionalPics map[string]interface{}
	if activityImg.PromotionalPics != "" {
		if err := json.Unmarshal([]byte(activityImg.PromotionalPics), &promotionalPics); err != nil {
			log.Printf("解析现有宣传图信息失败: %v", err)
			promotionalPics = map[string]interface{}{}
		}
	} else {
		promotionalPics = map[string]interface{}{}
	}

	// 确定新宣传图的顺序
	newOrder := "1"
	for i := 1; ; i++ {
		orderKey := fmt.Sprintf("%d", i)
		if _, exists := promotionalPics[orderKey]; !exists {
			newOrder = orderKey
			break
		}
	}

	// 添加新宣传图信息
	uploadTime := time.Now().Format("2006-01-02 15:04:05")
	promotionalPics[newOrder] = map[string]interface{}{
		"upload_time": uploadTime,
		"image_url":   imagePath,
	}

	// 转换为JSON字符串
	promotionalPicsJSON, err := json.Marshal(promotionalPics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("序列化宣传图信息失败"))
		return
	}

	// 更新活动图
	activityImg.PromotionalPics = string(promotionalPicsJSON)
	if err := method.UpdateActivityImage(activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新活动图失败"))
		return
	}

	data := map[string]interface{}{
		"order":       newOrder,
		"upload_time": uploadTime,
		"image_url":   imagePath,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("添加宣传图成功", &data))
}

// UpdatePromotionalPicOrder 调整宣传图位置
func (ac *ActivityController) UpdatePromotionalPicOrder(c *gin.Context) {
	var requestData requestbody.UpdatePromotionalPicOrderRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	activityID := int(requestData.ActivityID)
	oldOrder := requestData.OldOrder
	newOrder := requestData.NewOrder

	if oldOrder == 0 || newOrder == 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("顺序参数无效"))
		return
	}

	// 查询活动图
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 解析宣传图信息
	var promotionalPics map[string]interface{}
	if activityImg.PromotionalPics != "" {
		if err := json.Unmarshal([]byte(activityImg.PromotionalPics), &promotionalPics); err != nil {
			log.Printf("解析宣传图信息失败: %v", err)
			promotionalPics = map[string]interface{}{}
		}
	} else {
		promotionalPics = map[string]interface{}{}
	}

	oldOrderKey := fmt.Sprintf("%d", oldOrder)
	newOrderKey := fmt.Sprintf("%d", newOrder)

	// 检查原顺序是否存在
	oldPic, exists := promotionalPics[oldOrderKey]
	if !exists {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("原顺序的宣传图不存在"))
		return
	}

	// 如果原顺序和新顺序相同，直接返回
	if oldOrder == newOrder {
		c.JSON(http.StatusOK, msg.SuccessResponseStr("顺序未变化"))
		return
	}

	// 如果新位置已存在，需要重新排序
	if _, exists := promotionalPics[newOrderKey]; exists {
		// 创建临时有序结构
		type PicItem struct {
			Order int
			Data  interface{}
		}
		var items []PicItem

		// 转换为有序列表
		for k, v := range promotionalPics {
			order, _ := strconv.Atoi(k)
			items = append(items, PicItem{Order: order, Data: v})
		}

		// 重新排序
		var newItems []PicItem
		for _, item := range items {
			if item.Order == oldOrder {
				// 跳过原位置的图片
				continue
			}
			if oldOrder < newOrder {
				// 向后移动：原顺序<新顺序，中间的图片前移
				if item.Order > oldOrder && item.Order <= newOrder {
					item.Order--
				}
			} else {
				// 向前移动：原顺序>新顺序，中间的图片后移
				if item.Order < oldOrder && item.Order >= newOrder {
					item.Order++
				}
			}
			newItems = append(newItems, item)
		}

		// 添加移动的图片到新位置
		newItems = append(newItems, PicItem{Order: newOrder, Data: oldPic})

		// 重建map
		newPromotionalPics := map[string]interface{}{}
		for _, item := range newItems {
			key := fmt.Sprintf("%d", item.Order)
			newPromotionalPics[key] = item.Data
		}
		promotionalPics = newPromotionalPics
	} else {
		// 新位置不存在，直接移动
		delete(promotionalPics, oldOrderKey)
		promotionalPics[newOrderKey] = oldPic
	}

	// 转换为JSON字符串
	promotionalPicsJSON, err := json.Marshal(promotionalPics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("序列化宣传图信息失败"))
		return
	}

	// 更新活动图
	activityImg.PromotionalPics = string(promotionalPicsJSON)
	if err := method.UpdateActivityImage(activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新活动图失败"))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("调整宣传图位置成功"))
}

// DeletePromotionalPic 删除宣传图
func (ac *ActivityController) DeletePromotionalPic(c *gin.Context) {
	var requestData requestbody.DeletePromotionalPicRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	activityID := int(requestData.ActivityID)
	order := requestData.Order

	if order == 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("顺序参数无效"))
		return
	}

	// 查询活动图
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 解析宣传图信息
	var promotionalPics map[string]interface{}
	if activityImg.PromotionalPics != "" {
		if err := json.Unmarshal([]byte(activityImg.PromotionalPics), &promotionalPics); err != nil {
			log.Printf("解析宣传图信息失败: %v", err)
			promotionalPics = map[string]interface{}{}
		}
	} else {
		promotionalPics = map[string]interface{}{}
	}

	orderKey := fmt.Sprintf("%d", order)
	if _, exists := promotionalPics[orderKey]; !exists {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("指定顺序的宣传图不存在"))
		return
	}

	// 删除指定顺序的宣传图
	delete(promotionalPics, orderKey)

	// 重新排序剩余的宣传图
	type PicItem struct {
		Order int
		Data  interface{}
	}
	var items []PicItem

	// 转换为有序列表
	for k, v := range promotionalPics {
		o, _ := strconv.Atoi(k)
		items = append(items, PicItem{Order: o, Data: v})
	}

	// 重建map，重新编号
	newPromotionalPics := map[string]interface{}{}
	for i, item := range items {
		newKey := fmt.Sprintf("%d", i+1)
		newPromotionalPics[newKey] = item.Data
	}

	// 转换为JSON字符串
	promotionalPicsJSON, err := json.Marshal(newPromotionalPics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("序列化宣传图信息失败"))
		return
	}

	// 更新活动图
	activityImg.PromotionalPics = string(promotionalPicsJSON)
	if err := method.UpdateActivityImage(activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新活动图失败"))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("删除宣传图成功"))
}

// GetActivityImageDetail 根据活动图id查询详情
func (ac *ActivityController) GetActivityImageDetail(c *gin.Context) {
	var requestData requestbody.GetActivityImageDetailRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	activityID := int(requestData.ActivityID)

	// 查询活动图
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 解析commodities字段
	var commodityIDs []int
	if activityImg.Commodities != "" {
		ids := strings.Split(activityImg.Commodities, ",")
		for _, idStr := range ids {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				commodityIDs = append(commodityIDs, id)
			}
		}
	}

	// 解析style_codes字段
	var styleCodes []string
	if activityImg.StyleCodes != "" {
		styleCodes = strings.Split(activityImg.StyleCodes, ",")
		for i, code := range styleCodes {
			styleCodes[i] = strings.TrimSpace(code)
		}
	}

	// 解析promotional_pics字段
	var promotionalPics map[string]interface{}
	if activityImg.PromotionalPics != "" {
		if err := json.Unmarshal([]byte(activityImg.PromotionalPics), &promotionalPics); err != nil {
			log.Printf("解析宣传图信息失败: %v", err)
			promotionalPics = map[string]interface{}{}
		}
	}

	// 处理日期格式
	var onlineTime, offlineTime string
	if activityImg.OnlineTime != nil {
		onlineTime = activityImg.OnlineTime.Format("2006-01-02 15:04:05")
	}
	if activityImg.OfflineTime != nil {
		offlineTime = activityImg.OfflineTime.Format("2006-01-02 15:04:05")
	}

	// 构建完整的图片URL
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)
	imagePathWithForwardSlashes := strings.ReplaceAll(activityImg.Image, "\\", "/")
	fullImageURL := utils.BuildFullImageURL(baseURL, imagePathWithForwardSlashes, "media")

	// 构建返回数据
	result := map[string]interface{}{
		"id":                  activityImg.ID,
		"image":               fullImageURL,
		"status":              activityImg.Status,
		"online_time":         onlineTime,
		"offline_time":        offlineTime,
		"commodities":         commodityIDs,
		"style_codes":         styleCodes,
		"category":            activityImg.Category,
		"notes":               activityImg.Notes,
		"promotional_pics":    promotionalPics,
		"has_activity_detail": activityImg.HasActivityDetail,
		"created_at":          activityImg.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":          activityImg.UpdatedAt.Format("2006-01-02 15:04:05"),
		"order":               activityImg.Order,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &result))
}

// SetHasActivityDetail 设置活动详情状态
func (ac *ActivityController) SetHasActivityDetail(c *gin.Context) {
	var requestData requestbody.SetHasActivityDetailRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}

	activityID := int(requestData.ActivityID)

	// 查询活动图
	activityImg, err := method.GetActivityImageByID(activityID)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("活动图不存在"))
		return
	}

	// 更新字段
	activityImg.HasActivityDetail = requestData.HasActivityDetail

	// 保存
	if err := method.UpdateActivityImage(activityImg); err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新活动图失败"))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("设置成功"))
}
