package chat

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAgent    Role = "agent"
	RoleSystem   Role = "system"
	RoleHuman    Role = "human"
)

type EventType string

const (
	EventAgentThinking      EventType = "agent_thinking"
	EventToolCallStart      EventType = "tool_call_start"
	EventToolCallResult     EventType = "tool_call_result"
	EventAgentMessage       EventType = "agent_message"
	EventSystemConfirmation EventType = "system_confirmation"
	EventSystemEscalation   EventType = "system_escalation"
	EventHumanMessage       EventType = "human_message"
	EventTokenUpdate        EventType = "token_update"
	EventError              EventType = "error"
)

type Message struct {
	ID        string    `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	Type      EventType   `json:"type"`
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data"`
}
