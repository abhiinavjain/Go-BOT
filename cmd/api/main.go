package main

import (
	"context"
	"customer-support-bot/internal/handler"
	"customer-support-bot/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("api_key")
	ctx := context.Background()
	if apiKey == "" {
		log.Fatal("API key is not present")
	}

	dbURL := os.Getenv("dbConstr")

	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	geminSvc, err := service.BotService(ctx, apiKey, dbURL)

	if err != nil {
		log.Fatal(err)
	}

	chathandler := handler.NewChatHandler(geminSvc)

	r := chi.NewRouter()

	r.Post("/api/v1/session/{sessionID}", chathandler.HandleChat)
	r.Post("/api/v1/session", chathandler.HandleCreateSession)

	fmt.Println("Running server on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))

}
