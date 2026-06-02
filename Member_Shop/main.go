package main

import (
	"Member_shop/config"
	"Member_shop/db"
	"Member_shop/middleware"
	"Member_shop/routes"
	"Member_shop/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	appConfig := config.LoadConfig()
	//初始化sql数据库
	db.InitDB(appConfig)
	//初始化redis
	if err := db.InitRedis(appConfig); err != nil {
		log.Printf("⚠️ Redis初始化失败: %v，程序将继续运行但部分功能可能不可用", err)
	}
	// 同步数据库结构
	db.RunMigrations()

	//创建gin引擎
	router := gin.Default()

	//启用中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestLogMiddleware())
	//时间中间件
	router.Use(middleware.FormatTimeMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.AccessTokenValidationMiddleware())
	//订单消息中间件
	router.Use(middleware.OrderMessageMiddleware())

	// 设置静态文件服务
	if err := utils.EnsureMediaRoot(); err != nil {
		log.Fatalf("Failed to create media directory: %v", err)
	}
	router.Static("/static", "./staticfiles")
	router.Static("/media", utils.MediaRoot())

	//启用路由
	routes.InitRoutes(router)

	// 启动库存同步定时任务
	//go Automation.StartInventorySync()

	//设置商品定时同步
	//modifiedBegin := "2025-08-08 00:00:00"
	//modifiedEnd := "2025-12-15 23:59:59"
	//go Automation.StartScheduledSync(30*time.Second, modifiedBegin, modifiedEnd)

	// 启动服务器
	port := appConfig.ServerConfig.Port
	log.Printf("Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
