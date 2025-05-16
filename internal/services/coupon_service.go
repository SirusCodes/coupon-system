package services

import (
	"context"
	"coupon-system/internal/caching"
	"coupon-system/internal/models"
	"coupon-system/internal/storage/database"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CouponService struct {
	storage                database.CouponStorage
	applicableCouponsCache caching.Cache[string, *models.ApplicableCouponsResponse]
}

func NewCouponService(storage database.CouponStorage, applicableCouponsCache caching.Cache[string, *models.ApplicableCouponsResponse]) *CouponService {
	return &CouponService{
		storage:                storage,
		applicableCouponsCache: applicableCouponsCache,
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
		ID:                   couponID,
		CouponCode:           req.CouponCode,
		ExpiryDate:           req.ExpiryDate,
		UsageType:            req.UsageType,
		MinOrderValue:        req.MinOrderValue,
		ValidTimeWindowStart: req.ValidTimeWindowStart,
		ValidTimeWindowEnd:   req.ValidTimeWindowEnd,
		TermsAndConditions:   req.TermsAndConditions,
		DiscountType:         req.DiscountType,
		DiscountValue:        req.DiscountValue,
		MaxUsagePerUser:      req.MaxUsagePerUser,
		MaxTotalUsage:        req.MaxTotalUsage,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	for _, id := range req.ApplicableMedicineIDs {
		coupon.MedicineIDs = append(coupon.MedicineIDs, models.Medicine{ID: strings.TrimSpace(id)})
	}
	for _, category := range req.ApplicableCategories {
		coupon.Categories = append(coupon.Categories, models.Category{ID: strings.TrimSpace(category)})
	}

	return s.storage.CreateCoupon(ctx, coupon)
}

// ValidateCoupon handles the validation of a coupon against a cart.
func (s *CouponService) ValidateCoupon(ctx context.Context, userID string, req *models.ValidateCouponRequest) (*models.ValidateCouponResponse, error) {
	coupon, err := s.storage.GetCouponByCode(ctx, req.CouponCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.ValidateCouponResponse{IsValid: false, Message: "Coupon not found"}, nil
		}
		return nil, fmt.Errorf("error fetching coupon: %w", err)
	}

	if coupon == nil {
		return &models.ValidateCouponResponse{IsValid: false, Message: "Coupon not found"}, nil
	}

	validators := []CouponValidator{
		NewExpiryDateValidator(),
		NewMinOrderValueValidator(),
		NewApplicableCategoriesValidator(),
		NewMaxUsagePerUserValidator(s.storage, userID),
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
	itemsDiscount := calculateDiscount(coupon, req.CartItems, req.OrderTotal)

	discountDetails := &models.DiscountDetails{
		ItemsDiscount: itemsDiscount,
		TotalDiscount: itemsDiscount,
	}

	err = s.storage.UpdateCouponUsage(ctx, coupon, userID)
	if err != nil {
		return nil, fmt.Errorf("error updating coupon usage: %w", err)
	}

	return &models.ValidateCouponResponse{IsValid: true, Discount: discountDetails, Message: "Coupon applied successfully"}, nil
}

// GetApplicableCoupons fetches all coupons applicable to a given cart.
func (s *CouponService) GetApplicableCoupons(ctx context.Context, userID string, req *models.ApplicableCouponsRequest) (*models.ApplicableCouponsResponse, error) {
	cacheKey := generateApplicableCouponsCacheKey(userID, req)
	if cachedResponse, found := s.applicableCouponsCache.Get(cacheKey); found {
		return cachedResponse, nil
	}

	coupons, err := s.storage.GetApplicableCoupons(ctx, req.Timestamp, req.OrderTotal, getMedicineIDsFromCart(req.CartItems), getCategoryIDsFromCart(req.CartItems), userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching all coupons: %w", err)
	}

	var applicableCoupons []models.ApplicableCoupon
	for _, coupon := range coupons {
		applicableCoupons = append(applicableCoupons, models.ApplicableCoupon{
			CouponCode:    coupon.CouponCode,
			DiscountValue: coupon.DiscountValue,
			DiscountType:  coupon.DiscountType,
			Discount:      calculateDiscount(&coupon, req.CartItems, req.OrderTotal),
		})
	}

	response := &models.ApplicableCouponsResponse{ApplicableCoupons: applicableCoupons}
	s.applicableCouponsCache.Set(cacheKey, response)
	return response, nil
}

func calculateDiscount(coupon *models.Coupon, cartItems []models.CartItem, orderTotal float64) float64 {
	discountFor := []string{}

	for _, item := range cartItems {
		if slices.Contains(coupon.MedicineIDs, models.Medicine{ID: item.ID}) {
			discountFor = append(discountFor, "medicine")
		}
		if slices.Contains(coupon.Categories, models.Category{ID: item.Category}) {
			discountFor = append(discountFor, "category")
		}
	}

	if len(discountFor) == 0 {
		generalDiscount := NewGeneralDiscount(coupon, orderTotal)
		return generalDiscount.CalculateDiscount()
	}

	discountCalculator := []Discount{}

	if slices.Contains(discountFor, "medicine") {
		medicineDiscount := NewMedicineDiscount(coupon, cartItems)
		discountCalculator = append(discountCalculator, medicineDiscount)
	}

	if slices.Contains(discountFor, "category") {
		categoryDiscount := NewCategoryDiscount(coupon, cartItems)
		discountCalculator = append(discountCalculator, categoryDiscount)
	}

	var maxDiscount float64

	for _, discount := range discountCalculator {
		maxDiscount = math.Max(maxDiscount, discount.CalculateDiscount())
	}

	return math.Max(maxDiscount, 0)
}

func getMedicineIDsFromCart(cartItems []models.CartItem) []string {
	medicineIDs := make([]string, 0, len(cartItems))
	for _, item := range cartItems {
		medicineIDs = append(medicineIDs, item.ID)
	}
	return medicineIDs
}

func getCategoryIDsFromCart(cartItems []models.CartItem) []string {
	categoryIDs := make([]string, 0, len(cartItems))
	for _, item := range cartItems {
		categoryIDs = append(categoryIDs, item.Category)
	}
	return categoryIDs
}

func generateApplicableCouponsCacheKey(userID string, req *models.ApplicableCouponsRequest) string {
	// Use a combination of userID and the request parameters to create a unique key.
	// Be mindful of the order of items in slices for consistent key generation.
	// Marshalling the request can create a consistent string representation.
	reqBytes, _ := json.Marshal(req)
	data := userID + string(reqBytes)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
