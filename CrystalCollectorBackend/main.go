package main

import (
	"log"
	"net/http"
	"os"

	"game-backend/handlers"
	"game-backend/store"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	connStr := os.Getenv("DATABASE_URL")
	postgresStore, err := store.NewPostgresStore(connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	if err := store.EnsureSchema(postgresStore.DB); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}
	if err := store.SeedShopItems(postgresStore.DB); err != nil {
		log.Fatalf("failed to seed shop_items: %v", err)
	}
	api := handlers.NewAPI(postgresStore)

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
