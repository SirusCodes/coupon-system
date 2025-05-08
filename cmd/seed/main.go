package main

import (
	"context"
	"errors"
	"log"
	"time"

	"coupon-system/internal/config"
	"coupon-system/internal/models"
	"coupon-system/internal/storage/database"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Database
	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto Migrate the schemas
	err = db.AutoMigrate(&models.Coupon{}, &models.UserCouponUsage{})
	if err != nil {
		log.Fatalf("failed to automigrate database: %v", err)
	}

	couponStorage := database.NewSQLiteStore(db)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Database connection successful.")

	// Seed the database
	err = seedDatabase(couponStorage)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}

	log.Println("Database seeding complete.")
}

// mockCouponData is a helper struct for defining mock coupon data with
// applicable medicine IDs and categories as string slices.
type mockCouponData struct {
	ID                    string
	CouponCode            string
	ExpiryDate            time.Time
	UsageType             string
	MinOrderValue         float64
	ValidTimeWindowStart  *time.Time
	ValidTimeWindowEnd    *time.Time
	TermsAndConditions    string
	DiscountType          string
	DiscountValue         float64
	MaxUsagePerUser       int
	MaxTotalUsage         int
	CurrentTotalUsage     int
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ApplicableMedicineIDs []string
	ApplicableCategories  []string
}

// seedDatabase populates the database with mock data.
func seedDatabase(db database.CouponStorage) error {
	ctx := context.Background()
	log.Println("Starting database seeding...")

	mockCouponsData := []mockCouponData{
		{
			ID:                 uuid.New().String(),
			CouponCode:         "WELCOME10",
			ExpiryDate:         time.Now().AddDate(0, 6, 0), // Expires in 6 months
			UsageType:          "one_time",
			MinOrderValue:      50.00,
			TermsAndConditions: "Valid for new users on their first order.",
			DiscountType:       "percentage",
			DiscountValue:      10.00,
			MaxUsagePerUser:    1,
			MaxTotalUsage:      1000,
			CurrentTotalUsage:  0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New().String(),
			CouponCode:         "FLAT20",
			ExpiryDate:         time.Now().AddDate(1, 0, 0), // Expires in 1 year
			UsageType:          "multi_use",
			MinOrderValue:      100.00,
			TermsAndConditions: "Get a flat 20 INR discount on orders above 100 INR.",
			DiscountType:       "fixed_amount",
			DiscountValue:      20.00,
			MaxUsagePerUser:    0, // Unlimited per user
			MaxTotalUsage:      0, // Unlimited total usage
			CurrentTotalUsage:  0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                   uuid.New().String(),
			CouponCode:           "CATEGORY50",
			ExpiryDate:           time.Now().AddDate(0, 3, 0), // Expires in 3 months
			UsageType:            "multi_use",
			ApplicableCategories: []string{"Painkillers", "Vitamins"},
			MinOrderValue:        75.00,
			TermsAndConditions:   "Valid on Painkillers and Vitamins categories.",
			DiscountType:         "percentage",
			DiscountValue:        15.00,
			MaxUsagePerUser:      5,
			MaxTotalUsage:        500,
			CurrentTotalUsage:    0,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ID:                    uuid.New().String(),
			CouponCode:            "MEDBUY",
			ExpiryDate:            time.Now().AddDate(0, 9, 0), // Expires in 9 months
			UsageType:             "multi_use",
			ApplicableMedicineIDs: []string{"med1", "med3", "med5"},
			MinOrderValue:         150.00,
			TermsAndConditions:    "Valid on specific medicines.",
			DiscountType:          "fixed_amount",
			DiscountValue:         30.00,
			MaxUsagePerUser:       2,
			MaxTotalUsage:         200,
			CurrentTotalUsage:     0,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		},
		{
			ID:                   uuid.New().String(),
			CouponCode:           "TIMEBOUND",
			ExpiryDate:           time.Now().AddDate(0, 1, 0), // Expires in 1 month
			UsageType:            "time_based",
			ValidTimeWindowStart: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),     // Starts tomorrow
			ValidTimeWindowEnd:   func() *time.Time { t := time.Now().Add(7 * 24 * time.Hour); return &t }(), // Ends in a week
			MinOrderValue:        200.00,
			TermsAndConditions:   "Valid only during a specific time window.",
			DiscountType:         "percentage",
			DiscountValue:        25.00,
			MaxUsagePerUser:      0,
			MaxTotalUsage:        0,
			CurrentTotalUsage:    0,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
	}

	for _, mockCoupon := range mockCouponsData {
		coupon := models.Coupon{
			ID:                   mockCoupon.ID,
			CouponCode:           mockCoupon.CouponCode,
			ExpiryDate:           mockCoupon.ExpiryDate,
			UsageType:            mockCoupon.UsageType,
			MinOrderValue:        mockCoupon.MinOrderValue,
			ValidTimeWindowStart: mockCoupon.ValidTimeWindowStart,
			ValidTimeWindowEnd:   mockCoupon.ValidTimeWindowEnd,
			TermsAndConditions:   mockCoupon.TermsAndConditions,
			DiscountType:         mockCoupon.DiscountType,
			DiscountValue:        mockCoupon.DiscountValue,
			MaxUsagePerUser:      mockCoupon.MaxUsagePerUser,
			MaxTotalUsage:        mockCoupon.MaxTotalUsage,
			CurrentTotalUsage:    mockCoupon.CurrentTotalUsage,
			CreatedAt:            mockCoupon.CreatedAt,
			UpdatedAt:            mockCoupon.UpdatedAt,
		}
		for _, id := range mockCoupon.ApplicableMedicineIDs {
			coupon.MedicineIDs = append(coupon.MedicineIDs, models.Medicine{ID: id})
		}
		for _, category := range mockCoupon.ApplicableCategories {
			coupon.Categories = append(coupon.Categories, models.Category{ID: category})
		}

		err := db.CreateCoupon(ctx, &coupon)
		if err != nil {
			return errors.New("failed to create coupon: " + err.Error())
		}
		log.Printf("Seeded coupon: %s\n", coupon.CouponCode)
	}

	return nil
}
