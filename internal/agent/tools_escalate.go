package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

const humanAgentName = "Alex"

const humanAgentMessage = "Hi, I'm Alex from the support team. I've reviewed your case and I'm here to help. Let me look into this for you."

type EscalateToHumanInput struct {
	Reason string `json:"reason" jsonschema:"description=Reason for escalating to a human agent,required"`
}

type EscalateResult struct {
	Escalated bool   `json:"escalated"`
	Reason    string `json:"reason"`
}

func NewEscalateToHumanTool(eventBus *chat.EventBus) (tool.InvokableTool, error) {
	return utils.InferTool(
		"escalate_to_human",
		"Escalates the conversation to a human support agent.",
		func(ctx context.Context, input *EscalateToHumanInput) (string, error) {
			sessionID := SessionIDFromContext(ctx)

			if sessionID != "" {
				eventBus.Publish(chat.Event{
					Type:      chat.EventSystemEscalation,
					SessionID: sessionID,
					Data: map[string]string{
						"reason":           input.Reason,
						"human_agent_name": humanAgentName,
					},
				})

				time.Sleep(2 * time.Second)

				eventBus.Publish(chat.Event{
					Type:      chat.EventHumanMessage,
					SessionID: sessionID,
					Data: map[string]string{
						"content":    humanAgentMessage,
						"agent_name": humanAgentName,
					},
				})
			}

			result := EscalateResult{
				Escalated: true,
				Reason:    input.Reason,
			}

			out, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("marshaling result: %w", err)
			}
			return string(out), nil
		},
	)
}
