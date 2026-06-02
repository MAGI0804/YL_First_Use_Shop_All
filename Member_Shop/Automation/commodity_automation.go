package Automation

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"Member_shop/db"
	"Member_shop/models"
)

var (
	AppKey    = os.Getenv("JST_APP_KEY_PROD")
	AppSelect = os.Getenv("JST_APP_SECRET_PROD")
	ShopID    = getEnv("JST_SHOP_ID", "")
	Code      = os.Getenv("JST_AUTH_CODE_PROD")
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type ImageCache struct {
	Path        string
	Name        string
	Hash        string
	Downloading bool
	WaitGroup   sync.WaitGroup
}

var (
	// 使用全局缓存，只以图片URL为键，确保一个图片只下载一次
	globalImageCache sync.Map
)

type TokenResponse struct {
	Code int `json:"code"`
	Data struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type InventoryResponse struct {
	Code int `json:"code"`
	Data struct {
		HasNext bool       `json:"has_next"`
		Datas   []ItemData `json:"datas"`
	} `json:"data"`
}

type ItemData struct {
	SKUID           string  `json:"sku_id"`
	RawSKUID        string  `json:"raw_sku_id"`
	IID             string  `json:"i_id"`
	Name            string  `json:"name"`
	Pic             string  `json:"pic"`
	PropertiesValue string  `json:"properties_value"`
	ShopQty         int     `json:"shop_qty"`
	Other2          string  `json:"other_2"`
	Other3          string  `json:"other_3"`
	Other4          string  `json:"other_4"`
	Other5          string  `json:"other_5"`
	Other6          string  `json:"other_6"`
	Other7          string  `json:"other_7"`
	Other8          string  `json:"other_8"`
	Other9          string  `json:"other_9"`
	VCName          string  `json:"vc_name"`
	SalePrice       float64 `json:"sale_price"`
}

func MD5Encrypt(input string) string {
	hash := md5.Sum([]byte(input))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

func GetToken() (string, error) {
	timestamp := time.Now().Unix()
	charset := "uft-8" // 注意：Python代码中使用了错误的拼写，API可能期望这个值
	grantType := "authorization_code"

	// 按照Python代码的格式构建签名字符串
	convertedStr := fmt.Sprintf("%sapp_key%scharset%scode%sgrant_type%stimestamp%d",
		AppSelect, AppKey, charset, Code, grantType, timestamp)

	sign := MD5Encrypt(convertedStr)

	apiURL := "https://openapi.jushuitan.com/openWeb/auth/getInitToken"

	data := url.Values{}
	data.Set("app_key", AppKey)
	data.Set("grant_type", grantType)
	data.Set("timestamp", fmt.Sprintf("%d", timestamp))
	data.Set("code", Code)
	data.Set("charset", charset)
	data.Set("sign", sign)

	client := &http.Client{Timeout: 30 * time.Second}

	// 显式设置Content-Type头，与Python代码保持一致
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	fmt.Printf("tokenResp: %+v\n", tokenResp)

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("API返回错误码: %d", tokenResp.Code)
	}

	return tokenResp.Data.AccessToken, nil
}

func SendInventoryQuery(appKey, accessToken, timestamp, charset string, version int, sign, biz string) (*InventoryResponse, error) {
	apiURL := "https://openapi.jushuitan.com/open/skumap/query"

	data := url.Values{}
	data.Set("app_key", appKey)
	data.Set("access_token", accessToken)
	data.Set("timestamp", timestamp)
	data.Set("charset", charset)
	data.Set("version", fmt.Sprintf("%d", version))
	data.Set("sign", sign)
	data.Set("biz", biz)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(apiURL, data)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var invResp InventoryResponse
	if err := json.Unmarshal(body, &invResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &invResp, nil
}

func CalculateContentHash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GenerateSafeDirName(styleCode string) string {
	var result strings.Builder
	for _, c := range styleCode {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result.WriteRune(c)
		}
	}
	safeName := result.String()
	// 如果过滤后为空，使用默认值
	if safeName == "" {
		safeName = "default"
	}
	return safeName
}

func FindExistingImageByContent(imageContent []byte, styleCode string) string {
	contentHash := CalculateContentHash(imageContent)

	safeStyleCode := GenerateSafeDirName(styleCode)
	targetDir := filepath.Join("media", "commodities", safeStyleCode)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(targetDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		if baseName == contentHash {
			return filepath.Join("commodities", safeStyleCode, entry.Name())
		}

		fileHash, err := CalculateFileHash(filePath)
		if err != nil {
			continue
		}

		if fileHash == contentHash {
			return filepath.Join("commodities", safeStyleCode, entry.Name())
		}
	}

	return ""
}

func DownloadImage(imageURL, iID string) (string, string, error) {
	fmt.Printf("=== 进入DownloadImage函数 ===\n")
	fmt.Printf("参数: imageURL=%s, iID=%s\n", imageURL, iID)

	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		fmt.Printf("无效的图片URL: %s\n", imageURL)
		return "", "", fmt.Errorf("无效的图片URL: %s", imageURL)
	}

	// 检查全局缓存，使用imageURL作为唯一键，确保一个图片只下载一次
	fmt.Printf("检查URL缓存: %s\n", imageURL)
	if cached, ok := globalImageCache.Load(imageURL); ok {
		cache := cached.(*ImageCache)
		fmt.Printf("URL缓存存在，状态: %+v\n", cache)

		// 如果图片正在下载中，等待下载完成
		if cache.Downloading {
			fmt.Printf("图片正在下载中，等待完成...\n")
			cache.WaitGroup.Wait()
			fmt.Printf("图片下载完成，返回缓存路径\n")
			return cache.Path, cache.Name, nil
		}

		// 检查文件是否存在
		if _, err := os.Stat(cache.Path); err == nil {
			fmt.Printf("缓存文件存在，直接返回: %s\n", cache.Path)
			return cache.Path, cache.Name, nil
		}
		fmt.Printf("缓存文件不存在，重新下载\n")
	}

	// 创建新的缓存对象，标记为正在下载
	cache := &ImageCache{
		Downloading: true,
	}
	cache.WaitGroup.Add(1)

	// 尝试存储缓存对象，如果已存在则使用已有的
	if existing, loaded := globalImageCache.LoadOrStore(imageURL, cache); loaded {
		fmt.Printf("其他协程已创建缓存，使用现有缓存\n")
		existingCache := existing.(*ImageCache)
		// 如果图片正在下载中，等待下载完成
		if existingCache.Downloading {
			fmt.Printf("图片正在下载中，等待完成...\n")
			existingCache.WaitGroup.Wait()
			fmt.Printf("图片下载完成，返回缓存路径\n")
			return existingCache.Path, existingCache.Name, nil
		}
		// 如果文件存在，直接返回
		if _, err := os.Stat(existingCache.Path); err == nil {
			fmt.Printf("缓存文件存在，直接返回: %s\n", existingCache.Path)
			return existingCache.Path, existingCache.Name, nil
		}
		// 更新existingCache为正在下载状态
		existingCache.Downloading = true
		existingCache.WaitGroup.Add(1)
		cache = existingCache
	}

	fmt.Printf("开始下载图片: %s\n", imageURL)

	// 添加请求超时
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		fmt.Printf("图片下载失败: %v\n", err)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("图片下载HTTP错误: %d\n", resp.StatusCode)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 获取Content-Type，检查是否为图片
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("图片Content-Type: %s\n", contentType)
	if !strings.HasPrefix(contentType, "image/") {
		fmt.Printf("非图片类型，跳过下载: %s\n", contentType)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("非图片类型: %s", contentType)
	}

	imageContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取图片失败: %v\n", err)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("读取图片失败: %v", err)
	}

	fmt.Printf("图片大小: %d 字节\n", len(imageContent))

	imageHash := CalculateContentHash(imageContent)
	fmt.Printf("图片Hash: %s\n", imageHash)

	// 生成安全的文件名
	var finalFilename string
	var ext string

	// 根据content-type获取正确的扩展名
	fmt.Printf("根据Content-Type确定扩展名\n")
	switch contentType {
	case "image/png":
		ext = ".png"
		fmt.Printf("使用PNG扩展名\n")
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
		fmt.Printf("使用JPG扩展名\n")
	case "image/gif":
		ext = ".gif"
		fmt.Printf("使用GIF扩展名\n")
	default:
		ext = ".png" // 默认使用png
		fmt.Printf("使用默认PNG扩展名\n")
	}

	// 生成最终文件名
	fmt.Printf("生成最终文件名\n")
	if iID != "" {
		safeIID := GenerateSafeDirName(iID)
		fmt.Printf("使用iID生成文件名，iID: %s, safeIID: %s\n", iID, safeIID)
		finalFilename = fmt.Sprintf("%s%s", safeIID, ext)
	} else {
		finalFilename = fmt.Sprintf("%s%s", imageHash, ext)
		fmt.Printf("使用Hash生成文件名: %s\n", finalFilename)
	}
	fmt.Printf("最终文件名: %s\n", finalFilename)

	// 获取当前文件所在目录，构建绝对路径
	fmt.Printf("获取当前文件目录\n")
	_, currentGoFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Printf("获取当前文件目录失败\n")
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("获取当前文件目录失败")
	}
	fmt.Printf("当前Go文件: %s\n", currentGoFile)

	// 当前文件在Automation目录，需要向上一级到项目根目录
	currentFileDir := filepath.Dir(currentGoFile)
	fmt.Printf("当前文件目录: %s\n", currentFileDir)
	projectRoot := filepath.Join(currentFileDir, "..")
	fmt.Printf("项目根目录: %s\n", projectRoot)

	// 构建最终的保存目录
	safeStyleCode := GenerateSafeDirName(iID)
	fmt.Printf("安全目录名: %s\n", safeStyleCode)
	finalDir := filepath.Join(projectRoot, "media", "commodities", safeStyleCode)
	fmt.Printf("最终保存目录: %s\n", finalDir)

	// 确保目录存在
	fmt.Printf("创建目录: %s\n", finalDir)
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("创建目录失败: %v", err)
	}
	fmt.Printf("目录创建成功\n")

	// 构建最终的文件路径
	finalPath := filepath.Join(finalDir, finalFilename)
	fmt.Printf("保存图片到: %s\n", finalPath)

	// 保存图片到本地
	fmt.Printf("开始保存图片\n")
	if err := os.WriteFile(finalPath, imageContent, 0644); err != nil {
		fmt.Printf("保存图片失败: %v\n", err)
		// 标记下载完成，释放等待的协程
		cache.Downloading = false
		cache.WaitGroup.Done()
		return "", "", fmt.Errorf("保存图片失败: %v", err)
	}
	fmt.Printf("图片保存成功: %s\n", finalPath)

	// 更新缓存对象
	cache.Path = finalPath
	cache.Name = finalFilename
	cache.Hash = imageHash
	cache.Downloading = false
	fmt.Printf("更新缓存对象: %+v\n", cache)

	// 标记下载完成，释放等待的协程
	cache.WaitGroup.Done()

	fmt.Printf("=== 退出DownloadImage函数 ===\n")
	return finalPath, finalFilename, nil
}

