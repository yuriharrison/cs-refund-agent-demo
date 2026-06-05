package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/yuriharrison/empirical-proj/internal/agent"
	"github.com/yuriharrison/empirical-proj/internal/api"
	"github.com/yuriharrison/empirical-proj/internal/chat"
	"github.com/yuriharrison/empirical-proj/internal/db"
	"github.com/yuriharrison/empirical-proj/internal/domain"
	"github.com/yuriharrison/empirical-proj/internal/token"
	"github.com/yuriharrison/empirical-proj/internal/usecase"
)

func demoForceErrorEnabled() bool {
	v := os.Getenv("DEMO_FORCE_ERROR")
	return v == "1" || v == "true" || v == "TRUE"
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	ctx := context.Background()

	database, err := db.New("file::memory:?cache=shared")
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	if err := db.Seed(database); err != nil {
		slog.Error("failed to seed database", "error", err)
		os.Exit(1)
	}

	var customer domain.Customer
	if err := database.First(&customer).Error; err != nil {
		slog.Error("failed to load customer", "error", err)
		os.Exit(1)
	}

	eventBus := chat.NewEventBus()
	chatService := chat.NewService(eventBus)

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPEN_ROUTER_API_KEY"),
		BaseURL: "https://openrouter.ai/api/v1",
		Model:   "deepseek/deepseek-v4-flash",
	})
	if err != nil {
		slog.Error("failed to create chat model", "error", err)
		os.Exit(1)
	}

	tools, err := agent.BuildTools(database, customer.ID, eventBus)
	if err != nil {
		slog.Error("failed to build tools", "error", err)
		os.Exit(1)
	}

	tokenTracker := token.NewTracker()

	ag, err := agent.New(ctx, agent.Config{
		ChatModel:    chatModel,
		DB:           database,
		Customer:     customer,
		EventBus:     eventBus,
		Tools:        tools,
		TokenTracker: tokenTracker,
	})
	if err != nil {
		slog.Error("failed to create agent", "error", err)
		os.Exit(1)
	}

	chatHandler := api.NewChatHandler(chatService, ag, demoForceErrorEnabled())
	sseHandler := api.NewSSEHandler(eventBus)
	ucRunner := usecase.NewRunner(chatHandler, eventBus)
	usecaseHandler := api.NewUsecaseHandler(ucRunner, chatHandler)

	router := api.NewRouter(&api.Deps{
		ChatHandler:    chatHandler,
		SSEHandler:     sseHandler,
		UsecaseHandler: usecaseHandler,
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced shutdown", "error", err)
		os.Exit(1)
	}

	fmt.Println(tokenTracker.Report())
	slog.Info("server stopped")
}
