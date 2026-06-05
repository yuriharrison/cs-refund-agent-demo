package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yuriharrison/empirical-proj/internal/chat"
)

// errSessionEscalated matches the sentinel from the chat handler.
// Duplicated here to avoid a circular import with the api package.
var errSessionEscalated = errors.New("session already escalated to human agent")

const stepDelay = 2 * time.Second

type ChatSender interface {
	SendMessageInternal(ctx context.Context, sessionID string, content string, headers http.Header) error
}

type Runner struct {
	chatSender ChatSender
	eventBus   *chat.EventBus
}

func NewRunner(chatSender ChatSender, eventBus *chat.EventBus) *Runner {
	return &Runner{
		chatSender: chatSender,
		eventBus:   eventBus,
	}
}

func (r *Runner) Run(ctx context.Context, sessionID string, uc *Usecase) error {
	slog.Info("usecase started", "usecase_id", uc.ID, "session_id", sessionID, "steps", uc.StepCount)

	r.publishDemoEvent(sessionID, "demo_started", map[string]string{
		"usecase_id":   uc.ID,
		"usecase_name": uc.Name,
	})

	for i, step := range uc.Steps {
		select {
		case <-ctx.Done():
			slog.Info("usecase cancelled", "usecase_id", uc.ID, "step", i+1)
			r.publishDemoEvent(sessionID, "demo_ended", map[string]string{
				"usecase_id": uc.ID,
				"reason":     "cancelled",
			})
			return ctx.Err()
		default:
		}

		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(stepDelay):
			}
		}

		hdr := make(http.Header)
		for k, v := range step.Headers {
			hdr.Set(k, v)
		}

		r.eventBus.Publish(chat.Event{
			Type:      chat.EventDemoCustomerMsg,
			SessionID: sessionID,
			Data:      map[string]string{"content": step.Content},
		})

		slog.Info("usecase step", "usecase_id", uc.ID, "step", i+1, "content", step.Content)
		if err := r.chatSender.SendMessageInternal(ctx, sessionID, step.Content, hdr); err != nil {
			if err.Error() == errSessionEscalated.Error() {
				slog.Info("usecase stopped: session escalated", "usecase_id", uc.ID, "step", i+1)
				r.publishDemoEvent(sessionID, "demo_ended", map[string]string{
					"usecase_id": uc.ID,
					"reason":     "escalated",
				})
				return nil
			}
			slog.Error("usecase step failed", "usecase_id", uc.ID, "step", i+1, "error", err)
			r.publishDemoEvent(sessionID, "demo_ended", map[string]string{
				"usecase_id": uc.ID,
				"reason":     fmt.Sprintf("step %d failed: %v", i+1, err),
			})
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}

	slog.Info("usecase completed", "usecase_id", uc.ID, "session_id", sessionID)
	r.publishDemoEvent(sessionID, "demo_ended", map[string]string{
		"usecase_id": uc.ID,
		"reason":     "completed",
	})

	return nil
}

func (r *Runner) publishDemoEvent(sessionID string, eventType chat.EventType, data interface{}) {
	r.eventBus.Publish(chat.Event{
		Type:      eventType,
		SessionID: sessionID,
		Data:      data,
	})
}
