package handlers

import (
	"encoding/json"
	"net/http"

	"game-backend/store"
)

type API struct {
	store *store.PostgresStore
}

func NewAPI(store *store.PostgresStore) *API {
	return &API{store: store}
}

func (api *API) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (api *API) GetItems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.store.Items())
}

func (api *API) GetMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.store.User())
}

// GET /v1/me/items
func (api *API) GetMyItems(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	userClaims, err := extractXsollaUserClaims(authHeader)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	items, err := api.store.GetOwnedItems(userClaims.Sub)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"player_id": userClaims.Sub,
		"items":     items,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
