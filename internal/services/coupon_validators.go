package services

import (
	"context"
	"coupon-system/internal/models"
	"coupon-system/internal/storage/database"
	"fmt"
)

// CouponValidator defines the interface for coupon validation rules.
type CouponValidator interface {
	Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error
}

// ExpiryDateValidator validates if the coupon has expired.
type ExpiryDateValidator struct{}

func NewExpiryDateValidator() *ExpiryDateValidator {
	return &ExpiryDateValidator{}
}
func (v *ExpiryDateValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if req.Timestamp.After(coupon.ExpiryDate) {
		return fmt.Errorf("coupon has expired")
	}
	return nil
}

// MinOrderValueValidator validates if the order total meets the minimum requirement.
type MinOrderValueValidator struct{}

func NewMinOrderValueValidator() *MinOrderValueValidator {
	return &MinOrderValueValidator{}
}

func (v *MinOrderValueValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if req.OrderTotal < coupon.MinOrderValue {
		return fmt.Errorf("minimum order value of %.2f required", coupon.MinOrderValue)
	}
	return nil
}

// ValidTimeWindowValidator validates if the coupon is used within its valid time window.
type ValidTimeWindowValidator struct{}

func NewValidTimeWindowValidator() *ValidTimeWindowValidator {
	return &ValidTimeWindowValidator{}
}

func (v *ValidTimeWindowValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if coupon.ValidTimeWindowStart != nil && req.Timestamp.Before(*coupon.ValidTimeWindowStart) {
		return fmt.Errorf("coupon is not yet valid")
	}
	if coupon.ValidTimeWindowEnd != nil && req.Timestamp.After(*coupon.ValidTimeWindowEnd) {
		return fmt.Errorf("coupon is no longer valid")
	}
	return nil
}

// ApplicableItemsValidator validates if the coupon is applicable to any items in the cart by medicine ID.
type ApplicableItemsValidator struct{}

func NewApplicableItemsValidator() *ApplicableItemsValidator {
	return &ApplicableItemsValidator{}
}

func (v *ApplicableItemsValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if len(coupon.MedicineIDs) > 0 {
		found := false
		for _, medicine := range coupon.MedicineIDs {
			for _, item := range req.CartItems {
				if medicine.ID == item.ID {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("coupon not applicable to any items in the cart")
		}
	}
	return nil
}

// ApplicableCategoriesValidator validates if the coupon is applicable to any categories in the cart.
type ApplicableCategoriesValidator struct{}

func NewApplicableCategoriesValidator() *ApplicableCategoriesValidator {
	return &ApplicableCategoriesValidator{}
}

func (v *ApplicableCategoriesValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if len(coupon.Categories) > 0 {
		found := false
		for _, category := range coupon.Categories {
			for _, item := range req.CartItems {
				if category.ID == item.Category {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("coupon not applicable to any categories in the cart")
		}
	}
	return nil
}

// MaxUsagePerUserValidator validates if the user has exceeded the maximum usage limit for this coupon.
type MaxUsagePerUserValidator struct {
	storage database.CouponStorage
	userID  string
}

func NewMaxUsagePerUserValidator(storage database.CouponStorage, userID string) *MaxUsagePerUserValidator {
	return &MaxUsagePerUserValidator{storage: storage, userID: userID}
}

func (v *MaxUsagePerUserValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if coupon.MaxUsagePerUser > 0 {
		userUsage, err := v.storage.GetUserUsageForCoupon(context.Background(), v.userID, coupon.ID)
		if err != nil {
			return fmt.Errorf("error checking user usage")
		}
		if userUsage >= coupon.MaxUsagePerUser {
			return fmt.Errorf("maximum usage per user exceeded")
		}
	}
	return nil
}

// MaxTotalUsageValidator validates if the coupon has exceeded its total usage limit.
type MaxTotalUsageValidator struct{}

func NewMaxTotalUsageValidator() *MaxTotalUsageValidator {
	return &MaxTotalUsageValidator{}
}

func (v *MaxTotalUsageValidator) Validate(coupon *models.Coupon, req *models.ValidateCouponRequest) error {
	if coupon.MaxTotalUsage > 0 && coupon.CurrentTotalUsage >= coupon.MaxTotalUsage {
		return fmt.Errorf("maximum total usage exceeded")
	}
	return nil
}
