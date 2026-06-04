package testutil

import (
	"testing"

	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/gorm"
)

func AssertRefundCount(t *testing.T, db *gorm.DB, want int) {
	t.Helper()
	var count int64
	if err := db.Model(&domain.Refund{}).Count(&count).Error; err != nil {
		t.Fatalf("count refunds: %v", err)
	}
	if int(count) != want {
		t.Errorf("expected %d refund(s), got %d", want, count)
	}
}

func AssertLatestRefund(t *testing.T, db *gorm.DB, wantType domain.RefundType, wantAmount float64) {
	t.Helper()
	var refund domain.Refund
	if err := db.Order("id DESC").First(&refund).Error; err != nil {
		t.Fatalf("fetch latest refund: %v", err)
	}
	if refund.Type != wantType {
		t.Errorf("refund type: want %q, got %q", wantType, refund.Type)
	}
	if refund.Amount != wantAmount {
		t.Errorf("refund amount: want %.2f, got %.2f", wantAmount, refund.Amount)
	}
	if refund.Status != domain.RefundStatusApproved {
		t.Errorf("refund status: want %q, got %q", domain.RefundStatusApproved, refund.Status)
	}
}

func CollectEvents(ch <-chan chat.Event) []chat.Event {
	var events []chat.Event
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return events
			}
			events = append(events, ev)
		default:
			return events
		}
	}
}

func DrainEvents(ch <-chan chat.Event) []chat.Event {
	var events []chat.Event
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func HasEventType(events []chat.Event, eventType chat.EventType) bool {
	for _, ev := range events {
		if ev.Type == eventType {
			return true
		}
	}
	return false
}

func HasToolCall(events []chat.Event, toolName string) bool {
	for _, ev := range events {
		if ev.Type != chat.EventToolCallStart {
			continue
		}
		data, ok := ev.Data.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := data["tool"].(string); ok && name == toolName {
			return true
		}
	}
	return false
}

func AssertNoToolCall(t *testing.T, events []chat.Event, toolName string) {
	t.Helper()
	if HasToolCall(events, toolName) {
		t.Errorf("unexpected tool call: %s", toolName)
	}
}
