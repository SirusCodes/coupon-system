package main

import (
	"context"
	"coupon-system/internal/api/handlers"
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

	"github.com/gin-gonic/gin"
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

	// Setup Gin Router
	router := gin.Default()

	// Define Routes
	adminGroup := router.Group("/admin")
	{
		adminGroup.POST("/coupons", couponHandlers.CreateCoupon)
	}

	couponsGroup := router.Group("/coupons")
	{
		couponsGroup.GET("/applicable", couponHandlers.GetApplicableCoupons)
		couponsGroup.POST("/validate", couponHandlers.ValidateCoupon)
	}

	// Start HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort, // Default to ":8080"
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
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
