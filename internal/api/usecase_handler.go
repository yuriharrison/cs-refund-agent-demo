package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yuriharrison/empirical-proj/internal/usecase"
)

type UsecaseHandler struct {
	runner     *usecase.Runner
	chatHandler *ChatHandler
}

func NewUsecaseHandler(runner *usecase.Runner, chatHandler *ChatHandler) *UsecaseHandler {
	return &UsecaseHandler{
		runner:     runner,
		chatHandler: chatHandler,
	}
}

type UsecaseListResponse struct {
	Usecases []UsecaseItem `json:"usecases"`
}

type UsecaseItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Steps       int    `json:"steps"`
}

func (h *UsecaseHandler) List(w http.ResponseWriter, r *http.Request) {
	items := make([]UsecaseItem, len(usecase.Registry))
	for i, uc := range usecase.Registry {
		items[i] = UsecaseItem{
			ID:          uc.ID,
			Name:        uc.Name,
			Description: uc.Description,
			Steps:       uc.StepCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UsecaseListResponse{Usecases: items})
}

type RunUsecaseRequest struct {
	SessionID string `json:"session_id"`
}

type RunUsecaseResponse struct {
	UsecaseID string `json:"usecase_id"`
	SessionID string `json:"session_id"`
}

func (h *UsecaseHandler) Run(w http.ResponseWriter, r *http.Request) {
	usecaseID := chi.URLParam(r, "id")
	uc := usecase.FindByID(usecaseID)
	if uc == nil {
		http.Error(w, `{"error":"usecase not found"}`, http.StatusNotFound)
		return
	}

	var req RunUsecaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	sessionID := h.chatHandler.chatService.GetOrCreateSession(req.SessionID)

	go h.runner.Run(context.Background(), sessionID, uc)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(RunUsecaseResponse{
		UsecaseID: usecaseID,
		SessionID: sessionID,
	})
}
