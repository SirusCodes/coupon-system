package services

import (
	"coupon-system/internal/models"
	"slices"
)

type Discount interface {
	CalculateDiscount() float64
}

type MedicineDiscount struct {
	coupon    *models.Coupon
	cartItems []models.CartItem
}

func NewMedicineDiscount(coupon *models.Coupon, cartItems []models.CartItem) *MedicineDiscount {
	return &MedicineDiscount{
		coupon:    coupon,
		cartItems: cartItems,
	}
}

func (md *MedicineDiscount) CalculateDiscount() float64 {
	var discountApplicableMeds float64
	for _, item := range md.cartItems {
		if slices.Contains(md.coupon.MedicineIDs, models.Medicine{ID: item.ID}) {
			discountApplicableMeds += item.Price
		}
	}

	return calculateDiscountValue(discountApplicableMeds, md.coupon.DiscountValue, md.coupon.DiscountType)
}

type CategoryDiscount struct {
	coupon    *models.Coupon
	cartItems []models.CartItem
}

func NewCategoryDiscount(coupon *models.Coupon, cartItems []models.CartItem) *CategoryDiscount {
	return &CategoryDiscount{
		coupon:    coupon,
		cartItems: cartItems,
	}
}

func (cd *CategoryDiscount) CalculateDiscount() float64 {
	var discountApplicableCategories float64
	for _, item := range cd.cartItems {
		if slices.Contains(cd.coupon.Categories, models.Category{ID: item.Category}) {
			discountApplicableCategories += item.Price
		}
	}

	return calculateDiscountValue(discountApplicableCategories, cd.coupon.DiscountValue, cd.coupon.DiscountType)
}

type GeneralDiscount struct {
	coupon      *models.Coupon
	totalAmount float64
}

func NewGeneralDiscount(coupon *models.Coupon, totalAmount float64) *GeneralDiscount {
	return &GeneralDiscount{
		coupon:      coupon,
		totalAmount: totalAmount,
	}
}

func (gd *GeneralDiscount) CalculateDiscount() float64 {
	return calculateDiscountValue(gd.totalAmount, gd.coupon.DiscountValue, gd.coupon.DiscountType)
}

func calculateDiscountValue(amountForDiscount, discount float64, discountType string) float64 {
	if discountType == "fixed_amount" {
		return discount
	} else if discountType == "percentage" {
		return (amountForDiscount * (discount / 100))
	}
	return 0
}
