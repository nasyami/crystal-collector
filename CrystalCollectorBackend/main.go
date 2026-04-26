package main

import (
	"log"
	"net/http"

	"game-backend/handlers"
	"game-backend/store"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	mockStore := store.NewMockStore()
	api := handlers.NewAPI(mockStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.Health)
	mux.HandleFunc("GET /v1/items", api.GetItems)
	mux.HandleFunc("GET /v1/me", api.GetMe)
	mux.HandleFunc("GET /v1/me/items", api.GetMyItems)
	mux.HandleFunc("POST /v1/payments/token", api.CreatePaymentToken)
	mux.HandleFunc("POST /v1/webhooks/xsolla", api.XsollaWebhook)

	handler := handlers.CORS(mux)

	addr := ":8081"
	log.Printf("game-backend listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
