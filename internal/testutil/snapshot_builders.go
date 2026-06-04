package testutil

import (
	"encoding/json"
	"fmt"
)

const snapshotModel = "deepseek/deepseek-v4-flash"

func emptyRequest() json.RawMessage {
	return json.RawMessage(`{}`)
}

func contentResponse(content string) json.RawMessage {
	payload := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%s", hashContent(content)),
		"object":  "chat.completion",
		"created": 1717430400,
		"model":   snapshotModel,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]int{"prompt_tokens": 500, "completion_tokens": 50, "total_tokens": 550},
	}
	return mustMarshal(payload)
}

func toolCallResponse(callID, toolName, args string) json.RawMessage {
	payload := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%s", callID),
		"object":  "chat.completion",
		"created": 1717430400,
		"model":   snapshotModel,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": nil,
					"tool_calls": []map[string]interface{}{
						{
							"id":   callID,
							"type": "function",
							"function": map[string]string{
								"name":      toolName,
								"arguments": args,
							},
						},
					},
				},
				"finish_reason": "tool_calls",
			},
		},
		"usage": map[string]int{"prompt_tokens": 600, "completion_tokens": 40, "total_tokens": 640},
	}
	return mustMarshal(payload)
}

func pairContent(content string) RequestResponsePair {
	return RequestResponsePair{
		Request:  emptyRequest(),
		Response: contentResponse(content),
		Status:   200,
	}
}

func pairToolCall(callID, toolName, args string) RequestResponsePair {
	return RequestResponsePair{
		Request:  emptyRequest(),
		Response: toolCallResponse(callID, toolName, args),
		Status:   200,
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func hashContent(s string) string {
	if len(s) > 8 {
		return fmt.Sprintf("%x", s[:8])
	}
	return fmt.Sprintf("%x", s)
}
