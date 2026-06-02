package controllers

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/service/method"
	"Member_shop/service/msg"
	"Member_shop/utils"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommodityController struct{}

// 新增商品
func (cc *CommodityController) AddGoods(t *gin.Context) {
	var req requestbody.AddGoodsRequestBody
	//校验参数
	if err := t.ShouldBind(&req); err != nil {
		fmt.Println(req)
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求格式"))
		return
	}
	//检查商品id是否存在
	if ex := method.SearchExistence("Commodity_data", "commodity_id", req.CommodityID); ex == true {
		t.JSON(http.StatusBadRequest, msg.ErrResponseStr("该商品已存在"))
		return
	}
	if ex := method.SearchExistence("StyleCode_Data", "style_code", req.StyleCode); ex != true {
		var styleCodeData models.StyleCodeData
		styleCodeData = models.StyleCodeData{
			StyleCode:       req.StyleCode,
			Name:            req.Name,
			Category:        req.Category,
			Price:           req.Price,
			Image:           "",   // 暂时为空，实际应从上传文件中获取
			DisplayPictures: "{}", // 初始化为空JSON对象，符合MySQL JSON类型要求
		}
		//创建StyleCodeData记录
		if err := method.CreateStyleCodeData(&styleCodeData); err != nil {
			t.JSON(http.StatusInternalServerError, msg.ErrResponse("创建商品失败", err))
			return
		}
		//创建StyleCodeSituation记录
		if err := method.CreateStyleCodeSituation(req.StyleCode); err != nil {
			t.JSON(http.StatusInternalServerError, msg.ErrResponse("创建商品失败", err))
			return
		}
	}
	// 创建商品对象
	commodity := models.Commodity{
		CommodityID: req.CommodityID,
		Name:        req.Name,
		Price:       req.Price,
		Category:    req.Category,
		StyleCode:   req.StyleCode,
		Size:        req.Size,
		Notes:       req.Notes,
		// 确保image字段不为空
		Image: "",
	}
	// 开始事务
	begin := db.DB.Begin()
	if begin.Error != nil {
		log.Printf("开启事务失败: %v", begin.Error)
		t.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	// 处理文件上传
	var mainImageURL string
	var numImages int = 0
	mainImageFile, err := t.FormFile("image")
	if err == nil && mainImageFile != nil {
		// 获取协议
		proto := utils.GetRequestProto(t)
		// 构建基础URL
		baseURL := fmt.Sprintf("%s://%s", proto, t.Request.Host)

		// 为每个商品ID创建独立的子文件夹
		commodityDir := filepath.Join("commodities", req.CommodityID)
		// 保存上传的文件
		imagePath, err := utils.SaveUploadedFile(t, mainImageFile, commodityDir, "commodity_main_")
		if err != nil {
			begin.Rollback()
			t.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "上传主图失败: " + err.Error(),
			})
			return
		}
		// 实际保存文件到指定路径
		fullPath := utils.MediaPath(imagePath)
		if err := t.SaveUploadedFile(mainImageFile, fullPath); err != nil {
			begin.Rollback()
			t.JSON(http.StatusInternalServerError, msg.ErrResponse("保存主图失败", err))
			return
		}
		if err := os.Chmod(fullPath, 0755); err != nil {
			log.Printf("设置文件权限失败: %v", err)
		}
		// 构建完整的图片URL
		mainImageURL = utils.BuildFullImageURL(baseURL, imagePath, "media")
		// 更新商品的图片URL
		commodity.Image = mainImageURL
		// 设置图片数量
		numImages = 1

		// 如果提供了StyleCode，同时更新StyleCodeData的图片
		if req.StyleCode != "" {
			var styleCodeData models.StyleCodeData
			if err := begin.Where("style_code = ?", req.StyleCode).First(&styleCodeData).Error; err == nil {
				styleCodeData.Image = mainImageURL
				if err := begin.Save(&styleCodeData).Error; err != nil {
					begin.Rollback()
					log.Printf("更新款式图片失败: %v", err)
					t.JSON(http.StatusInternalServerError, msg.ErrResponse("添加商品失败: 更新款式图片失败", err))
					return
				}
			}
		}

	}
	// 提交事务
	if err := begin.Commit().Error; err != nil {
		begin.Rollback()
		log.Printf("提交事务失败: %v", err)
		t.JSON(http.StatusInternalServerError, msg.ErrResponse("添加商品失败", err))
		return
	}
	//创建商品
	if err := method.CreateCommodity(&commodity); err != nil {
		begin.Rollback()
		log.Printf("创建商品失败: %v", err)
		t.JSON(http.StatusInternalServerError, msg.ErrResponse("添加商品失败", err))
		return
	}

	//创建商品状态记录
	if err := method.CreateStyleCodeSituation(req.CommodityID); err != nil {
		begin.Rollback()
		log.Printf("创建商品状态记录失败: %v", err)
		t.JSON(http.StatusInternalServerError, msg.ErrResponse("添加商品失败", err))
		return
	}
	info := map[string]any{
		"code":    200,
		"message": "添加成功",
		"data": gin.H{"commodity_id": req.CommodityID,
			"image_count": numImages,
		},
	}
	t.JSON(http.StatusOK, msg.SuccessResponse("添加成功", &info))
	return
}

