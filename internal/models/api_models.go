package models

import "time"

// @Description SuccessResponse represents a generic success response with a message.
type SuccessResponse struct {
	Message string `json:"message"`
}

// @Description ErrorResponse represents a generic error response with an error message and details.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// @Description CreateCouponRequest represents the request to create a new coupon
type CreateCouponRequest struct {
	CouponCode            string     `json:"coupon_code" binding:"required"`
	ExpiryDate            time.Time  `json:"expiry_date" binding:"required"`
	UsageType             string     `json:"usage_type" binding:"required,oneof=one_time multi_use time_based"`
	ApplicableMedicineIDs []string   `json:"applicable_medicine_ids"`
	ApplicableCategories  []string   `json:"applicable_categories"`
	MinOrderValue         float64    `json:"min_order_value"`
	ValidTimeWindowStart  *time.Time `json:"valid_time_window_start,omitempty"`
	ValidTimeWindowEnd    *time.Time `json:"valid_time_window_end,omitempty"`
	TermsAndConditions    string     `json:"terms_and_conditions"`
	// DiscountType is the type of discount (percentage or fixed_amount)
	DiscountType    string  `json:"discount_type" binding:"required,oneof=percentage fixed_amount"`
	DiscountValue   float64 `json:"discount_value" binding:"required"`
	MaxUsagePerUser int     `json:"max_usage_per_user"`
	MaxTotalUsage   int     `json:"max_total_usage"`
}

// @Description ApplicableCouponsRequest represents the request to find applicable coupons for a cart
type ApplicableCouponsRequest struct {
	CartItems  []CartItem `json:"cart_items" binding:"required"`
	OrderTotal float64    `json:"order_total" binding:"required"`
	Timestamp  time.Time  `json:"timestamp" binding:"required"`
}

// @Description ApplicableCoupon represents a coupon that is applicable to the current cart.
type ApplicableCoupon struct {
	CouponCode    string  `json:"coupon_code"`
	DiscountValue float64 `json:"discount_value"`
	DiscountType  string  `json:"discount_type"`
}

// @Description ApplicableCouponsResponse represents the response body for applicable coupons.
type ApplicableCouponsResponse struct {
	ApplicableCoupons []ApplicableCoupon `json:"applicable_coupons"`
}

// @Description ValidateCouponRequest represents the request body for validating a coupon.
type ValidateCouponRequest struct {
	CouponCode string     `json:"coupon_code" binding:"required"`
	CartItems  []CartItem `json:"cart_items" binding:"required"`
	OrderTotal float64    `json:"order_total" binding:"required"`
	Timestamp  time.Time  `json:"timestamp" binding:"required"`
}

// @Description DiscountDetails represents the details of the discount applied by a coupon.
type DiscountDetails struct {
	ItemsDiscount float64 `json:"items_discount"`
	TotalDiscount float64 `json:"total_discount"`
}

// @Description ValidateCouponResponse represents the response body for validating a coupon.
type ValidateCouponResponse struct {
	IsValid  bool             `json:"is_valid"`
	Discount *DiscountDetails `json:"discount,omitempty"`
	Message  string           `json:"message"`
}

// @Description CartItem holds cart items
type CartItem struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}
