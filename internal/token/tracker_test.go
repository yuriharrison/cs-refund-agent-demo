package token

import (
	"strings"
	"testing"
)

func TestTokenTracker_Record(t *testing.T) {
	tracker := NewTracker()

	tracker.Record(100, 50)
	tracker.Record(200, 75)

	if tracker.PromptTokens != 300 {
		t.Errorf("expected 300 prompt tokens, got %d", tracker.PromptTokens)
	}
	if tracker.CompletionTokens != 125 {
		t.Errorf("expected 125 completion tokens, got %d", tracker.CompletionTokens)
	}
}

func TestTokenTracker_Cost(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(1_000_000, 1_000_000)

	cost := tracker.Cost()
	expected := 0.0
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.4f", expected, cost)
	}
}

func TestTokenTracker_ShutdownReport(t *testing.T) {
	tracker := NewTracker()
	tracker.IncrementSessions()
	tracker.IncrementSessions()
	tracker.IncrementSessions()
	tracker.Record(12450, 3210)

	report := tracker.Report()

	checks := []string{
		"ShopEase Demo — Session Summary",
		"Sessions:          3",
		"Prompt Tokens:     12,450",
		"Completion Tokens: 3,210",
		"Total Tokens:      15,660",
		"Input Cost:",
		"Output Cost:",
		"Total Cost:",
		"══════════════════════════════════════════",
	}
	for _, check := range checks {
		if !strings.Contains(report, check) {
			t.Errorf("report missing %q\nGot:\n%s", check, report)
		}
	}
}
