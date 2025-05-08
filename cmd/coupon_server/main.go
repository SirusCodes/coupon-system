// Package main is the entry point of the coupon system server.
//
//	@title						Coupon System API
//	@version					1.0
//	@BasePath					/
//	@securityDefinitions.apiKey	BearerAuth
//	@in							header
//	@name						Authorization

package main

import (
	"context"
	"coupon-system/internal/api/handlers"
	"coupon-system/internal/auth"
	"coupon-system/internal/caching"
	"coupon-system/internal/config"
	"coupon-system/internal/models"
	"coupon-system/internal/services"
	"coupon-system/internal/storage/database"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "coupon-system/docs"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load Configuration. In this diff, we will not implement the loading of configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Database
	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto Migrate the schemas
	err = db.AutoMigrate(&models.Coupon{}, &models.UserCouponUsage{})
	if err != nil {
		log.Fatalf("failed to automigrate database: %v", err)
	}

	// Initialize Cache
	cache := caching.NewLRUCache[string, *models.Coupon](cfg.CacheSize, time.Duration(cfg.CacheTTLMinutes)*time.Minute)

	// Initialize Storage
	couponStorage := database.NewSQLiteStore(db)

	// Initialize Service
	couponService := services.NewCouponService(couponStorage, cache)

	// Initialize Handlers
	couponHandlers := handlers.NewCouponHandlers(couponService)
	authHandlers := handlers.NewAuthHandlers()

	// Setup Gin Router
	router := gin.Default()
	// Apply CORS middleware to allow all origins, headers, and methods
	router.Use(cors.Default())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Define Routes
	adminGroup := router.Group("/admin")
	{
		adminGroup.POST("/coupons", couponHandlers.CreateCoupon)
	}

	couponsGroup := router.Group("/coupons")
	{
		couponsGroup.GET("/applicable", auth.AuthMiddleware(), couponHandlers.GetApplicableCoupons)
		couponsGroup.POST("/validate", auth.AuthMiddleware(), couponHandlers.ValidateCoupon)
	}

	router.POST("/generate-tokens", authHandlers.GenerateTokenHandler)
	{
	}

	// Start HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort, // Default to ":8080"
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on http://localhost:%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
