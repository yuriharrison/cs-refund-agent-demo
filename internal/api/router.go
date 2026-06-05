package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	ChatHandler    *ChatHandler
	SSEHandler     *SSEHandler
	UsecaseHandler *UsecaseHandler
}

func NewRouter(deps *Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", HealthHandler)

		if deps != nil && deps.ChatHandler != nil {
			r.With(jsonContentType).Post("/chat/message", deps.ChatHandler.SendMessage)
			r.With(jsonContentType).Get("/chat/history", deps.ChatHandler.GetHistory)
			r.With(jsonContentType).Post("/chat/reset", deps.ChatHandler.ResetSession)
		}

		if deps != nil && deps.SSEHandler != nil {
			r.Get("/chat/stream", deps.SSEHandler.Stream)
		}

		if deps != nil && deps.UsecaseHandler != nil {
			r.Get("/usecases", deps.UsecaseHandler.List)
			r.With(jsonContentType).Post("/usecases/{id}/run", deps.UsecaseHandler.Run)
		}
	})

	return r
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
