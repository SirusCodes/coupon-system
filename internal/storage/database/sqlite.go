package database

import (
	"context"
	"coupon-system/internal/models"
	"errors"
	"fmt"
	"time"

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
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Ensure Medicine records exist and associate with coupon
	for _, medicine := range coupon.MedicineIDs {
		if err := tx.FirstOrCreate(&medicine, models.Medicine{ID: medicine.ID}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find or create medicine: %w", err)
		}
		// reload the medicine to get the full object in case it existed
		if err := tx.First(&medicine, "id = ?", medicine.ID).Error; err != nil {
			return fmt.Errorf("failed to find medicine after creation %w", err)
		}
		// Establish the many-to-many relationship
		err := tx.Model(coupon).Association("MedicineIDs").Append(&medicine)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to associate medicine with coupon: %w", err)
		}
	}

	// Ensure Category records exist and associate with coupon
	for _, category := range coupon.Categories {
		if err := tx.FirstOrCreate(&category, models.Category{ID: category.ID}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find or create category: %w", err)
		}
		// reload the category to get the full object in case it existed
		if err := tx.First(&category, "id = ?", category.ID).Error; err != nil {
			return fmt.Errorf("failed to find category after creation %w", err)
		}
		// Establish the many-to-many relationship
		err := tx.Model(coupon).Association("Categories").Append(&category)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to associate category with coupon: %w", err)
		}
	}

	// Create the coupon record
	err := tx.Create(coupon).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create coupon: %w", err)
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// GetCouponByCode retrieves a coupon by its coupon code.
func (s *SQLiteStore) GetCouponByCode(ctx context.Context, couponCode string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := s.db.WithContext(ctx).Where("coupon_code = ?", couponCode).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// GetAllCoupons retrieves all coupons from the database.
func (s *SQLiteStore) GetApplicableCoupons(ctx context.Context, timestamp time.Time, orderTotal float64, medicineIDs []string, categoryIDs []string, userID string) ([]models.Coupon, error) {
	var coupons []models.Coupon                            // Explicitly declare coupons
	query := s.db.WithContext(ctx).Model(&models.Coupon{}) // Start with the Coupon model
	// Basic filtering conditions
	query = query.
		Where("coupons.expiry_date > ?", timestamp).
		Where("coupons.min_order_value <= ?", orderTotal).
		Where("coupons.current_total_usage < coupons.max_total_usage OR coupons.max_total_usage = 0").
		Where("(coupons.valid_time_window_start IS NULL OR coupons.valid_time_window_start <= ?)", timestamp).
		Where("(coupons.valid_time_window_end IS NULL OR coupons.valid_time_window_end >= ?)", timestamp).
		// Filter out coupons that the user has already used up to the maximum allowed limit
		Joins("LEFT JOIN user_coupon_usages ON coupons.id = user_coupon_usages.coupon_id AND user_coupon_usages.user_id = ?", userID).
		Where("user_coupon_usages.times_used < coupons.max_usage_per_user OR coupons.max_usage_per_user = 0 OR user_coupon_usages.user_id IS NULL") // Include coupons the user hasn't used yet or have no per-user limit
	// Conditional filtering based on medicine and category IDs
	// This part implements the logic for general coupons and specific coupons
	query = query.Where("("+
		// Condition for general coupons
		"NOT EXISTS (SELECT 1 FROM coupon_medicine_ids WHERE coupon_id = coupons.id) AND NOT EXISTS (SELECT 1 FROM coupon_categories WHERE coupon_id = coupons.id)"+
		") OR ("+
		// Condition for coupons with associated medicine or category IDs
		"EXISTS (SELECT 1 FROM coupon_medicine_ids WHERE coupon_id = coupons.id AND medicine_id IN (?))"+
		"OR EXISTS (SELECT 1 FROM coupon_categories WHERE coupon_id = coupons.id AND category_id IN (?))"+
		")",
		medicineIDs,
		categoryIDs,
	)

	err := query.Find(&coupons).Error
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

// GetUserUsageForCoupon retrieves the number of times a user has used a specific coupon.
func (s *SQLiteStore) GetUserUsageForCoupon(ctx context.Context, userID string, couponID string) (int, error) {
	var userUsage models.UserCouponUsage
	err := s.db.WithContext(ctx).Where("user_id = ? AND coupon_id = ?", userID, couponID).First(&userUsage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // No usage record found, treat as zero usage
		}
		return 0, fmt.Errorf("failed to get user usage for coupon: %w", err)
	}
	return userUsage.TimesUsed, nil
}
