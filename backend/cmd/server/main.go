package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hcjokem/llm-gateway/internal/config"
	"github.com/hcjokem/llm-gateway/internal/handler"
	"github.com/hcjokem/llm-gateway/internal/middleware"
	"github.com/hcjokem/llm-gateway/internal/provider"
	"github.com/hcjokem/llm-gateway/internal/repository"
	"github.com/hcjokem/llm-gateway/internal/service"
	"github.com/hcjokem/llm-gateway/internal/util"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger := util.NewLogger(cfg.LogLevel)

	// 初始化数据库
	db, err := config.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化 repositories
	modelRepo := repository.NewModelRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	adminUserRepo := repository.NewAdminUserRepository(db)

	// 初始化 providers
	providers := map[string]provider.Provider{
		"openai":     provider.NewOpenAIProvider(),
		"anthropic":  provider.NewAnthropicProvider(),
		"zhipu":      provider.NewZhipuProvider(),
		"qwen":       provider.NewQwenProvider(),
	}

	// 初始化 services
	modelService := service.NewModelService(modelRepo)
	keyService := service.NewKeyService(keyRepo, modelRepo)
	usageService := service.NewUsageService(usageRepo, alertRepo)
	billingService := service.NewBillingService(billingRepo, modelRepo)
	alertService := service.NewAlertService(alertRepo)
	proxyService := service.NewProxyService(providers, keyRepo, usageRepo, modelRepo, logger)

	// 初始化 handlers
	adminHandler := handler.NewAdminHandler(
		modelService,
		keyService,
		usageService,
		billingService,
		alertService,
		logger,
	)
	proxyHandler := handler.NewProxyHandler(proxyService, logger)

	// 初始化 Gin 路由
	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// 管理端 API (需要 JWT 认证)
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.Auth(cfg.JWTSecret, adminUserRepo))
	{
		// 模型管理
		adminGroup.GET("/models", adminHandler.GetModels)
		adminGroup.POST("/models", adminHandler.CreateModel)
		adminGroup.PUT("/models/:id", adminHandler.UpdateModel)
		adminGroup.DELETE("/models/:id", adminHandler.DeleteModel)

		// Key 管理
		adminGroup.GET("/keys", adminHandler.GetKeys)
		adminGroup.POST("/keys", adminHandler.CreateKey)
		adminGroup.PUT("/keys/:id", adminHandler.UpdateKey)
		adminGroup.DELETE("/keys/:id", adminHandler.DeleteKey)

		// 用量统计
		adminGroup.GET("/usage", adminHandler.GetUsageStats)
		adminGroup.GET("/usage/keys/:key", adminHandler.GetKeyUsage)
		adminGroup.GET("/usage/realtime", adminHandler.GetRealtimeUsage)

		// 计费配置
		adminGroup.GET("/pricing", adminHandler.GetPricing)
		adminGroup.PUT("/pricing", adminHandler.UpdatePricing)
		adminGroup.GET("/packages", adminHandler.GetPackages)
		adminGroup.POST("/packages", adminHandler.CreatePackage)
		adminGroup.PUT("/packages/:id", adminHandler.UpdatePackage)
		adminGroup.DELETE("/packages/:id", adminHandler.DeletePackage)

		// 告警配置
		adminGroup.GET("/alerts", adminHandler.GetAlerts)
		adminGroup.PUT("/alerts/:id", adminHandler.UpdateAlert)
	}

	// 代理端 API (兼容 OpenAI 格式)
	proxyGroup := r.Group("/v1")
	proxyGroup.Use(middleware.KeyAuth(keyRepo))
	{
		proxyGroup.POST("/chat/completions", proxyHandler.ChatCompletion)
		proxyGroup.POST("/embeddings", proxyHandler.Embedding)
	}

	// 启动 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info(fmt.Sprintf("Starting server on port %d", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("Server exited")
}
