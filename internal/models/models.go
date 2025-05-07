package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	ID                    string     `json:"id" gorm:"primaryKey;column:id"`                                          // Auto-generated unique ID (e.g., UUID)
	CouponCode            string     `json:"coupon_code" gorm:"unique;column:coupon_code"`                            // User-facing unique identifier
	ExpiryDate            time.Time  `json:"expiry_date" gorm:"column:expiry_date"`                                   // Expiry date of the coupon
	UsageType             string     `json:"usage_type" gorm:"column:usage_type"`                                     // "one_time", "multi_use", "time_based"
	ApplicableMedicineIDs []string   `json:"applicable_medicine_ids" gorm:"type:text;column:applicable_medicine_ids"`           // Store as JSON string or comma-separated in DB
	ApplicableCategories  []string   `json:"applicable_categories" gorm:"type:text;column:applicable_categories"`               // Store as JSON string or comma-separated in DB
	MinOrderValue         float64    `json:"min_order_value" gorm:"column:min_order_value"`                           // Minimum order value for the coupon to be applicable
	ValidTimeWindowStart  *time.Time `json:"valid_time_window_start,omitempty" gorm:"column:valid_time_window_start"` // Pointer for optionality
	ValidTimeWindowEnd    *time.Time `json:"valid_time_window_end,omitempty" gorm:"column:valid_time_window_end"`     // Pointer for optionality
	TermsAndConditions    string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`                 // Terms and conditions for the coupon
	DiscountType          string     `json:"discount_type" gorm:"column:discount_type"`                               // e.g., "percentage", "fixed_amount"
	DiscountValue         float64    `json:"discount_value" gorm:"column:discount_value"`                             // The amount or percentage of discount
	MaxUsagePerUser       int        `json:"max_usage_per_user" gorm:"column:max_usage_per_user"`                     // 0 for unlimited
	MaxTotalUsage         int        `json:"max_total_usage" gorm:"column:max_total_usage"`                           // For "multi_use" if there's a global cap, 0 for unlimited
	CurrentTotalUsage     int        `json:"current_total_usage" gorm:"column:current_total_usage"`                   // Current total usage of the coupon
	CreatedAt             time.Time  `json:"created_at" gorm:"column:created_at"`                                     // Timestamp of when the coupon was created
	UpdatedAt             time.Time  `json:"updated_at" gorm:"column:updated_at"`                                     // Timestamp of when the coupon was last updated
	gorm.Model
}

type UserCouponUsage struct {
	UserID    string `gorm:"primaryKey;column:user_id"`
	CouponID  string `gorm:"primaryKey;column:coupon_id"`
	TimesUsed int    `gorm:"column:times_used"`
}

// Scan implements the sql.Scanner interface for ApplicableMedicineIDs.
func (c *Coupon) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), &c.ApplicableMedicineIDs)
	case []byte:
		return json.Unmarshal(v, &c.ApplicableMedicineIDs)
	default:
		return errors.New("incompatible type for ApplicableMedicineIDs")
	}
}

// Value implements the driver.Valuer interface for ApplicableMedicineIDs.
func (c Coupon) Value() (driver.Value, error) {
	return json.Marshal(c.ApplicableMedicineIDs)
}

// Scan implements the sql.Scanner interface for ApplicableCategories.
func (c *Coupon) ScanApplicableCategories(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), &c.ApplicableCategories)
	case []byte:
		return json.Unmarshal(v, &c.ApplicableCategories)
	default:
		return errors.New("incompatible type for ApplicableCategories")
	}
}

// Value implements the driver.Valuer interface for ApplicableCategories.
func (c Coupon) ValueApplicableCategories() (driver.Value, error) {
	return json.Marshal(c.ApplicableCategories)
}