type DateRange struct {
	Start string
	End   string
}

func SplitDateRange(startDateStr, endDateStr string, daysPerChunk int) []DateRange {
	layout := "2006-01-02 15:04:05"
	startDate, _ := time.Parse(layout, startDateStr)
	endDate, _ := time.Parse(layout, endDateStr)

	var dateRanges []DateRange
	currentStart := startDate

	for currentStart.Before(endDate) || currentStart.Equal(endDate) {
		currentEnd := currentStart.AddDate(0, 0, daysPerChunk-1)
		if currentEnd.After(endDate) {
			currentEnd = endDate
		}

		dateRanges = append(dateRanges, DateRange{
			Start: currentStart.Format(layout),
			End:   currentEnd.Format(layout),
		})

		currentStart = currentEnd.Add(time.Second)
	}

	return dateRanges
}

func ImportDataForDateRange(accessToken, modifiedBegin, modifiedEnd string) (int, error) {
	pageIndex := 1
	hasNext := true
	importedCount := 0
	charset := "UTF-8"
	version := 2

	processedStyleCodes := make(map[string]bool)

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for hasNext {
		<-ticker.C

		fmt.Printf("正在查询日期范围 [%s 至 %s] 的第 %d 页数据...\n", modifiedBegin, modifiedEnd, pageIndex)

		timestamp := fmt.Sprintf("%d", time.Now().Unix())

		biz := fmt.Sprintf(`{"page_index":"%d","page_size":"100","modified_begin":"%s","modified_end":"%s","shop_id":"%s"}`,
			pageIndex, modifiedBegin, modifiedEnd, ShopID)

		convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
			AppSelect, accessToken, AppKey, biz, charset, timestamp, version)
		sign := MD5Encrypt(convertedStr)

		response, err := SendInventoryQuery(AppKey, accessToken, timestamp, charset, version, sign, biz)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			break
		}

		if response.Code != 0 {
			fmt.Printf("API返回错误: %d\n", response.Code)
			break
		}

		if response.Data.Datas == nil || len(response.Data.Datas) == 0 {
			fmt.Printf("第 %d 页没有数据\n", pageIndex)
			break
		}

		fmt.Printf("正在导入第 %d 页的 %d 条数据...\n", pageIndex, len(response.Data.Datas))

		for _, item := range response.Data.Datas {
			commodityID := item.RawSKUID
			if commodityID == "" {
				fmt.Printf("跳过无commodity_id的商品\n")
				continue
			}

			propertiesValue := item.PropertiesValue
			var color, height string
			if propertiesValue != "" {
				props := strings.Split(propertiesValue, ",")
				for i, prop := range props {
					props[i] = strings.TrimSpace(prop)
				}
				if len(props) >= 1 {
					color = props[0]
				}
				if len(props) >= 2 {
					height = props[1]
				}
			}

			safeStyleCode := GenerateSafeDirName(item.IID)

			// 处理图片
			imagePath := ""
			imageName := ""
			itemImage := ""
			if item.Pic != "" && item.IID != "" {
				path, name, err := DownloadImage(item.Pic, item.IID)
				if err != nil {
					fmt.Printf("下载图片失败: %v\n", err)
				} else {
					imagePath = path
					imageName = name
				}

				// 如果有图片路径，设置图片路径
				if imagePath != "" {
					itemImage = filepath.Join("commodities", safeStyleCode, imageName)
				}
			}

			// 创建商品对象 - 初始导入时将Category设为空，以便后续批处理更新
			commodity := models.Commodity{
				CommodityID:    commodityID,
				Name:           item.Name,
				StyleCode:      safeStyleCode,
				Category:       "",
				CategoryDetail: "",
				Price:          0.0,
				Color:          color,
				Height:         height,
				Size:           height,
				Inventory:      item.ShopQty,
				CreatedAt:      time.Now(),
				Image:          itemImage,
			}

			fmt.Printf("准备保存商品，ID类型: string, ID值: %s\n", commodityID)

			// 检查商品是否已存在，只创建新记录
			var existingCommodity models.Commodity
			if err := db.DB.Where("commodity_id = ?", commodityID).First(&existingCommodity).Error; err != nil {
				// 商品不存在，创建新记录
				if err := db.DB.Create(&commodity).Error; err != nil {
					fmt.Printf("创建商品失败: %v\n", err)
					continue
				}
				fmt.Printf("创建商品成功: %s\n", commodityID)
			} else {
				fmt.Printf("商品已存在，跳过创建: %s\n", commodityID)
			}

			// 检查CommoditySituation是否已存在，只创建新记录
			var existingCommoditySituation models.CommoditySituation
			if err := db.DB.Where("commodity_id = ?", commodityID).First(&existingCommoditySituation).Error; err != nil {
				// CommoditySituation不存在，创建新记录
				commoditySituation := models.CommoditySituation{
					CommodityID: commodityID,
					Status:      "online", // 默认状态为online
					StyleCode:   safeStyleCode,
					Category:    commodity.Category,
					OnlineTime:  time.Now(),
				}

				if err := db.DB.Create(&commoditySituation).Error; err != nil {
					fmt.Printf("创建CommoditySituation失败: %v\n", err)
				} else {
					fmt.Printf("创建CommoditySituation成功，commodity_id: %s\n", commodityID)
				}
			} else {
				fmt.Printf("CommoditySituation已存在，跳过创建: %s\n", commodityID)
			}

			// 处理StyleCodeData和StyleCodeSituation（仅当未处理过时）
			if !processedStyleCodes[safeStyleCode] {
				processedStyleCodes[safeStyleCode] = true

				// 检查StyleCodeData是否已存在，只创建新记录
				var existingStyleCodeData models.StyleCodeData
				if err := db.DB.Where("style_code = ?", safeStyleCode).First(&existingStyleCodeData).Error; err != nil {
					// StyleCodeData不存在，创建新记录
					styleCodeData := models.StyleCodeData{
						StyleCode:       safeStyleCode,
						Name:            item.Name,
						Image:           itemImage,
						Category:        "其它",
						CategoryDetail:  "",
						Price:           0.0,
						CreatedAt:       time.Now(),
						DisplayPictures: "[]", // 空JSON数组作为默认值
					}

					if err := db.DB.Create(&styleCodeData).Error; err != nil {
						fmt.Printf("创建StyleCodeData失败: %v\n", err)
					} else {
						fmt.Printf("创建StyleCodeData成功，style_code: %s\n", safeStyleCode)
					}
				} else {
					fmt.Printf("StyleCodeData已存在，跳过创建: %s\n", safeStyleCode)
				}

				// 检查StyleCodeSituation是否已存在，只创建新记录
				var existingStyleCodeSituation models.StyleCodeSituation
				if err := db.DB.Where("style_code = ?", safeStyleCode).First(&existingStyleCodeSituation).Error; err != nil {
					// StyleCodeSituation不存在，创建新记录
					styleCodeSituation := models.StyleCodeSituation{
						StyleCode:   safeStyleCode,
						Status:      "online", // 默认状态为pending
						OfflineTime: nil,      // 使用nil表示NULL值
					}

					if err := db.DB.Create(&styleCodeSituation).Error; err != nil {
						fmt.Printf("创建StyleCodeSituation失败: %v\n", err)
					} else {
						fmt.Printf("创建StyleCodeSituation成功，style_code: %s\n", safeStyleCode)
					}
				}
			}

			fmt.Printf("商品 %s 保存成功\n", commodityID)
			importedCount++
		}

		hasNext = response.Data.HasNext
		pageIndex++

		time.Sleep(1 * time.Second)
	}

	return importedCount, nil
}

