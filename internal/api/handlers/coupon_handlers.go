package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"coupon-system/internal/models"
	"coupon-system/internal/services"
)

// CouponHandlers defines the handlers for coupon-related API endpoints.
type CouponHandlers struct {
	couponService *services.CouponService
}

// NewCouponHandlers creates a new CouponHandlers instance.
func NewCouponHandlers(couponService *services.CouponService) *CouponHandlers {
	return &CouponHandlers{
		couponService: couponService,
	}
}

// CreateCoupon handles the creation of a new coupon.
func (h *CouponHandlers) CreateCoupon(c *gin.Context) {
	var req models.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.couponService.CreateCoupon(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Coupon created successfully"})
}

// GetApplicableCoupons retrieves the coupons applicable to a cart.
func (h *CouponHandlers) GetApplicableCoupons(c *gin.Context) {
	var req models.ApplicableCouponsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	applicableCoupons, err := h.couponService.GetApplicableCoupons(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applicable coupons", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, applicableCoupons)
}

// ValidateCoupon validates a coupon against a shopping cart.
func (h *CouponHandlers) ValidateCoupon(c *gin.Context) {
	var req models.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationResponse, err := h.couponService.ValidateCoupon(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate coupon", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, validationResponse)
}
