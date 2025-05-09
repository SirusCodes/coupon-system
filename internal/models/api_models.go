package models

import "time"

// SuccessResponse represents a generic success response with a message.
type SuccessResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents a generic error response with an error message and details.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// CreateCouponRequest represents the request body for creating a coupon.
// @Description CreateCouponRequest represents the request to create a new coupon
type CreateCouponRequest struct {
	// CouponCode is the unique identifier for the coupon
	CouponCode string `json:"coupon_code" binding:"required"`
	// ExpiryDate is the date when the coupon will expire
	ExpiryDate time.Time `json:"expiry_date" binding:"required"`
	// UsageType is the type of usage allowed for the coupon (one_time, multi_use, time_based)
	UsageType string `json:"usage_type" binding:"required,oneof=one_time multi_use time_based"`
	// ApplicableMedicineIDs is a list of medicine IDs to which the coupon applies
	ApplicableMedicineIDs []string `json:"applicable_medicine_ids"`
	// ApplicableCategories is a list of categories to which the coupon applies
	ApplicableCategories []string `json:"applicable_categories"`
	// MinOrderValue is the minimum order value required to apply the coupon
	MinOrderValue float64 `json:"min_order_value"`
	// ValidTimeWindowStart is the start of the time window during which the coupon is valid
	ValidTimeWindowStart *time.Time `json:"valid_time_window_start,omitempty"`
	// ValidTimeWindowEnd is the end of the time window during which the coupon is valid
	ValidTimeWindowEnd *time.Time `json:"valid_time_window_end,omitempty"`
	// TermsAndConditions are the terms and conditions for the coupon
	TermsAndConditions string `json:"terms_and_conditions"`
	// DiscountType is the type of discount (percentage or fixed_amount)
	DiscountType string `json:"discount_type" binding:"required,oneof=percentage fixed_amount"`
	// DiscountValue is the value of the discount
	DiscountValue float64 `json:"discount_value" binding:"required"`
	// MaxUsagePerUser is the maximum number of times a user can use the coupon
	MaxUsagePerUser int `json:"max_usage_per_user"`
	// MaxTotalUsage is the maximum number of times the coupon can be used in total
	MaxTotalUsage int `json:"max_total_usage"`
}

// ApplicableCouponsRequest represents the request body for finding applicable coupons.
// @Description ApplicableCouponsRequest represents the request to find applicable coupons for a cart
type ApplicableCouponsRequest struct {
	// CartItems is the list of items in the cart
	CartItems []CartItem `json:"cart_items" binding:"required"`
	// OrderTotal is the total value of the order
	OrderTotal float64 `json:"order_total" binding:"required"`
	// Timestamp is the current timestamp
	Timestamp time.Time `json:"timestamp" binding:"required"`
}

// ApplicableCoupon represents a coupon that is applicable to the current cart.
type ApplicableCoupon struct {
	CouponCode    string  `json:"coupon_code"`    // CouponCode is the code of the coupon
	DiscountValue float64 `json:"discount_value"` // DiscountValue is the discount value for this coupon
	DiscountType  string  `json:"discount_type"`  // DiscountType is the type of discount (percentage or fixed_amount)
}

// ApplicableCouponsResponse represents the response body for applicable coupons.
type ApplicableCouponsResponse struct {
	ApplicableCoupons []ApplicableCoupon `json:"applicable_coupons"`
}

// ValidateCouponRequest represents the request body for validating a coupon.
type ValidateCouponRequest struct {
	CouponCode string     `json:"coupon_code" binding:"required"`
	CartItems  []CartItem `json:"cart_items" binding:"required"`
	OrderTotal float64    `json:"order_total" binding:"required"`
	Timestamp  time.Time  `json:"timestamp" binding:"required"`
}

// DiscountDetails represents the details of the discount applied by a coupon.
type DiscountDetails struct {
	ItemsDiscount   float64 `json:"items_discount"`
	ChargesDiscount float64 `json:"charges_discount"`
	TotalDiscount   float64 `json:"total_discount"`
}

// ValidateCouponResponse represents the response body for validating a coupon.
type ValidateCouponResponse struct {
	IsValid  bool             `json:"is_valid"`
	Discount *DiscountDetails `json:"discount,omitempty"`
	Message  string           `json:"message"`
}

type CartItem struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`    // Price per unit
	Quantity int     `json:"quantity"` // Quantity of this item in the cart
}
