package models

import (
	"time"

	"gorm.io/gorm"
)

// Coupon represents a coupon entity.
type Coupon struct {
	ID                   string     `json:"id" gorm:"primaryKey;column:id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`                          // Auto-generated unique ID (e.g., UUID)
	CouponCode           string     `json:"coupon_code" gorm:"unique;column:coupon_code" example:"SUMMER20"`                                        // User-facing unique identifier
	ExpiryDate           time.Time  `json:"expiry_date" gorm:"column:expiry_date" example:"2024-12-31T23:59:59Z"`                                   // Expiry date of the coupon
	UsageType            string     `json:"usage_type" gorm:"column:usage_type" example:"multi_use"`                                                // "one_time", "multi_use", "time_based"
	MinOrderValue        float64    `json:"min_order_value" gorm:"column:min_order_value" example:"100.00"`                                         // Minimum order value for the coupon to be applicable
	ValidTimeWindowStart *time.Time `json:"valid_time_window_start,omitempty" gorm:"column:valid_time_window_start" example:"2024-01-01T00:00:00Z"` // Pointer for optionality
	MedicineIDs          []Medicine `gorm:"many2many:coupon_medicine_ids;"`
	Categories           []Category `gorm:"many2many:coupon_categories;"`

	ValidTimeWindowEnd *time.Time `json:"valid_time_window_end,omitempty" gorm:"column:valid_time_window_end" example:"2024-01-07T23:59:59Z"` // Pointer for optionality
	TermsAndConditions string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions" example:"Valid for new users only"`         // Terms and conditions for the coupon
	DiscountType       string     `json:"discount_type" gorm:"column:discount_type" example:"percentage"`                                     // e.g., "percentage", "fixed_amount"
	DiscountValue      float64    `json:"discount_value" gorm:"column:discount_value" example:"10.00"`                                        // The amount or percentage of discount
	MaxUsagePerUser    int        `json:"max_usage_per_user" gorm:"column:max_usage_per_user" example:"1"`                                    // 0 for unlimited
	MaxTotalUsage      int        `json:"max_total_usage" gorm:"column:max_total_usage" example:"100"`                                        // For "multi_use" if there's a global cap, 0 for unlimited
	CurrentTotalUsage  int        `json:"current_total_usage" gorm:"column:current_total_usage" example:"50"`                                 // Current total usage of the coupon
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at" example:"2024-01-01T00:00:00Z"`                                 // Timestamp of when the coupon was created
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at" example:"2024-01-01T00:00:00Z"`                                 // Timestamp of when the coupon was last updated
	gorm.Model
}

type UserCouponUsage struct {
	UserID    string `gorm:"primaryKey;column:user_id"`
	CouponID  string `gorm:"primaryKey;column:coupon_id"`
	TimesUsed int    `gorm:"column:times_used"`
}

type Medicine struct {
	ID string `gorm:"primaryKey"`
}

type Category struct {
	ID string `gorm:"primaryKey"`
}
