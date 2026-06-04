package domain

import "time"

type OrderStatus string

const (
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID         uint        `gorm:"primarykey" json:"id"`
	CustomerID uint        `json:"customer_id"`
	Customer   Customer    `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	Status     OrderStatus `json:"status"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID         uint    `gorm:"primarykey" json:"id"`
	OrderID    uint    `json:"order_id"`
	ProductID  uint    `json:"product_id"`
	Product    Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}
