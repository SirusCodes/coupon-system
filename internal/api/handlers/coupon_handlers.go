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
// CreateCoupon godoc
//
//	@Summary		Create a new coupon
//	@Security		BearerAuth
//	@Description	Creates a new coupon with the provided details.
//	@Tags			coupons
//	@Accept			json
//	@Produce		json
//	@Param			coupon	body		models.CreateCouponRequest	true	"Coupon Data"
//	@Success		201		{object}	models.SuccessResponse		"Coupon created successfully"
//	@Failure		400		{object}	models.ErrorResponse		"Bad request"
//	@Failure		500		{object}	models.ErrorResponse		"Internal server error"
//	@Router			/admin/coupons [post]
func (h *CouponHandlers) CreateCoupon(c *gin.Context) {
	var req models.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	if err := h.couponService.CreateCoupon(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create coupon", Details: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{Message: "Coupon created successfully"})
}

// GetApplicableCoupons retrieves the coupons applicable to a cart.
// GetApplicableCoupons godoc
//
//	@Summary		Get applicable coupons
//	@Security		BearerAuth
//	@Description	Retrieves a list of coupons applicable to the current cart.
//	@Tags			coupons
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ApplicableCouponsRequest		true	"Cart details"
//	@Success		200		{object}	models.ApplicableCouponsResponse	"List of applicable coupons"
//	@Failure		400		{object}	models.ErrorResponse				"Bad request"
//	@Failure		500		{object}	models.ErrorResponse				"Internal server error"
//	@Router			/coupons/applicable [get]
func (h *CouponHandlers) GetApplicableCoupons(c *gin.Context) {
	var req models.ApplicableCouponsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	applicableCoupons, err := h.couponService.GetApplicableCoupons(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get applicable coupons", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, applicableCoupons)
}

// ValidateCoupon validates a coupon against a shopping cart.
// ValidateCoupon godoc
//
//	@Summary		Validate a coupon
//	@Security		BearerAuth
//	@Description	Validates a coupon code against the provided cart details.
//	@Tags			coupons
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ValidateCouponRequest	true	"Validation request"
//	@Success		200		{object}	models.ValidateCouponResponse	"Coupon validation result"
//	@Failure		400		{object}	models.ErrorResponse			"Bad request"
//	@Failure		500		{object}	models.ErrorResponse			"Internal server error"
//	@Router			/coupons/validate [post]
func (h *CouponHandlers) ValidateCoupon(c *gin.Context) {
	var req models.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request", Details: err.Error()})
		return
	}

	validationResponse, err := h.couponService.ValidateCoupon(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to validate coupon", Details: err.Error()})
		return
	}

	c.JSON(http.StatusOK, validationResponse)
}