// CommodityList 获取商品列表
func (cc *CommodityController) CommodityList(c *gin.Context) {

	// 获取查询参数
	pageNum, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	keyword := c.Query("keyword")

	// 调用封装的函数获取商品列表
	commodities, totalCount, err := method.GetCommodityListByPage(category, keyword, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("获取商品列表失败", err))
		return
	}

	// 格式化返回结果
	data := map[string]any{
		"items":     commodities,
		"total":     totalCount,
		"page":      pageNum,
		"page_size": pageSize,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

// SearchStyleCodes 搜索款式编码名称
func (cc *CommodityController) SearchStyleCodes(c *gin.Context) {
	var requestData requestbody.SearchStyleCodesRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的请求数据"))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 设置默认分页参数
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	} else if requestData.PageSize > 50 {
		requestData.PageSize = 50
	}

	// 调用封装的函数搜索款式编码，使用原始的搜索方法确保结果完整性
	styleCodeDataList, totalCount, err := method.SearchStyleCodes(requestData.SearchKeyword, requestData.Category, requestData.Page, requestData.PageSize)
	if err != nil {
		log.Printf("搜索款式编码失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	// 构建响应数据，与GetCommodityList的style_code模式一致
	result := make([]map[string]interface{}, 0, len(styleCodeDataList))
	// 获取请求的协议，考虑反向代理环境
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	for _, styleData := range styleCodeDataList {
		styleInfo := make(map[string]interface{})

		// 构建promo_image_url，当为空时使用image的值
		var promoImageURL string
		if styleData.Image != "" {
			promoImageURL = utils.BuildFullImageURL(baseURL, styleData.Image, "media")
		} else {
			promoImageURL = ""
		}

		// 查询款式上线情况
		var styleSituation models.StyleCodeSituation
		var onlineStatus string
		var onlineTime string
		db.DB.Where("style_code = ?", styleData.StyleCode).First(&styleSituation)
		if styleSituation.Status != "" {
			onlineStatus = styleSituation.Status
			if styleSituation.OnlineTime != nil {
				onlineTime = styleSituation.OnlineTime.Format("2006-01-02 15:04:05")
			}
		} else {
			onlineStatus = ""
			onlineTime = ""
		}

		styleInfo["promo_image_url"] = promoImageURL
		styleInfo["price"] = styleData.Price
		styleInfo["name"] = styleData.Name
		styleInfo["style_code"] = styleData.StyleCode
		styleInfo["created_at"] = styleData.CreatedAt.Format("2006-01-02 15:04:05")
		styleInfo["online_status"] = onlineStatus
		styleInfo["online_time"] = onlineTime

		result = append(result, styleInfo)
	}

	// 计算总页数
	totalPages := (totalCount + int64(requestData.PageSize) - 1) / int64(requestData.PageSize)

	// 返回与GoodsQuery方法一致的格式
	responseData := map[string]any{
		"data":      result,
		"total":     totalCount,
		"page":      requestData.Page,
		"page_size": requestData.PageSize,
		"pages":     totalPages,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse(fmt.Sprintf("查询成功，共%d件商品", totalCount), &responseData))
}

// DeleteGoods 删除商品
func (cc *CommodityController) DeleteGoods(c *gin.Context) {
	var requestData requestbody.DeleteGoodsRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数删除商品
	if err := method.DeleteGoods(commodityIDStr); err != nil {
		log.Printf("删除商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器处理失败", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("删除成功"))
}

// SearchCommodityData 查询商品信息
func (cc *CommodityController) SearchCommodityData(c *gin.Context) {
	var requestData requestbody.SearchCommodityDataRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数获取商品数据
	result, err := method.GetCommodityData(commodityIDStr, requestData.DataList, c)
	if err != nil {
		log.Printf("查询商品失败，commodity_id: %s, 错误: %v", commodityIDStr, err)
		c.JSON(http.StatusNotFound, msg.ErrResponse("商品不存在", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &result))
}

// GoodsQuery 商品查询
func (cc *CommodityController) GoodsQuery(c *gin.Context) {
	var requestData requestbody.GoodsQueryRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的JSON格式", err))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 设置默认分页参数
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	} else if requestData.PageSize > 50 {
		requestData.PageSize = 50
	}

	// 当类别为"全部"时，设为nil以返回全部内容
	var categoryParam interface{}
	if requestData.Category != "全部" && requestData.Category != nil {
		categoryParam = requestData.Category
	}
	// 调用封装的函数查询商品，传递标签参数
	result, total, totalPages, err := method.GetCommodityList(
		requestData.Demand,
		requestData.StyleCode,
		categoryParam,
		requestData.Status,
		requestData.LabelOne,
		requestData.LabelTwo,
		requestData.LabelThree,
		requestData.LabelFour,
		requestData.LabelSeven,
		requestData.BeginTime,
		requestData.EndTime,
		requestData.Page,
		requestData.PageSize,
		c,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	// 返回与Django版本完全一致的格式
	responseData := map[string]any{
		"data":      result,
		"total":     total,
		"page":      requestData.Page,
		"page_size": requestData.PageSize,
		"pages":     totalPages,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse(fmt.Sprintf("查询成功，共%d件商品", total), &responseData))
}

// GoodsQueryWX 商品查询（小程序专用，不包含新增的时间筛选和字段）
func (cc *CommodityController) GoodsQueryWX(c *gin.Context) {
	var requestData requestbody.GoodsQueryRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的JSON格式", err))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 设置默认分页参数
	if requestData.Page <= 0 {
		requestData.Page = 1
	}
	if requestData.PageSize <= 0 {
		requestData.PageSize = 20
	} else if requestData.PageSize > 50 {
		requestData.PageSize = 50
	}

	// 当类别为"全部"时，设为nil以返回全部内容
	var categoryParam interface{}
	if requestData.Category != "全部" && requestData.Category != nil {
		categoryParam = requestData.Category
	}
	// 调用封装的函数查询商品，传递标签参数
	result, total, totalPages, err := method.GetCommodityListWX(
		requestData.Demand,
		requestData.StyleCode,
		categoryParam,
		requestData.Status,
		requestData.LabelOne,
		requestData.LabelTwo,
		requestData.LabelThree,
		requestData.LabelFour,
		requestData.LabelSeven,
		requestData.Page,
		requestData.PageSize,
		c,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	// 返回与Django版本完全一致的格式
	responseData := map[string]any{
		"data":      result,
		"total":     total,
		"page":      requestData.Page,
		"page_size": requestData.PageSize,
		"pages":     totalPages,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse(fmt.Sprintf("查询成功，共%d件商品", total), &responseData))
}

// ChangeCommodityData 修改商品信息
func (cc *CommodityController) ChangeCommodityData(c *gin.Context) {
	var requestData requestbody.ChangeCommodityDataRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数更新商品信息
	updatedFields, err := method.UpdateCommodity(commodityIDStr, requestData.UpdateFields)
	if err != nil {
		log.Printf("更新商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器处理失败", err))
		return
	}

	responseData := map[string]any{
		"updated_fields": updatedFields,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("更新成功", &responseData))
}

// ChangeCommodityStatusOnline 商品上线
func (cc *CommodityController) ChangeCommodityStatusOnline(c *gin.Context) {
	var requestData requestbody.ChangeCommodityStatusRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数更新商品状态
	formattedTime, err := method.UpdateCommodityStatusOnline(commodityIDStr)
	if err != nil {
		log.Printf("更新商品状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器处理失败", err))
		return
	}

	responseData := map[string]any{
		"online_time": formattedTime,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("商品上线成功", &responseData))
}

// ChangeCommodityStatusOffline 商品下线
func (cc *CommodityController) ChangeCommodityStatusOffline(c *gin.Context) {
	var requestData requestbody.ChangeCommodityStatusRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数更新商品状态
	formattedTime, err := method.UpdateCommodityStatusOffline(commodityIDStr)
	if err != nil {
		log.Printf("更新商品状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器处理失败", err))
		return
	}

	responseData := map[string]any{
		"offline_time": formattedTime,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("商品下线成功", &responseData))
}

// GetCommodityStatus 获取商品状态
func (cc *CommodityController) GetCommodityStatus(c *gin.Context) {
	var requestData requestbody.GetCommodityStatusRequestBody

	// 尝试从JSON和查询参数中绑定数据
	if err := c.ShouldBind(&requestData); err != nil {
		// 检查URL查询参数中是否有commodity_id
		commodityID := c.Query("commodity_id")
		if commodityID == "" {
			// 如果URL查询参数中也没有，返回详细错误信息
			c.JSON(http.StatusBadRequest, msg.ErrResponse("缺少commodity_id参数，请在请求体或URL查询参数中提供", err))
			return
		}
		// 使用URL查询参数中的commodity_id
		requestData.CommodityID = commodityID
	}

	// 将CommodityID转换为字符串
	commodityIDStr := ""
	log.Printf("接收到的commodity_id参数: %v, 类型: %T", requestData.CommodityID, requestData.CommodityID)
	switch v := requestData.CommodityID.(type) {
	case string:
		commodityIDStr = v
		log.Printf("commodity_id参数是字符串类型: %s", commodityIDStr)
	case int, int64:
		// 整数类型直接转换为字符串
		commodityIDStr = fmt.Sprintf("%d", v)
		log.Printf("commodity_id参数是整数类型，转换为字符串: %s", commodityIDStr)
	case float64:
		// 浮点数类型，判断是否为整数
		if v == float64(int64(v)) {
			// 如果是整数，转换为整数格式的字符串
			commodityIDStr = fmt.Sprintf("%.0f", v)
			log.Printf("commodity_id参数是浮点整数类型，转换为整数格式字符串: %s", commodityIDStr)
		} else {
			// 非整数浮点数，保持原样
			commodityIDStr = fmt.Sprintf("%v", v)
			log.Printf("commodity_id参数是浮点数类型，转换为字符串: %s", commodityIDStr)
		}
	default:
		log.Printf("commodity_id参数格式不正确，类型: %T", v)
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("commodity_id参数格式不正确"))
		return
	}

	// 调用封装的函数获取商品状态
	responseData, err := method.GetCommoditySituation(commodityIDStr)
	if err != nil {
		log.Printf("查询商品失败，commodity_id: %s, 错误: %v", commodityIDStr, err)
		c.JSON(http.StatusNotFound, msg.ErrResponse("商品不存在", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &responseData))
}

// CommodityDetail 获取商品详情
func (cc *CommodityController) CommodityDetail(c *gin.Context) {
	commodityIDStr := c.Param("id")
	commodityID, err := strconv.Atoi(commodityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的商品ID"))
		return
	}

	// 调用封装的函数获取商品详情
	detailMap, err := method.GetCommodityDetail(strconv.Itoa(commodityID), c)
	if err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponse("商品不存在", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("获取成功", &detailMap))
}

// CommodityCreate 创建商品
func (cc *CommodityController) CommodityCreate(c *gin.Context) {
	var requestData models.Commodity
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的JSON格式", err))
		return
	}

	// 验证必要字段
	if requestData.Name == "" || requestData.Price <= 0 {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("缺少必要的商品信息"))
		return
	}

	// 创建商品
	commodity := models.Commodity{
		Name:      requestData.Name,
		StyleCode: requestData.StyleCode,
		Category:  requestData.Category,
		Price:     requestData.Price,
		Size:      requestData.Size,
		Color:     requestData.Color,
		Image:     "default.png", // 默认图片路径
	}

	// 调用封装的函数创建商品
	if err := method.CreateCommodity(&commodity); err != nil {
		log.Printf("创建商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	// 创建商品状态
	commoditySituation := models.CommoditySituation{
		CommodityID: commodity.CommodityID,
		Status:      "online", // 设置为在线状态
	}

	if err := db.DB.Create(&commoditySituation).Error; err != nil {
		log.Printf("创建商品状态失败: %v", err)
	}

	// 如果有款式代码，处理款式相关数据
	if requestData.StyleCode != "" {
		// 查找或创建款式数据
		var styleCodeData models.StyleCodeData
		if err := db.DB.Where("style_code = ?", requestData.StyleCode).First(&styleCodeData).Error; err != nil {
			styleCodeData = models.StyleCodeData{
				StyleCode:       requestData.StyleCode,
				Name:            "",
				Category:        "",
				Price:           0,
				Image:           "",
				DisplayPictures: "{}", // 初始化为空JSON对象，符合MySQL JSON类型要求
			}
			if err := method.CreateStyleCodeData(&styleCodeData); err != nil {
				log.Printf("创建款式数据失败: %v", err)
			}
		}

		// 查找或创建款式状态
		var styleCodeSituation models.StyleCodeSituation
		if err := db.DB.Where("style_code = ?", requestData.StyleCode).First(&styleCodeSituation).Error; err != nil {
			styleCodeSituation = models.StyleCodeSituation{
				StyleCode: requestData.StyleCode,
				Status:    "online",
			}
			if err := method.CreateStyleCodeSituation(requestData.StyleCode); err != nil {
				log.Printf("创建款式状态失败: %v", err)
			}
		}
	}

	data := map[string]any{
		"commodity_id": commodity.CommodityID,
	}

	c.JSON(http.StatusCreated, msg.SuccessResponse("商品创建成功", &data))
}

// CommodityUpdate 更新商品
func (cc *CommodityController) CommodityUpdate(c *gin.Context) {
	commodityIDStr := c.Param("id")
	commodityID, err := strconv.Atoi(commodityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的商品ID"))
		return
	}

	// 查询商品
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", commodityID).First(&commodity).Error; err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("商品不存在"))
		return
	}

	// 绑定请求数据
	var updateData models.Commodity
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的JSON格式", err))
		return
	}

	// 更新字段
	if updateData.Name != "" {
		commodity.Name = updateData.Name
	}
	if updateData.StyleCode != "" {
		commodity.StyleCode = updateData.StyleCode
	}
	if updateData.Category != "" {
		commodity.Category = updateData.Category
	}
	if updateData.Price > 0 {
		commodity.Price = updateData.Price
	}
	if updateData.Size != "" {
		commodity.Size = updateData.Size
	}
	if updateData.Color != "" {
		commodity.Color = updateData.Color
	}
	if updateData.CategoryDetail != "" {
		commodity.CategoryDetail = updateData.CategoryDetail
	}
	if updateData.Height != "" {
		commodity.Height = updateData.Height
	}
	if updateData.SpecCode != "" {
		commodity.SpecCode = updateData.SpecCode
	}
	if updateData.Notes != "" {
		commodity.Notes = updateData.Notes
	}

	// 保存更新
	if err := db.DB.Save(&commodity).Error; err != nil {
		log.Printf("更新商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("商品更新成功"))
}

// CommodityDelete 删除商品
func (cc *CommodityController) CommodityDelete(c *gin.Context) {
	commodityIDStr := c.Param("id")
	commodityID, err := strconv.Atoi(commodityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的商品ID"))
		return
	}

	// 查询商品
	var commodity models.Commodity
	if err := db.DB.Where("commodity_id = ?", strconv.Itoa(commodityID)).First(&commodity).Error; err != nil {
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("商品不存在"))
		return
	}

	// 删除商品（软删除）
	if err := db.DB.Delete(&commodity).Error; err != nil {
		log.Printf("删除商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("商品删除成功"))
}

// SearchProductsByName 根据名称搜索商品
func (cc *CommodityController) SearchProductsByName(c *gin.Context) {
	var requestData requestbody.SearchProductsByNameRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的请求数据", err))
		return
	}

	// 限制每页最大数量
	if requestData.PageSize > 100 {
		requestData.PageSize = 100
	}

	// 调用封装的函数搜索商品
	commodities, totalCount, err := method.SearchCommoditiesByName(requestData.SearchStr, requestData.Page, requestData.PageSize)
	if err != nil {
		log.Printf("搜索商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	// 转换为map数组
	result := method.ConvertCommoditiesToMap(commodities, c)

	// 计算总页数
	totalPages := (totalCount + int64(requestData.PageSize) - 1) / int64(requestData.PageSize)

	data := map[string]any{
		"items":     result,
		"total":     totalCount,
		"page":      requestData.Page,
		"page_size": requestData.PageSize,
		"pages":     totalPages,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

// BatchGetProductsByIDs 批量获取商品信息
func (cc *CommodityController) BatchGetProductsByIDs(c *gin.Context) {
	var requestData requestbody.BatchGetProductsByIDsRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("请求数据无效", err))
		return
	}

	log.Printf("接收到的commodity_ids参数: %v", requestData.CommodityIDs)

	// 调用封装的函数批量获取商品
	commodities, err := method.BatchGetCommodities(requestData.CommodityIDs)
	if err != nil {
		log.Printf("批量获取商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	// 转换为map数组
	result := method.ConvertCommoditiesToMap(commodities, c)

	data := map[string]any{
		"data":  result,
		"count": len(commodities),
	}

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &data))
}

// ChangeStyleCodeStatusOnline 设置款式代码为在线状态
func (cc *CommodityController) ChangeStyleCodeStatusOnline(c *gin.Context) {
	var requestData requestbody.ChangeStyleCodeStatusRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("款式代码不能为空", err))
		return
	}

	// 调用封装的函数更新款式代码状态
	if err := method.UpdateStyleCodeStatusOnline(requestData.StyleCode); err != nil {
		log.Printf("更新款式状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("款式代码状态更新成功"))
}

// ChangeStyleCodeStatusOffline 设置款式代码为离线状态
func (cc *CommodityController) ChangeStyleCodeStatusOffline(c *gin.Context) {
	var requestData requestbody.ChangeStyleCodeStatusRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("款式代码不能为空", err))
		return
	}

	// 获取当前时间
	currentTime := time.Now()

	// 查找款式状态
	var styleCodeSituation models.StyleCodeSituation
	if err := db.DB.Where("style_code = ?", requestData.StyleCode).First(&styleCodeSituation).Error; err != nil {
		log.Printf("款式不存在，style_code: %s, 错误: %v", requestData.StyleCode, err)
		c.JSON(http.StatusNotFound, msg.ErrResponse("款式不存在", err))
		return
	}

	// 更新状态为离线
	styleCodeSituation.Status = "offline"
	styleCodeSituation.OfflineTime = &currentTime
	if err := db.DB.Save(&styleCodeSituation).Error; err != nil {
		log.Printf("更新款式状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponse("服务器内部错误", err))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("款式代码状态更新成功"))
}

// GetCommoditiesByStyleCode 根据款式代码获取商品列表
func (cc *CommodityController) GetCommoditiesByStyleCode(c *gin.Context) {
	var requestData requestbody.Stylecode_commoditiesstruct

	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, msg.ErrResponse("参数错误", err))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名"))
		return
	}

	// 查询该style_code下的所有商品，库存大于0
	var commodities []models.Commodity
	if err := db.DB.Where("style_code = ? AND inventory > 0", requestData.StyleCode).Find(&commodities).Error; err != nil {
		log.Printf("根据款式代码获取商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器错误"))
		return
	}

	// 查询商品状态，只返回上线商品
	var onlineCommodityIDs []string
	if err := db.DB.Model(&models.CommoditySituation{}).
		Where("status = ?", "online").
		Where("commodity_id IN ?", method.ExtractCommodityIDs(commodities)).
		Pluck("commodity_id", &onlineCommodityIDs).Error; err != nil {
		log.Printf("获取上线商品ID失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器错误"))
		return
	}

	// 过滤出上线的商品
	var onlineCommodities []models.Commodity
	if err := db.DB.Where("commodity_id IN ?", onlineCommodityIDs).Find(&onlineCommodities).Error; err != nil {
		log.Printf("获取上线商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器处理失败",
		})
		return
	}

	// 获取请求的协议，考虑反向代理环境
	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	// 查询StyleCodeData表，获取款式的基本信息
	var styleCodeData models.StyleCodeData
	if err := db.DB.Where("style_code = ?", requestData.StyleCode).First(&styleCodeData).Error; err != nil {
		log.Printf("获取款式信息失败: %v", err)
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("款式不存在"))
		return
	}

	// 构建响应数据 - 基本信息从款式表获取
	result := map[string]interface{}{
		"name":             styleCodeData.Name,
		"price":            styleCodeData.Price,
		"inventory":        0,
		"items":            []map[string]interface{}{},
		"images":           []map[string]interface{}{},
		"main_image":       nil,
		"other_images":     []map[string]interface{}{},
		"display_pictures": make(map[string]string),
		"category":         styleCodeData.Category,
		"labels": map[string]string{
			"label_one":   styleCodeData.LabelOne,
			"label_two":   styleCodeData.LabelTwo,
			"label_three": styleCodeData.LabelThree,
			"label_four":  styleCodeData.LabelFour,
			"label_five":  styleCodeData.LabelFive,
			"label_six":   styleCodeData.LabelSix,
			"label_seven": styleCodeData.LabelSeven,
		},
	}

	// 计算款式总库存
	totalInventory := 0
	for _, commodity := range commodities {
		totalInventory += commodity.Inventory
	}
	result["inventory"] = totalInventory

	// 处理主图
	if styleCodeData.Image != "" {
		mainImageURL := utils.BuildFullImageURL(baseURL, styleCodeData.Image, "media")
		imageInfo := map[string]interface{}{
			"id":         nil,
			"url":        mainImageURL,
			"is_main":    true,
			"created_at": styleCodeData.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		result["images"] = []map[string]interface{}{imageInfo}
		result["main_image"] = imageInfo
	}

	// 处理展示图片
	if styleCodeData.DisplayPictures != "" && styleCodeData.DisplayPictures != "{}" {
		var displayPicturesMap map[string]string
		if err := json.Unmarshal([]byte(styleCodeData.DisplayPictures), &displayPicturesMap); err == nil {
			processedPictures := make(map[string]string)
			for key, imagePath := range displayPicturesMap {
				cleanPath := strings.ReplaceAll(imagePath, "\\\\", "\\")
				cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
				processedPictures[key] = utils.BuildFullImageURL(baseURL, cleanPath, "media")
			}
			result["display_pictures"] = processedPictures
		} else {
			log.Printf("解析display_pictures失败: %v", err)
		}
	}

	// 创建颜色分组的字典
	colorGroups := make(map[string]map[string]interface{})

	for _, commodity := range onlineCommodities {
		// 获取color_image
		var colorImage string
		if commodity.ColorImage != "" {
			colorImage = utils.BuildFullImageURL(baseURL, commodity.ColorImage, "media")
		} else if commodity.Image != "" {
			colorImage = utils.BuildFullImageURL(baseURL, commodity.Image, "media")
		} else {
			colorImage = ""
		}

		// 按颜色分组
		color := commodity.Color
		if _, exists := colorGroups[color]; !exists {
			colorGroups[color] = map[string]interface{}{
				"color":       color,
				"color_image": colorImage,
				"sizes":       []map[string]interface{}{},
			}
		}

		// 添加尺码信息到颜色组
		colorGroups[color]["sizes"] = append(colorGroups[color]["sizes"].([]map[string]interface{}), map[string]interface{}{
			"commodity_id": commodity.CommodityID,
			"size":         commodity.Size,
			"inventory":    commodity.Inventory,
		})
	}

	// 将颜色分组字典转换为列表格式
	items := make([]map[string]interface{}, 0, len(colorGroups))
	for _, colorInfo := range colorGroups {
		if colorInfo["color"].(string) != "尺码推荐" {
			items = append(items, colorInfo)
		}
	}
	result["items"] = items

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &result))
}

