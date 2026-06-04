package domain

import "time"

type RefundCondition string

const (
	RefundConditionAny                RefundCondition = "any"
	RefundConditionDefective          RefundCondition = "defective"
	RefundConditionWrongItem          RefundCondition = "wrong_item"
	RefundConditionNotAsDescribed     RefundCondition = "not_as_described"
	RefundConditionChangeOfMind       RefundCondition = "change_of_mind"
	RefundConditionSubscriptionCancel RefundCondition = "subscription_cancel"
)

type PolicyAction string

const (
	PolicyActionFullRefund    PolicyAction = "full_refund"
	PolicyActionPartialRefund PolicyAction = "partial_refund"
	PolicyActionNoRefund      PolicyAction = "no_refund"
	PolicyActionEscalate      PolicyAction = "escalate"
)

type RefundPolicy struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	ProductType    ProductType     `json:"product_type"`
	Condition      RefundCondition `json:"condition"`
	Action         PolicyAction    `json:"action"`
	PartialPercent *int            `json:"partial_percent,omitempty"`
	WindowDays     *int            `json:"window_days,omitempty"`
	Notes          string          `json:"notes"`
}

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusDenied    RefundStatus = "denied"
	RefundStatusEscalated RefundStatus = "escalated"
)

type RefundType string

const (
	RefundTypeFull    RefundType = "full"
	RefundTypePartial RefundType = "partial"
)

type RefundDecidedBy string

const (
	RefundDecidedByAgent RefundDecidedBy = "agent"
	RefundDecidedByHuman RefundDecidedBy = "human"
)

type Refund struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	OrderItemID uint            `json:"order_item_id"`
	OrderItem   OrderItem       `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	Status      RefundStatus    `json:"status"`
	Type        RefundType      `json:"type"`
	Amount      float64         `json:"amount"`
	Reason      string          `json:"reason"`
	DecidedBy   RefundDecidedBy `json:"decided_by"`
	CreatedAt   time.Time       `json:"created_at"`
}
