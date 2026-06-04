package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/gorm"
)

type IssueRefundInput struct {
	OrderItemID uint    `json:"order_item_id" jsonschema:"description=ID of the order item to refund,required"`
	RefundType  string  `json:"refund_type" jsonschema:"description=Refund type: full or partial,required"`
	Amount      float64 `json:"amount" jsonschema:"description=Refund amount in dollars,required"`
	Reason      string  `json:"reason" jsonschema:"description=Reason for the refund,required"`
}

type IssueRefundResult struct {
	RefundID uint    `json:"refund_id"`
	Status   string  `json:"status"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
}

func NewIssueRefundTool(db *gorm.DB, eventBus *chat.EventBus) (tool.InvokableTool, error) {
	return utils.InferTool(
		"issue_refund",
		"Issues an approved refund for a specific order item.",
		func(ctx context.Context, input *IssueRefundInput) (string, error) {
			refund := domain.Refund{
				OrderItemID: input.OrderItemID,
				Status:      domain.RefundStatusApproved,
				Type:        domain.RefundType(input.RefundType),
				Amount:      input.Amount,
				Reason:      input.Reason,
				DecidedBy:   domain.RefundDecidedByAgent,
			}

			if err := db.Create(&refund).Error; err != nil {
				return "", fmt.Errorf("creating refund: %w", err)
			}

			sessionID := SessionIDFromContext(ctx)
			if sessionID != "" {
				eventBus.Publish(chat.Event{
					Type:      chat.EventSystemConfirmation,
					SessionID: sessionID,
					Data: map[string]interface{}{
						"action": "refund_issued",
						"details": map[string]interface{}{
							"refund_id":     refund.ID,
							"amount":        refund.Amount,
							"type":          string(refund.Type),
							"order_item_id": refund.OrderItemID,
						},
					},
				})
			}

			result := IssueRefundResult{
				RefundID: refund.ID,
				Status:   string(refund.Status),
				Amount:   refund.Amount,
				Type:     string(refund.Type),
			}

			out, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("marshaling result: %w", err)
			}
			return string(out), nil
		},
	)
}
