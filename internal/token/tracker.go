package token

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type TokenTracker struct {
	mu               sync.Mutex
	PromptTokens     int64
	CompletionTokens int64
	Sessions         int
}

func NewTracker() *TokenTracker {
	return &TokenTracker{}
}

func (t *TokenTracker) Record(prompt, completion int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.PromptTokens += int64(prompt)
	t.CompletionTokens += int64(completion)
}

func (t *TokenTracker) Cost() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	inputCost := float64(t.PromptTokens) / 1_000_000 * 0.0
	outputCost := float64(t.CompletionTokens) / 1_000_000 * 0.0
	return inputCost + outputCost
}

func (t *TokenTracker) IncrementSessions() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Sessions++
}

func (t *TokenTracker) Report() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	total := t.PromptTokens + t.CompletionTokens
	inputCost := float64(t.PromptTokens) / 1_000_000 * 0.0
	outputCost := float64(t.CompletionTokens) / 1_000_000 * 0.0
	totalCost := inputCost + outputCost

	return fmt.Sprintf(`══════════════════════════════════════════
  ShopEase Demo — Session Summary
──────────────────────────────────────────
  Sessions:          %d
  Prompt Tokens:     %s
  Completion Tokens: %s
  Total Tokens:      %s
──────────────────────────────────────────
  Input Cost:        $%.4f
  Output Cost:       $%.4f
  Total Cost:        $%.4f
══════════════════════════════════════════`,
		t.Sessions,
		formatInt(t.PromptTokens),
		formatInt(t.CompletionTokens),
		formatInt(total),
		inputCost,
		outputCost,
		totalCost,
	)
}

func formatInt(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}

	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, ",")
}