// handleDisplayPicture 处理单个展示图片的上传和URL构建
func handleDisplayPicture(c *gin.Context, file *multipart.FileHeader, styleCode string, baseURL string, position string, displayPictures map[string]string) error {
	// 为每个款式编码创建独立的子文件夹
	displayDir := filepath.Join("styles", styleCode, "display")
	log.Printf("准备保存到目录: %s", displayDir)
	// 保存上传的文件
	imagePath, err := utils.SaveUploadedFile(c, file, displayDir, "style_display_")
	if err != nil {
		log.Printf("SaveUploadedFile失败: %v", err)
		return fmt.Errorf("上传展示图片(位置%s)失败: %v", position, err)
	}
	log.Printf("SaveUploadedFile成功, 返回路径: %s", imagePath)
	// 实际保存文件到指定路径
	fullPath := utils.MediaPath(imagePath)
	log.Printf("准备保存文件到完整路径: %s", fullPath)
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		log.Printf("c.SaveUploadedFile失败: %v", err)
		return fmt.Errorf("保存展示图片(位置%s)失败: %v", position, err)
	}
	// 设置文件权限为 0644（读权限给所有人），确保 nginx 可以访问
	if err := os.Chmod(fullPath, 0644); err != nil {
		log.Printf("设置文件权限失败: %v", err)
	}
	// 确保目录也有正确的权限
	dirPath := filepath.Dir(fullPath)
	if err := os.Chmod(dirPath, 0755); err != nil {
		log.Printf("设置目录权限失败: %v", err)
	}
	log.Printf("文件保存成功: %s", fullPath)
	// 只保存相对路径到数据库
	log.Printf("保存的相对路径: %s", imagePath)
	// 使用指定位置作为key
	displayPictures[position] = imagePath
	log.Printf("添加到displayPictures映射, key: %s, value: %s", position, imagePath)
	return nil
}

