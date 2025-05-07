package services

import (
	"context"
	"coupon-system/internal/caching"
	"coupon-system/internal/models"
	"coupon-system/internal/storage/database"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

type CouponService struct {
	storage database.CouponStorage
	cache   caching.Cache[string, *models.Coupon]
}

func NewCouponService(storage database.CouponStorage, cache caching.Cache[string, *models.Coupon]) *CouponService {
	return &CouponService{
		storage: storage,
		cache:   cache,
	}
}

// CreateCoupon handles the creation of a new coupon.
func (s *CouponService) CreateCoupon(ctx context.Context, req *models.CreateCouponRequest) error {
	if req.CouponCode == "" {
		return fmt.Errorf("coupon code is required")
	}
	if req.ExpiryDate.IsZero() {
		return fmt.Errorf("expiry date is required")
	}
	if req.DiscountType == "" {
		return fmt.Errorf("discount type is required")
	}
	if req.DiscountValue <= 0 {
		return fmt.Errorf("discount value must be greater than 0")
	}

	couponID := uuid.New().String()

	coupon := &models.Coupon{
		ID:                    couponID,
		CouponCode:            req.CouponCode,
		ExpiryDate:            req.ExpiryDate,
		UsageType:             req.UsageType,
		ApplicableMedicineIDs: req.ApplicableMedicineIDs,
		ApplicableCategories:  req.ApplicableCategories,
		MinOrderValue:         req.MinOrderValue,
		ValidTimeWindowStart:  req.ValidTimeWindowStart,
		ValidTimeWindowEnd:    req.ValidTimeWindowEnd,
		TermsAndConditions:    req.TermsAndConditions,
		DiscountType:          req.DiscountType,
		DiscountValue:         req.DiscountValue,
		MaxUsagePerUser:       req.MaxUsagePerUser,
		MaxTotalUsage:         req.MaxTotalUsage,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	return s.storage.CreateCoupon(ctx, coupon)
}

// ValidateCoupon handles the validation of a coupon against a cart.
func (s *CouponService) ValidateCoupon(ctx context.Context, req *models.ValidateCouponRequest) (*models.ValidateCouponResponse, error) {
	// Fetch coupon from cache, fallback to storage
	cachedCoupon, ok := s.cache.Get(req.CouponCode)
	var coupon *models.Coupon
	if ok {
		coupon = cachedCoupon
	} else {
		var err error
		coupon, err = s.storage.GetCouponByCode(ctx, req.CouponCode)
		if err != nil {
			if err == sql.ErrNoRows {
				return &models.ValidateCouponResponse{IsValid: false, Message: "Coupon not found"}, nil
			}
			return nil, fmt.Errorf("error fetching coupon: %w", err)
		}
	}

	if coupon == nil {
		return &models.ValidateCouponResponse{IsValid: false, Message: "Coupon not found"}, nil
	}

	validators := []CouponValidator{
		NewExpiryDateValidator(),
		NewMinOrderValueValidator(),
		NewApplicableCategoriesValidator(),
		NewMaxUsagePerUserValidator(s.storage),
		NewMaxTotalUsageValidator(),
	}

	for _, validator := range validators {
		err := validator.Validate(coupon, req)

		if err != nil {
			return &models.ValidateCouponResponse{
				IsValid: false,
				Message: err.Error(),
			}, nil
		}
	}

	// Calculate discount
	var itemsDiscount float64

	if coupon.DiscountType == "fixed_amount" {
		itemsDiscount = coupon.DiscountValue
		if itemsDiscount > req.OrderTotal {
			itemsDiscount = req.OrderTotal
		}
	} else if coupon.DiscountType == "percentage" {
		itemsDiscount = req.OrderTotal * (coupon.DiscountValue / 100.0)
	}

	discountDetails := &models.DiscountDetails{
		ItemsDiscount:   itemsDiscount,
		ChargesDiscount: 0,
		TotalDiscount:   itemsDiscount,
	}

	err := s.storage.UpdateCouponUsage(ctx, coupon, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("error updating coupon usage: %w", err)
	}

	s.cache.Delete(req.CouponCode)

	return &models.ValidateCouponResponse{IsValid: true, Discount: discountDetails, Message: "Coupon applied successfully"}, nil
}

// GetApplicableCoupons fetches all coupons applicable to a given cart.
func (s *CouponService) GetApplicableCoupons(ctx context.Context, req *models.ApplicableCouponsRequest) (*models.ApplicableCouponsResponse, error) {
	coupons, err := s.storage.GetAllCoupons(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching all coupons: %w", err)
	}

	var applicableCoupons []models.ApplicableCoupon
	for _, coupon := range coupons {
		// Soft validation checks
		if req.Timestamp.After(coupon.ExpiryDate) {
			continue // Expired
		}
		if req.OrderTotal < coupon.MinOrderValue {
			continue // Minimum order value not met
		}
		if coupon.ValidTimeWindowStart != nil && req.Timestamp.Before(*coupon.ValidTimeWindowStart) {
			continue // Not within valid time window
		}
		if coupon.ValidTimeWindowEnd != nil && req.Timestamp.After(*coupon.ValidTimeWindowEnd) {
			continue // Not within valid time window
		}
		if len(coupon.ApplicableMedicineIDs) > 0 {
			found := false
			for _, item := range req.CartItems {
				if slices.Contains(coupon.ApplicableMedicineIDs, item.ID) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if len(coupon.ApplicableCategories) > 0 {
			found := false
			for _, item := range req.CartItems {
				if slices.Contains(coupon.ApplicableCategories, item.Category) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if coupon.MaxUsagePerUser > 0 {
			userUsage, _ := s.storage.GetUserUsageForCoupon(ctx, req.UserID, coupon.ID)
			if userUsage >= coupon.MaxUsagePerUser {
				continue // Max usage per user exceeded
			}
		}
		applicableCoupons = append(applicableCoupons, models.ApplicableCoupon{CouponCode: coupon.CouponCode, DiscountValue: coupon.DiscountValue, DiscountType: coupon.DiscountType})
	}
	return &models.ApplicableCouponsResponse{ApplicableCoupons: applicableCoupons}, nil
}
