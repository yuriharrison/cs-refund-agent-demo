package agent

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"gorm.io/gorm"
)

func BuildTools(db *gorm.DB, customerID uint, eventBus *chat.EventBus) ([]tool.BaseTool, error) {
	lookupOrders, err := NewLookupOrdersTool(db, customerID)
	if err != nil {
		return nil, err
	}

	getPolicy, err := NewGetRefundPolicyTool(db)
	if err != nil {
		return nil, err
	}

	issueRefund, err := NewIssueRefundTool(db, eventBus)
	if err != nil {
		return nil, err
	}

	escalate, err := NewEscalateToHumanTool(eventBus)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		lookupOrders,
		getPolicy,
		issueRefund,
		escalate,
	}, nil
}
