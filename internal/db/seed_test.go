package db_test

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/db"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	err = database.AutoMigrate(
		&domain.Customer{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.RefundPolicy{},
		&domain.Refund{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func seededTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database := newTestDB(t)
	if err := db.Seed(database); err != nil {
		t.Fatalf("failed to seed database: %v", err)
	}
	return database
}

func TestSeedData_CustomerCount(t *testing.T) {
	database := seededTestDB(t)
	var count int64
	database.Model(&domain.Customer{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 customer, got %d", count)
	}

	var customer domain.Customer
	database.First(&customer)
	if customer.Name != "Sarah Mitchell" {
		t.Errorf("expected customer name 'Sarah Mitchell', got %q", customer.Name)
	}
	if customer.Email != "sarah.mitchell@email.com" {
		t.Errorf("expected customer email 'sarah.mitchell@email.com', got %q", customer.Email)
	}
}

func TestSeedData_ProductCount(t *testing.T) {
	database := seededTestDB(t)
	var count int64
	database.Model(&domain.Product{}).Count(&count)
	if count != 7 {
		t.Errorf("expected 7 products, got %d", count)
	}

	expectedTypes := map[domain.ProductType]int{
		domain.ProductTypeElectronics:  2,
		domain.ProductTypeClothing:     2,
		domain.ProductTypeFood:         1,
		domain.ProductTypeSoftware:     1,
		domain.ProductTypeSubscription: 1,
	}

	var products []domain.Product
	database.Find(&products)

	typeCounts := make(map[domain.ProductType]int)
	for _, p := range products {
		typeCounts[p.Type]++
	}

	for ptype, expected := range expectedTypes {
		if typeCounts[ptype] != expected {
			t.Errorf("expected %d products of type %q, got %d", expected, ptype, typeCounts[ptype])
		}
	}
}

func TestSeedData_OrderCount(t *testing.T) {
	database := seededTestDB(t)
	var count int64
	database.Model(&domain.Order{}).Count(&count)
	if count != 5 {
		t.Errorf("expected 5 orders, got %d", count)
	}

	var orders []domain.Order
	database.Preload("Items").Find(&orders)

	expectedItemCounts := map[uint]int{
		101: 2,
		102: 1,
		103: 1,
		104: 1,
		105: 2,
	}

	for _, order := range orders {
		expected, ok := expectedItemCounts[order.ID]
		if !ok {
			t.Errorf("unexpected order ID %d", order.ID)
			continue
		}
		if len(order.Items) != expected {
			t.Errorf("order %d: expected %d items, got %d", order.ID, expected, len(order.Items))
		}
	}
}

func TestSeedData_RefundPolicyCount(t *testing.T) {
	database := seededTestDB(t)
	var count int64
	database.Model(&domain.RefundPolicy{}).Count(&count)

	// 4 electronics + 4 clothing + 3 food + 1 software + 3 subscription = 15
	if count != 15 {
		t.Errorf("expected 15 refund policy rows, got %d", count)
	}

	type policyKey struct {
		ProductType domain.ProductType
		Condition   domain.RefundCondition
	}

	requiredPolicies := []policyKey{
		{domain.ProductTypeElectronics, domain.RefundConditionDefective},
		{domain.ProductTypeElectronics, domain.RefundConditionWrongItem},
		{domain.ProductTypeElectronics, domain.RefundConditionNotAsDescribed},
		{domain.ProductTypeElectronics, domain.RefundConditionChangeOfMind},
		{domain.ProductTypeClothing, domain.RefundConditionDefective},
		{domain.ProductTypeClothing, domain.RefundConditionWrongItem},
		{domain.ProductTypeClothing, domain.RefundConditionNotAsDescribed},
		{domain.ProductTypeClothing, domain.RefundConditionChangeOfMind},
		{domain.ProductTypeFood, domain.RefundConditionDefective},
		{domain.ProductTypeFood, domain.RefundConditionWrongItem},
		{domain.ProductTypeFood, domain.RefundConditionAny},
		{domain.ProductTypeSoftware, domain.RefundConditionAny},
		{domain.ProductTypeSubscription, domain.RefundConditionSubscriptionCancel},
		{domain.ProductTypeSubscription, domain.RefundConditionChangeOfMind},
		{domain.ProductTypeSubscription, domain.RefundConditionAny},
	}

	var policies []domain.RefundPolicy
	database.Find(&policies)

	existing := make(map[policyKey]bool)
	for _, p := range policies {
		existing[policyKey{p.ProductType, p.Condition}] = true
	}

	for _, required := range requiredPolicies {
		if !existing[required] {
			t.Errorf("missing refund policy for product_type=%q condition=%q", required.ProductType, required.Condition)
		}
	}
}

func TestSeedData_OrderItemPrices(t *testing.T) {
	database := seededTestDB(t)
	var items []domain.OrderItem
	database.Preload("Product").Find(&items)

	for _, item := range items {
		if item.UnitPrice != item.Product.Price {
			t.Errorf("order item %d: unit_price %.2f != product price %.2f",
				item.ID, item.UnitPrice, item.Product.Price)
		}
		expectedTotal := item.UnitPrice * float64(item.Quantity)
		if item.TotalPrice != expectedTotal {
			t.Errorf("order item %d: total_price %.2f != expected %.2f (qty %d × unit %.2f)",
				item.ID, item.TotalPrice, expectedTotal, item.Quantity, item.UnitPrice)
		}
	}
}

func TestDatabase_SeedIdempotent(t *testing.T) {
	database := newTestDB(t)

	if err := db.Seed(database); err != nil {
		t.Fatalf("first seed failed: %v", err)
	}
	if err := db.Seed(database); err != nil {
		t.Fatalf("second seed failed: %v", err)
	}

	var count int64
	database.Model(&domain.Customer{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 customer after double seed, got %d", count)
	}

	database.Model(&domain.Product{}).Count(&count)
	if count != 7 {
		t.Errorf("expected 7 products after double seed, got %d", count)
	}
}
