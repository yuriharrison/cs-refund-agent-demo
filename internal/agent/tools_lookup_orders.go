package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/gorm"
)

type LookupOrdersInput struct {
	Limit int `json:"limit" jsonschema:"description=Maximum number of orders to return (default 5)"`
}

type OrderResult struct {
	ID        uint              `json:"id"`
	Date      string            `json:"date"`
	Status    string            `json:"status"`
	Items     []OrderItemResult `json:"items"`
}

type OrderItemResult struct {
	OrderItemID uint    `json:"order_item_id"`
	ProductName string  `json:"product_name"`
	ProductType string  `json:"product_type"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

func NewLookupOrdersTool(db *gorm.DB, customerID uint) (tool.InvokableTool, error) {
	return utils.InferTool(
		"lookup_customer_orders",
		"Retrieves the current customer's recent orders with item details.",
		func(ctx context.Context, input *LookupOrdersInput) (string, error) {
			limit := input.Limit
			if limit <= 0 {
				limit = 5
			}

			var orders []domain.Order
			err := db.Where("customer_id = ?", customerID).
				Preload("Items.Product").
				Order("created_at DESC").
				Limit(limit).
				Find(&orders).Error
			if err != nil {
				return "", fmt.Errorf("querying orders: %w", err)
			}

			results := make([]OrderResult, len(orders))
			for i, order := range orders {
				items := make([]OrderItemResult, len(order.Items))
				for j, item := range order.Items {
					items[j] = OrderItemResult{
						OrderItemID: item.ID,
						ProductName: item.Product.Name,
						ProductType: string(item.Product.Type),
						Quantity:    item.Quantity,
						UnitPrice:   item.UnitPrice,
						TotalPrice:  item.TotalPrice,
					}
				}
				results[i] = OrderResult{
					ID:     order.ID,
					Date:   order.CreatedAt.Format(time.RFC3339),
					Status: string(order.Status),
					Items:  items,
				}
			}

			out, err := json.Marshal(results)
			if err != nil {
				return "", fmt.Errorf("marshaling result: %w", err)
			}
			return string(out), nil
		},
	)
}