func ImportCommodityData(modifiedBegin string, modifiedEnd string) error {
	accessToken, err := GetToken()

	if err != nil {
		fmt.Printf("无法获取访问令牌: %v\n", err)
		return err
	}

	//modifiedBegin := "2025-11-20 00:00:00"
	//modifiedEnd := "2025-12-01 23:59:59"

	dateRanges := SplitDateRange(modifiedBegin, modifiedEnd, 7)
	fmt.Printf("日期范围已分割为 %d 个子范围\n", len(dateRanges))

	totalImported := 0

	for i, dateRange := range dateRanges {
		fmt.Printf("正在处理第 %d/%d 个子范围: %s 至 %s\n", i+1, len(dateRanges), dateRange.Start, dateRange.End)

		importedCount, err := ImportDataForDateRange(accessToken, dateRange.Start, dateRange.End)
		if err != nil {
			fmt.Printf("导入数据失败: %v\n", err)
			continue
		}

		totalImported += importedCount
		fmt.Printf("子范围 %d 导入完成，导入了 %d 条数据\n", i+1, importedCount)
	}

	fmt.Printf("数据导入完成，共导入 %d 条商品数据\n", totalImported)

	// 清理全局缓存，保留图片文件
	globalImageCache = sync.Map{}

	return nil
}

