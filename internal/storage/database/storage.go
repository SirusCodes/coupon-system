package database

import (
	"context"
	"coupon-system/internal/models"
	"time"
)

type CouponStorage interface {
	CreateCoupon(ctx context.Context, coupon *models.Coupon) error
	GetCouponByCode(ctx context.Context, couponCode string) (*models.Coupon, error)
	UpdateCouponUsage(ctx context.Context, coupon *models.Coupon, userID string) error                                                                                     // Handles usage count and user-specific usage
	GetApplicableCoupons(ctx context.Context, timestamp time.Time, orderTotal float64, medicineIDs []string, categoryIDs []string, userID string) ([]models.Coupon, error) // For finding applicable coupons
	GetUserUsageForCoupon(ctx context.Context, userID string, couponID string) (int, error)
}
