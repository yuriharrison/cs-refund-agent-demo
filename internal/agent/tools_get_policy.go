package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/gorm"
)

type GetRefundPolicyInput struct {
	ProductType string `json:"product_type" jsonschema:"description=Product category (electronics, clothing, food, software, subscription),required"`
	Condition   string `json:"condition" jsonschema:"description=Refund condition (defective, wrong_item, not_as_described, change_of_mind, subscription_cancel, any),required"`
}

type PolicyResult struct {
	Action         string `json:"action"`
	PartialPercent *int   `json:"partial_percent,omitempty"`
	WindowDays     *int   `json:"window_days,omitempty"`
	Notes          string `json:"notes"`
}

func NewGetRefundPolicyTool(db *gorm.DB) (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_refund_policy",
		"Looks up the refund policy for a product type and condition combination.",
		func(ctx context.Context, input *GetRefundPolicyInput) (string, error) {
			if ForceErrorFromContext(ctx) {
				out, err := json.Marshal(map[string]string{
					"error": "internal error: policy service unavailable",
				})
				if err != nil {
					return "", fmt.Errorf("marshaling error result: %w", err)
				}
				return string(out), nil
			}

			var policy domain.RefundPolicy
			err := db.Where("product_type = ? AND condition IN (?, ?)",
				domain.ProductType(input.ProductType),
				domain.RefundCondition(input.Condition),
				domain.RefundConditionAny,
			).Order("condition DESC").Limit(1).First(&policy).Error

			var result PolicyResult
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					result = PolicyResult{
						Action: string(domain.PolicyActionEscalate),
						Notes:  "No policy found for this combination",
					}
				} else {
					return "", fmt.Errorf("querying refund policy: %w", err)
				}
			} else {
				result = PolicyResult{
					Action:         string(policy.Action),
					PartialPercent: policy.PartialPercent,
					WindowDays:     policy.WindowDays,
					Notes:          policy.Notes,
				}
			}

			out, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("marshaling result: %w", err)
			}
			return string(out), nil
		},
	)
}