// modifiedBegin := "2025-11-20 00:00:00"
// modifiedEnd := "2025-12-01 23:59:59"
func StartScheduledSync(interval time.Duration, modifiedBegin string, modifiedEnd string) error {
	fmt.Printf("启动定时同步任务，间隔: %v\n", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	if err := ImportCommodityData(modifiedBegin, modifiedEnd); err != nil {
		fmt.Printf("同步任务执行失败: %v\n", err)
	}
	fmt.Printf("同步任务执行完成")
	if err := BatchProcessCommodities(); err != nil {
		fmt.Printf("批处理任务执行失败: %v\n", err)
	}
	fmt.Printf("批处理任务执行完成")
	return nil

	//for {
	//	select {
	//	case <-ticker.C:
	//		fmt.Println("\n=== 开始执行定时同步任务 ===")
	//		startTime := time.Now()
	//
	//
	//
	//		elapsed := time.Since(startTime)
	//		fmt.Printf("同步任务执行完成，耗时: %v\n", elapsed)
	//		fmt.Println("=================================\n")
	//	}
	//}
}

// --------------------------- 雷姆功能实现 ---------------------------

// 批次处理商品，每20个ID一组
func BatchProcessCommodities() error {
	fmt.Println("=== 开始批次处理商品 ===")

	// 获取所有商品ID
	commodityIDs, err := GetCommodityIDs()
	if err != nil {
		return fmt.Errorf("获取商品ID失败: %v", err)
	}

	if len(commodityIDs) == 0 {
		fmt.Println("没有找到商品ID")
		return nil
	}

	fmt.Printf("总共找到 %d 个商品ID\n", len(commodityIDs))

	// 分批处理，每20个ID一组
	batchSize := 20
	for i := 0; i < len(commodityIDs); i += batchSize {
		// 计算当前批次的结束索引
		end := i + batchSize
		if end > len(commodityIDs) {
			end = len(commodityIDs)
		}

		// 获取当前批次的ID
		batchIDs := commodityIDs[i:end]
		skuIDs := strings.Join(batchIDs, ",")

		fmt.Printf("处理批次 %d，包含 %d 个商品ID\n", i/batchSize+1, len(batchIDs))

		// 调用API查询这批商品
		result := ProcessCommodityBatch(skuIDs)

		// 如果处理失败，打印错误信息
		if !result {
			fmt.Printf("批次 %d 处理失败\n", i/batchSize+1)
		}

		// 避免请求过于频繁
		time.Sleep(1 * time.Second)
	}

	return nil
}

// 获取所有商品ID
func GetCommodityIDs() ([]string, error) {
	var commodityIDs []string

	// 查询所有商品的commodity_id
	if err := db.DB.Model(&models.Commodity{}).Pluck("commodity_id", &commodityIDs).Error; err != nil {
		return nil, fmt.Errorf("查询商品ID失败: %v", err)
	}

	return commodityIDs, nil
}

// 处理商品批次，调用API并更新数据库
func ProcessCommodityBatch(skuIDs string) bool {
	// 获取访问令牌
	accessToken, err := GetToken()
	if err != nil {
		fmt.Printf("获取令牌失败: %v\n", err)
		return false
	}

	// 保存当前批次的所有商品ID列表
	batchIDList := strings.Split(skuIDs, ",")

	// 构建请求参数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	charset := "UTF-8"
	version := 2

	biz := fmt.Sprintf(`{"page_index":"1","page_size":"100","sku_ids":"%s"}`, skuIDs)

	// 生成签名
	convertedStr := fmt.Sprintf("%saccess_token%sapp_key%sbiz%scharset%stimestamp%sversion%d",
		AppSelect, accessToken, AppKey, biz, charset, timestamp, version)
	sign := MD5Encrypt(convertedStr)

	// 发送请求 - 使用新的API端点
	apiURL := "https://openapi.jushuitan.com/open/sku/query"

	formData := url.Values{}
	formData.Set("app_key", AppKey)
	formData.Set("access_token", accessToken)
	formData.Set("timestamp", timestamp)
	formData.Set("charset", charset)
	formData.Set("version", fmt.Sprintf("%d", version))
	formData.Set("sign", sign)
	formData.Set("biz", biz)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		// API失败时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		// API失败时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		// API失败时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}

	// 解析响应
	var response InventoryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("解析响应失败: %v, 响应内容: %s\n", err, string(body))
		// API失败时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}

	if response.Code != 0 {
		fmt.Printf("API请求失败或返回异常: %+v\n", response)
		// API失败时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}

	// 处理返回的数据
	responseData := response.Data
	datas := responseData.Datas

	// 检查数据是否为空
	if len(datas) == 0 {
		fmt.Println("API返回数据为空")
		// 数据为空时，删除当前批次的所有商品
		DeleteCommoditiesNotFound(batchIDList, []string{})
		return false
	}

	// 获取在API响应中找到的商品ID
	var foundIDs []string
	for _, item := range datas {
		if item.SKUID != "" {
			foundIDs = append(foundIDs, item.SKUID)
		}
	}

	// 删除在API响应中找不到的商品
	DeleteCommoditiesNotFound(batchIDList, foundIDs)

	// 如果有找到的数据，更新数据库
	if len(datas) > 0 {
		UpdateCommodityData(datas)
	}

	return true
}

// 删除在API响应中找不到的商品记录
func DeleteCommoditiesNotFound(batchIDList []string, foundIDs []string) {
	// 找出在批次中但不在API响应中的商品ID
	var notFoundIDs []string
	for _, id := range batchIDList {
		found := false
		for _, foundID := range foundIDs {
			if id == foundID {
				found = true
				break
			}
		}
		if !found {
			notFoundIDs = append(notFoundIDs, id)
		}
	}

	if len(notFoundIDs) == 0 {
		return
	}

	// 删除这些商品记录
	result := db.DB.Where("commodity_id IN ?", notFoundIDs).Delete(&models.Commodity{})
	if result.Error != nil {
		fmt.Printf("删除商品失败: %v\n", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		fmt.Printf("删除了 %d 个在API中找不到数据的商品记录\n", result.RowsAffected)
	} else {
		fmt.Printf("没有需要删除的商品记录\n")
	}
}

// 根据API返回的数据更新商品和款式模型
func UpdateCommodityData(datas []ItemData) {
	// 已处理的style_code集合，用于去重
	processedStyleCodes := make(map[string]bool)

	for _, item := range datas {
		// 获取商品ID
		skuID := item.SKUID
		if skuID == "" {
			continue
		}

		// 根据用户需求，检查必要字段是否存在
		category := item.Other6
		categoryDetail := item.VCName
		price := item.SalePrice

		// 检查必要字段是否存在，如果不存在则跳过
		if category == "" || categoryDetail == "" || price == 0 {
			fmt.Printf("必要字段缺失，跳过商品记录: %s\n", skuID)
			continue
		}

		// 尝试获取商品记录
		var commodity models.Commodity
		if err := db.DB.Where("commodity_id = ?", skuID).First(&commodity).Error; err != nil {
			fmt.Printf("商品不存在: %s\n", skuID)
			continue
		}

		// 总是更新商品信息，确保数据最新
		commodity.Category = category
		commodity.CategoryDetail = categoryDetail
		commodity.Price = price

		if err := db.DB.Save(&commodity).Error; err != nil {
			fmt.Printf("更新商品失败: %s, 错误: %v\n", skuID, err)
			continue
		}

		// 更新CommoditySituation的Category和标签字段
		// 标签映射：标签1→other_2, 标签2→other_3, 标签3→other_4, 标签4→other_5, 标签5→other_7, 标签6→other_8, 标签7→other_9
		if err := db.DB.Model(&models.CommoditySituation{}).
			Where("commodity_id = ?", skuID).
			Updates(map[string]interface{}{
				"category":    category,
				"label_one":   item.Other2,
				"label_two":   item.Other3,
				"label_three": item.Other4,
				"label_four":  item.Other5,
				"label_five":  item.Other7,
				"label_six":   item.Other8,
				"label_seven": item.Other9,
			}).Error; err != nil {
			fmt.Printf("更新CommoditySituation失败: %v\n", err)
		} else {
			fmt.Printf("更新CommoditySituation成功: %s, category: %s\n", skuID, category)
		}

		fmt.Printf("更新商品成功: %s, category: %s, category_detail: %s, price: %v\n", skuID, category, categoryDetail, price)

		// 处理StyleCodeData和StyleCodeSituation（仅当未处理过时）
		safeStyleCode := GenerateSafeDirName(item.IID)
		if !processedStyleCodes[safeStyleCode] {
			processedStyleCodes[safeStyleCode] = true

			// 更新StyleCodeData（只更新现有记录，不创建）
			var existingStyleCodeData models.StyleCodeData
			if err := db.DB.Where("style_code = ?", safeStyleCode).First(&existingStyleCodeData).Error; err == nil {
				// 总是更新StyleCodeData，确保数据最新
				// 构建更新数据，包括标签字段
				// 标签映射：标签1→other_2, 标签2→other_3, 标签3→other_4, 标签4→other_5, 标签5→other_7, 标签6→other_8, 标签7→other_9
				updateData := map[string]interface{}{
					//"name":             item.Name,
					"image":            commodity.Image,
					"category":         category,
					"category_detail":  categoryDetail,
					"price":            price,
					"display_pictures": "[]",
					"label_one":        item.Other2,
					"label_two":        item.Other3,
					"label_three":      item.Other4,
					"label_four":       item.Other5,
					"label_five":       item.Other7,
					"label_six":        item.Other8,
					"label_seven":      item.Other9,
				}

				if err := db.DB.Model(&existingStyleCodeData).Updates(updateData).Error; err != nil {
					fmt.Printf("更新StyleCodeData失败: %v\n", err)
				} else {
					fmt.Printf("更新StyleCodeData成功，style_code: %s\n", safeStyleCode)
				}
			} else {
				fmt.Printf("StyleCodeData不存在，跳过更新: %s\n", safeStyleCode)
			}

			// 更新StyleCodeSituation（只更新现有记录，不创建）
			var existingStyleCodeSituation models.StyleCodeSituation
			if err := db.DB.Where("style_code = ?", safeStyleCode).First(&existingStyleCodeSituation).Error; err == nil {
				// 只在状态为空时更新
				if existingStyleCodeSituation.Status == "" {
					if err := db.DB.Model(&existingStyleCodeSituation).Updates(map[string]interface{}{
						"status": "pending",
					}).Error; err != nil {
						fmt.Printf("更新StyleCodeSituation失败: %v\n", err)
					} else {
						fmt.Printf("更新StyleCodeSituation成功，style_code: %s\n", safeStyleCode)
					}
				}
			} else {
				fmt.Printf("StyleCodeSituation不存在，跳过更新: %s\n", safeStyleCode)
			}
		}
	}
}

func GenerateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func init() {
	tempDir := filepath.Join("media", "temp")
	os.MkdirAll(tempDir, 0755)
}

//func main() {
//	StartScheduledSync(60 * time.Second)
//}
