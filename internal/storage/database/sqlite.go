package database

import (
	"context"
	"coupon-system/internal/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SQLiteStore implements the CouponStorage interface using SQLite.
type SQLiteStore struct {
	db *gorm.DB
}

// NewSQLiteStore creates a new instance of SQLiteStore.
func NewSQLiteStore(db *gorm.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	sqlDB, _ := s.db.DB()
	return sqlDB.Close()
}

// CreateCoupon inserts a new coupon into the database.
func (s *SQLiteStore) CreateCoupon(ctx context.Context, coupon *models.Coupon) error {

	err := s.db.WithContext(ctx).Create(coupon).Error

	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}

	return nil
}

// GetCouponByCode retrieves a coupon by its coupon code.
func (s *SQLiteStore) GetCouponByCode(ctx context.Context, couponCode string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := s.db.WithContext(ctx).Where("coupon_code = ?", couponCode).First(&coupon).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Coupon not found is not an error in this context
		}
		return nil, err
	}

	return &coupon, nil
}

// UpdateCouponUsage atomically updates coupon usage counts and records user-specific usage within a transaction.
func (s *SQLiteStore) UpdateCouponUsage(ctx context.Context, coupon *models.Coupon, userID string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Increment overall usage count
	err := tx.Model(coupon).Update("current_total_usage", coupon.CurrentTotalUsage+1).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to increment current_total_usage: %w", err)
	}

	// Check max usage per user and increment
	var userUsage models.UserCouponUsage
	if userID != "" {
		userUsage = models.UserCouponUsage{UserID: userID, CouponID: coupon.ID, TimesUsed: 1}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "coupon_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"times_used": gorm.Expr("times_used + ?", 1)}),
	}).Create(&userUsage).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user coupon usage: %w", err)
	}
	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *SQLiteStore) getCouponByIDTx(ctx context.Context, tx *gorm.DB, couponID string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := tx.WithContext(ctx).Where("id = ?", couponID).First(&coupon).Error
	return &coupon, err
}

// GetAllCoupons retrieves all coupons from the database.
func (s *SQLiteStore) GetAllCoupons(ctx context.Context) ([]models.Coupon, error) {
	var coupons []models.Coupon
	err := s.db.WithContext(ctx).Find(&coupons).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all coupons: %w", err)
	}

	return coupons, nil
}

// GetUserUsageForCoupon retrieves the number of times a user has used a specific coupon.
func (s *SQLiteStore) GetUserUsageForCoupon(ctx context.Context, userID string, couponID string) (int, error) {
	var userUsage models.UserCouponUsage
	err := s.db.WithContext(ctx).Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&userUsage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil // No usage record found, treat as zero usage
		}
		return 0, fmt.Errorf("failed to get user usage for coupon: %w", err)
	}
	return userUsage.TimesUsed, nil
}
