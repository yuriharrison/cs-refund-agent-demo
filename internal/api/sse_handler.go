package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/yuriharrison/empirical-proj/internal/chat"
)

type SSEHandler struct {
	eventBus *chat.EventBus
}

func NewSSEHandler(eventBus *chat.EventBus) *SSEHandler {
	return &SSEHandler{eventBus: eventBus}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, `{"error":"session_id is required"}`, http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subscriberID := uuid.New().String()
	events := h.eventBus.Subscribe(sessionID, subscriberID)
	defer h.eventBus.Unsubscribe(sessionID, subscriberID)

	slog.Info("SSE client connected", "session_id", sessionID, "subscriber_id", subscriberID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Info("SSE client disconnected", "session_id", sessionID, "subscriber_id", subscriberID)
			return

		case event, ok := <-events:
			if !ok {
				return
			}

			data, err := json.Marshal(event.Data)
			if err != nil {
				slog.Error("failed to marshal SSE event", "error", err)
				continue
			}

			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}