// SearchStyleCode 搜索商品编码
func (cc *CommodityController) SearchStyleCode(c *gin.Context) {
	var req requestbody.StyleCodeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的JSON格式", err))
		return
	}

	// 验证店铺名称
	if req.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	} else if req.PageSize > 50 {
		req.PageSize = 50
	}

	styleCodes, total, err := method.SearchStyleCode(req.StyleCode, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("搜索商品编码失败: "+err.Error()))
		return
	}

	// 计算总页数
	totalPages := (int(total) + req.PageSize - 1) / req.PageSize

	// 返回与GoodsQuery方法一致的格式
	responseData := map[string]any{
		"data":      styleCodes,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"pages":     totalPages,
	}

	c.JSON(http.StatusOK, msg.SuccessResponse(fmt.Sprintf("查询成功，共%d件商品", total), &responseData))
}

// UpdateStyleCodeInfo 修改款式信息
func (cc *CommodityController) UpdateStyleCodeInfo(c *gin.Context) {
	styleCode := c.PostForm("style_code")
	if styleCode == "" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("款式编码不能为空"))
		return
	}

	name := c.PostForm("name")
	category := c.PostForm("category")
	priceStr := c.PostForm("price")
	labelOne := c.PostForm("label_one")
	labelTwo := c.PostForm("label_two")
	labelThree := c.PostForm("label_three")
	labelFour := c.PostForm("label_four")
	labelFive := c.PostForm("label_five")
	labelSix := c.PostForm("label_six")
	labelSeven := c.PostForm("label_seven")

	proto := utils.GetRequestProto(c)
	baseURL := fmt.Sprintf("%s://%s", proto, c.Request.Host)

	var mainImageURL string
	mainImageFile, err := c.FormFile("image")
	if err == nil && mainImageFile != nil {
		styleDir := filepath.Join("styles", styleCode)
		imagePath, err := utils.SaveUploadedFile(c, mainImageFile, styleDir, "style_main_")
		if err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("上传主图失败", err))
			return
		}
		fullPath := utils.MediaPath(imagePath)
		if err := c.SaveUploadedFile(mainImageFile, fullPath); err != nil {
			c.JSON(http.StatusInternalServerError, msg.ErrResponse("保存主图失败", err))
			return
		}
		// 设置文件权限为 0644（读权限给所有人），确保 nginx 可以访问
		if err := os.Chmod(fullPath, 0644); err != nil {
			log.Printf("设置文件权限失败: %v", err)
		}
		// 确保目录也有正确的权限
		dirPath := filepath.Dir(fullPath)
		if err := os.Chmod(dirPath, 0755); err != nil {
			log.Printf("设置目录权限失败: %v", err)
		}
		mainImageURL = imagePath
	}

	displayPictures := make(map[string]string)
	oldDisplayPictures := make(map[string]string)
	var tempStyleData models.StyleCodeData
	err = db.DB.Where("style_code = ?", styleCode).First(&tempStyleData).Error
	if err == nil {
		if tempStyleData.DisplayPictures != "" && tempStyleData.DisplayPictures != "{}" {
			if err := json.Unmarshal([]byte(tempStyleData.DisplayPictures), &displayPictures); err != nil {
				log.Printf("解析现有展示图片数据失败: %v", err)
				displayPictures = make(map[string]string)
			}
			oldDisplayPictures = displayPictures
		}
	}

	c.Request.ParseForm()
	displayPicturesJSON := c.PostForm("display_pictures_json")
	if displayPicturesJSON != "" {
		if err := json.Unmarshal([]byte(displayPicturesJSON), &displayPictures); err != nil {
			c.JSON(http.StatusBadRequest, msg.ErrResponse("无效的display_pictures_json格式", err))
			return
		}
	} else {
		for key := range c.Request.PostForm {
			if strings.HasPrefix(key, "display_pictures[") && strings.HasSuffix(key, "]") {
				position := key[17 : len(key)-1]
				if _, err := strconv.Atoi(position); err == nil {
					value := c.PostForm(key)
					displayPictures[position] = value
				}
			}
		}
	}

	for position, imagePath := range displayPictures {
		if imagePath == "" {
			if oldPath, exists := oldDisplayPictures[position]; exists && oldPath != "" {
				fullPath := utils.MediaPath(oldPath)
				if err := os.Remove(fullPath); err != nil {
					log.Printf("删除图片文件失败: %v", err)
				}
			}
			delete(displayPictures, position)
		}
	}

	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if files, ok := form.File["display_pictures[]"]; ok {
			for _, file := range files {
				if err := handleDisplayPicture(c, file, styleCode, baseURL, fmt.Sprintf("%d", len(displayPictures)+1), displayPictures); err != nil {
					c.JSON(http.StatusInternalServerError, msg.ErrResponse(err.Error(), err))
					return
				}
			}
		}

		for fieldName, files := range form.File {
			if len(files) > 0 {
				if strings.HasPrefix(fieldName, "display_pictures[") && strings.HasSuffix(fieldName, "]") {
					positionStr := fieldName[17 : len(fieldName)-1]
					if _, err := strconv.Atoi(positionStr); err == nil {
						if err := handleDisplayPicture(c, files[0], styleCode, baseURL, positionStr, displayPictures); err != nil {
							c.JSON(http.StatusInternalServerError, msg.ErrResponse(err.Error(), err))
							return
						}
					}
				}
			}
		}
	}

	begin := db.DB.Begin()
	if begin.Error != nil {
		begin.Rollback()
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	var price float64
	if priceStr != "" {
		price, _ = strconv.ParseFloat(priceStr, 64)
	}

	// 更新款式表（StyleCodeData）
	var styleCodeData models.StyleCodeData
	if err := begin.Where("style_code = ?", styleCode).First(&styleCodeData).Error; err != nil {
		begin.Rollback()
		c.JSON(http.StatusNotFound, msg.ErrResponseStr("款式编码不存在"))
		return
	}

	if name != "" {
		styleCodeData.Name = name
	}
	if price > 0 {
		styleCodeData.Price = price
	}
	if category != "" {
		styleCodeData.Category = category
	}
	if labelOne != "" {
		styleCodeData.LabelOne = labelOne
	}
	if labelTwo != "" {
		styleCodeData.LabelTwo = labelTwo
	}
	if labelThree != "" {
		styleCodeData.LabelThree = labelThree
	}
	if labelFour != "" {
		styleCodeData.LabelFour = labelFour
	}
	if labelFive != "" {
		styleCodeData.LabelFive = labelFive
	}
	if labelSix != "" {
		styleCodeData.LabelSix = labelSix
	}
	if labelSeven != "" {
		styleCodeData.LabelSeven = labelSeven
	}
	if mainImageURL != "" {
		styleCodeData.Image = mainImageURL
	}
	if len(displayPictures) > 0 {
		displayPicturesJSON, err = utils.StringMapToJSONString(displayPictures)
		if err != nil {
			begin.Rollback()
			log.Printf("转换展示图片数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新款式信息失败"))
			return
		}
		styleCodeData.DisplayPictures = displayPicturesJSON
	}

	if err := begin.Save(&styleCodeData).Error; err != nil {
		begin.Rollback()
		log.Printf("更新款式数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("更新款式信息失败"))
		return
	}

	if err := begin.Commit().Error; err != nil {
		begin.Rollback()
		log.Printf("提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}

	c.JSON(http.StatusOK, msg.SuccessResponseStr("更新成功"))
}

// GetAllCategories 获取所有商品类别
func (cc *CommodityController) GetAllCategories(c *gin.Context) {
	var requestData struct {
		Shopname   string   `json:"shopname" binding:"required"`
		LabelOne   []string `json:"label_one" binding:"omitempty"`
		LabelTwo   []string `json:"label_two" binding:"omitempty"`
		LabelThree []string `json:"label_three" binding:"omitempty"`
		LabelFour  []string `json:"label_four" binding:"omitempty"`
		LabelSeven []string `json:"label_seven" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("参数错误", err))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 构建查询
	query := db.DB.Model(&models.StyleCodeData{})
	query = query.Where("style_code IN (SELECT style_code FROM StyleCode_Situation WHERE status = ?)", "online")

	// 应用标签过滤条件
	if len(requestData.LabelOne) > 0 {
		query = query.Where("label_one IN (?)", requestData.LabelOne)
	}
	if len(requestData.LabelTwo) > 0 {
		query = query.Where("label_two IN (?)", requestData.LabelTwo)
	}
	if len(requestData.LabelThree) > 0 {
		query = query.Where("label_three IN (?)", requestData.LabelThree)
	}
	if len(requestData.LabelFour) > 0 {
		query = query.Where("label_four IN (?)", requestData.LabelFour)
	}
	if len(requestData.LabelSeven) > 0 {
		query = query.Where("label_seven IN (?)", requestData.LabelSeven)
	}

	// 查询所有不重复的商品类别，排除"其它"类别
	var categories []string
	if err := query.Distinct("category").
		Where("category != ?", "其它").
		Order("category").
		Pluck("category", &categories).Error; err != nil {
		log.Printf("查询商品类别失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	// 添加"全部"作为第一个默认选项
	categories = append([]string{"全部"}, categories...)
	info := map[string]any{
		"categories": categories,
	}
	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
}

// GetAllLabels 获取所有标签（LabelOne到LabelFour）
func (cc *CommodityController) GetAllLabels(c *gin.Context) {
	var requestData requestbody.GetAllLabelsRequestBody

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, msg.ErrResponse("参数错误", err))
		return
	}

	// 验证店铺名称
	if requestData.Shopname != "youlan_kids" {
		c.JSON(http.StatusBadRequest, msg.ErrResponseStr("无效的店铺名称"))
		return
	}

	// 构建响应数据
	info := make(map[string]any)

	// 查询所有标签
	var labelOneList []string
	var labelTwoList []string
	var labelThreeList []string
	var labelFourList []string
	var labelSevenList []string

	// 创建基础查询构建函数
	buildBaseQuery := func() *gorm.DB {
		query := db.DB.Model(&models.StyleCodeData{})
		query = query.Where("style_code IN (SELECT style_code FROM StyleCode_Situation WHERE status = ?)", "online")

		// 处理分类过滤
		if requestData.Category != "全部" && requestData.Category != nil {
			if categoryList, ok := requestData.Category.([]interface{}); ok {
				stringList := make([]string, 0, len(categoryList))
				for _, cat := range categoryList {
					if strCat, ok := cat.(string); ok {
						stringList = append(stringList, strCat)
					}
				}
				if len(stringList) > 0 {
					query = query.Where("category IN (?)", stringList)
				}
			} else if strCat, ok := requestData.Category.(string); ok {
				query = query.Where("category = ?", strCat)
			}
		}

		// 应用标签过滤条件
		if len(requestData.LabelOne) > 0 {
			query = query.Where("label_one IN (?)", requestData.LabelOne)
		}
		if len(requestData.LabelTwo) > 0 {
			query = query.Where("label_two IN (?)", requestData.LabelTwo)
		}
		if len(requestData.LabelThree) > 0 {
			query = query.Where("label_three IN (?)", requestData.LabelThree)
		}
		if len(requestData.LabelFour) > 0 {
			query = query.Where("label_four IN (?)", requestData.LabelFour)
		}
		if len(requestData.LabelSeven) > 0 {
			query = query.Where("label_seven IN (?)", requestData.LabelSeven)
		}
		return query
	}

	// 查询LabelOne
	if err := buildBaseQuery().Distinct("label_one").
		Where("label_one != ?", "").
		Order("label_one").
		Pluck("label_one", &labelOneList).Error; err != nil {
		log.Printf("查询LabelOne失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	info["label_one"] = labelOneList

	// 查询LabelTwo
	if err := buildBaseQuery().Distinct("label_two").
		Where("label_two != ?", "").
		Order("label_two").
		Pluck("label_two", &labelTwoList).Error; err != nil {
		log.Printf("查询LabelTwo失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	info["label_two"] = labelTwoList

	// 查询LabelThree
	if err := buildBaseQuery().Distinct("label_three").
		Where("label_three != ?", "").
		Order("label_three").
		Pluck("label_three", &labelThreeList).Error; err != nil {
		log.Printf("查询LabelThree失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	info["label_three"] = labelThreeList

	// 查询LabelFour
	if err := buildBaseQuery().Distinct("label_four").
		Where("label_four != ?", "").
		Order("label_four").
		Pluck("label_four", &labelFourList).Error; err != nil {
		log.Printf("查询LabelFour失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	info["label_four"] = labelFourList

	// 查询LabelSeven
	if err := buildBaseQuery().Distinct("label_seven").
		Where("label_seven != ?", "").
		Order("label_seven").
		Pluck("label_seven", &labelSevenList).Error; err != nil {
		log.Printf("查询LabelSeven失败: %v", err)
		c.JSON(http.StatusInternalServerError, msg.ErrResponseStr("服务器内部错误"))
		return
	}
	info["label_seven"] = labelSevenList

	c.JSON(http.StatusOK, msg.SuccessResponse("查询成功", &info))
}
