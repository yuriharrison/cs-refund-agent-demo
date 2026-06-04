package agent_test

import (
	"encoding/json"
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/db"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newSeededTestDB(t *testing.T) *gorm.DB {
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
		t.Fatalf("failed to migrate: %v", err)
	}

	if err := db.Seed(database); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	return database
}

func TestLookupOrdersTool_ReturnsCorrectJSON(t *testing.T) {
	database := newSeededTestDB(t)

	tool, err := agent.NewLookupOrdersTool(database, 1)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	result, err := tool.InvokableRun(t.Context(), `{"limit": 10}`)
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var orders []agent.OrderResult
	if err := json.Unmarshal([]byte(result), &orders); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(orders) != 5 {
		t.Errorf("expected 5 orders, got %d", len(orders))
	}

	// Orders should be sorted by created_at DESC
	// Most recent first: order 104 (2 days ago), 103 (3 days), 101 (5 days), 105 (10 days), 102 (20 days)
	expectedIDs := []uint{104, 103, 101, 105, 102}
	for i, order := range orders {
		if order.ID != expectedIDs[i] {
			t.Errorf("order[%d]: expected ID %d, got %d", i, expectedIDs[i], order.ID)
		}
	}

	// Check that items have product names populated
	for _, order := range orders {
		for _, item := range order.Items {
			if item.ProductName == "" {
				t.Errorf("order %d: item has empty product name", order.ID)
			}
			if item.ProductType == "" {
				t.Errorf("order %d: item has empty product type", order.ID)
			}
		}
	}
}

func TestLookupOrdersTool_LimitParameter(t *testing.T) {
	database := newSeededTestDB(t)

	tool, err := agent.NewLookupOrdersTool(database, 1)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	result, err := tool.InvokableRun(t.Context(), `{"limit": 2}`)
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var orders []agent.OrderResult
	if err := json.Unmarshal([]byte(result), &orders); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("expected 2 orders with limit=2, got %d", len(orders))
	}
}

func TestLookupOrdersTool_DefaultLimit(t *testing.T) {
	database := newSeededTestDB(t)

	tool, err := agent.NewLookupOrdersTool(database, 1)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// 0 should use default limit of 5
	result, err := tool.InvokableRun(t.Context(), `{}`)
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var orders []agent.OrderResult
	if err := json.Unmarshal([]byte(result), &orders); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(orders) != 5 {
		t.Errorf("expected 5 orders with default limit, got %d", len(orders))
	}
}

func TestLookupOrdersTool_ItemDetails(t *testing.T) {
	database := newSeededTestDB(t)

	tool, err := agent.NewLookupOrdersTool(database, 1)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	result, err := tool.InvokableRun(t.Context(), `{"limit": 10}`)
	if err != nil {
		t.Fatalf("tool invocation failed: %v", err)
	}

	var orders []agent.OrderResult
	if err := json.Unmarshal([]byte(result), &orders); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Find order 101 which should have 2 items
	var order101 *agent.OrderResult
	for _, o := range orders {
		if o.ID == 101 {
			order101 = &o
			break
		}
	}

	if order101 == nil {
		t.Fatal("order 101 not found")
	}

	if len(order101.Items) != 2 {
		t.Fatalf("order 101: expected 2 items, got %d", len(order101.Items))
	}

	// Verify item details
	foundHeadphones := false
	foundTShirt := false
	for _, item := range order101.Items {
		switch item.ProductName {
		case "Wireless Noise-Cancelling Headphones":
			foundHeadphones = true
			if item.Quantity != 1 {
				t.Errorf("headphones: expected quantity 1, got %d", item.Quantity)
			}
			if item.UnitPrice != 149.99 {
				t.Errorf("headphones: expected unit price 149.99, got %.2f", item.UnitPrice)
			}
		case "Premium Cotton T-Shirt":
			foundTShirt = true
			if item.Quantity != 2 {
				t.Errorf("t-shirt: expected quantity 2, got %d", item.Quantity)
			}
		}
	}

	if !foundHeadphones {
		t.Error("order 101: missing headphones item")
	}
	if !foundTShirt {
		t.Error("order 101: missing t-shirt item")
	}
}
