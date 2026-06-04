package db

import (
	"log/slog"
	"time"

	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	var count int64
	db.Model(&domain.Customer{}).Count(&count)
	if count > 0 {
		slog.Info("database already seeded, skipping")
		return nil
	}

	slog.Info("seeding database")

	if err := seedCustomers(db); err != nil {
		return err
	}
	if err := seedProducts(db); err != nil {
		return err
	}
	if err := seedOrders(db); err != nil {
		return err
	}
	if err := seedRefundPolicies(db); err != nil {
		return err
	}

	slog.Info("database seeding complete")
	return nil
}

func seedCustomers(db *gorm.DB) error {
	customers := []domain.Customer{
		{ID: 1, Name: "Sarah Mitchell", Email: "sarah.mitchell@email.com", Phone: "(555) 123-4567"},
	}
	return db.Create(&customers).Error
}

func seedProducts(db *gorm.DB) error {
	products := []domain.Product{
		{ID: 1, Name: "Wireless Noise-Cancelling Headphones", Type: domain.ProductTypeElectronics, Price: 149.99},
		{ID: 2, Name: "Premium Cotton T-Shirt", Type: domain.ProductTypeClothing, Price: 34.99},
		{ID: 3, Name: "Organic Meal Kit Box", Type: domain.ProductTypeFood, Price: 59.99},
		{ID: 4, Name: "ProEdit Photo Suite License", Type: domain.ProductTypeSoftware, Price: 199.99},
		{ID: 5, Name: "CloudSync Pro (Annual)", Type: domain.ProductTypeSubscription, Price: 119.99},
		{ID: 6, Name: "Bluetooth Keyboard", Type: domain.ProductTypeElectronics, Price: 79.99},
		{ID: 7, Name: "Running Shoes", Type: domain.ProductTypeClothing, Price: 89.99},
	}
	return db.Create(&products).Error
}

func seedOrders(db *gorm.DB) error {
	now := time.Now()

	orders := []domain.Order{
		{ID: 101, CustomerID: 1, CreatedAt: now.AddDate(0, 0, -5), Status: domain.OrderStatusCompleted},
		{ID: 102, CustomerID: 1, CreatedAt: now.AddDate(0, 0, -20), Status: domain.OrderStatusCompleted},
		{ID: 103, CustomerID: 1, CreatedAt: now.AddDate(0, 0, -3), Status: domain.OrderStatusCompleted},
		{ID: 104, CustomerID: 1, CreatedAt: now.AddDate(0, 0, -2), Status: domain.OrderStatusCompleted},
		{ID: 105, CustomerID: 1, CreatedAt: now.AddDate(0, 0, -10), Status: domain.OrderStatusCompleted},
	}
	if err := db.Create(&orders).Error; err != nil {
		return err
	}

	items := []domain.OrderItem{
		// Order 101: Wireless Headphones ×1, Cotton T-Shirt ×2
		{OrderID: 101, ProductID: 1, Quantity: 1, UnitPrice: 149.99, TotalPrice: 149.99},
		{OrderID: 101, ProductID: 2, Quantity: 2, UnitPrice: 34.99, TotalPrice: 69.98},
		// Order 102: Organic Meal Kit ×1
		{OrderID: 102, ProductID: 3, Quantity: 1, UnitPrice: 59.99, TotalPrice: 59.99},
		// Order 103: ProEdit Photo Suite ×1
		{OrderID: 103, ProductID: 4, Quantity: 1, UnitPrice: 199.99, TotalPrice: 199.99},
		// Order 104: CloudSync Pro ×1
		{OrderID: 104, ProductID: 5, Quantity: 1, UnitPrice: 119.99, TotalPrice: 119.99},
		// Order 105: Bluetooth Keyboard ×1, Running Shoes ×1
		{OrderID: 105, ProductID: 6, Quantity: 1, UnitPrice: 79.99, TotalPrice: 79.99},
		{OrderID: 105, ProductID: 7, Quantity: 1, UnitPrice: 89.99, TotalPrice: 89.99},
	}
	return db.Create(&items).Error
}

func seedRefundPolicies(db *gorm.DB) error {
	intPtr := func(v int) *int { return &v }

	policies := []domain.RefundPolicy{
		// Electronics
		{ProductType: domain.ProductTypeElectronics, Condition: domain.RefundConditionDefective, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(30), Notes: "Full refund for defective electronics"},
		{ProductType: domain.ProductTypeElectronics, Condition: domain.RefundConditionWrongItem, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(30), Notes: "Full refund for wrong item shipped"},
		{ProductType: domain.ProductTypeElectronics, Condition: domain.RefundConditionNotAsDescribed, Action: domain.PolicyActionPartialRefund, PartialPercent: intPtr(80), WindowDays: intPtr(30), Notes: "80% refund if item doesn't match description"},
		{ProductType: domain.ProductTypeElectronics, Condition: domain.RefundConditionChangeOfMind, Action: domain.PolicyActionPartialRefund, PartialPercent: intPtr(70), WindowDays: intPtr(15), Notes: "70% refund within 15-day window"},
		// Clothing
		{ProductType: domain.ProductTypeClothing, Condition: domain.RefundConditionDefective, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(60), Notes: "Full refund for defective clothing"},
		{ProductType: domain.ProductTypeClothing, Condition: domain.RefundConditionWrongItem, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(60), Notes: "Full refund for wrong size/color shipped"},
		{ProductType: domain.ProductTypeClothing, Condition: domain.RefundConditionNotAsDescribed, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(30), Notes: "Full refund within 30 days"},
		{ProductType: domain.ProductTypeClothing, Condition: domain.RefundConditionChangeOfMind, Action: domain.PolicyActionNoRefund, Notes: "No refund for change of mind (hygiene policy)"},
		// Food
		{ProductType: domain.ProductTypeFood, Condition: domain.RefundConditionDefective, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(7), Notes: "Full refund for spoiled/expired food"},
		{ProductType: domain.ProductTypeFood, Condition: domain.RefundConditionWrongItem, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(7), Notes: "Full refund for wrong item"},
		{ProductType: domain.ProductTypeFood, Condition: domain.RefundConditionAny, Action: domain.PolicyActionNoRefund, Notes: "Food items are non-refundable otherwise"},
		// Software
		{ProductType: domain.ProductTypeSoftware, Condition: domain.RefundConditionAny, Action: domain.PolicyActionEscalate, Notes: "All software refunds require human review"},
		// Subscription
		{ProductType: domain.ProductTypeSubscription, Condition: domain.RefundConditionSubscriptionCancel, Action: domain.PolicyActionPartialRefund, PartialPercent: intPtr(50), WindowDays: intPtr(3), Notes: "Pro-rated 50% refund within 3 days of renewal"},
		{ProductType: domain.ProductTypeSubscription, Condition: domain.RefundConditionChangeOfMind, Action: domain.PolicyActionFullRefund, WindowDays: intPtr(7), Notes: "Full refund within 7-day trial window"},
		{ProductType: domain.ProductTypeSubscription, Condition: domain.RefundConditionAny, Action: domain.PolicyActionNoRefund, Notes: "No refund outside trial/renewal windows"},
	}
	return db.Create(&policies).Error
}
