package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"game-backend/models"
	"game-backend/store"
)

type API struct {
	store *store.PostgresStore
}

type createPaymentTokenRequest struct {
	ItemID string `json:"item_id"`
}

type payment struct {
	PaymentID string `json:"payment_id"`
	ItemID    string `json:"item_id"`
	Status    string `json:"status"`
}

type paymentTokenResponse struct {
	PaymentID     string `json:"payment_id"`
	Token         string `json:"token"`
	PayStationURL string `json:"pay_station_url"`
}

type xsollaConfig struct {
	ProjectID string
	APIKey    string
	ReturnURL string
}

type missingEnvError struct {
	Names []string
}

type xsollaPaymentTokenRequest struct {
	User struct {
		ID struct {
			Value string `json:"value"`
		} `json:"id"`
		Name struct {
			Value string `json:"value"`
		} `json:"name"`
		Email struct {
			Value string `json:"value"`
		} `json:"email"`
		Country struct {
			Value       string `json:"value"`
			AllowModify bool   `json:"allow_modify"`
		} `json:"country"`
	} `json:"user"`
	Purchase struct {
		Items []struct {
			SKU      string `json:"sku"`
			Quantity int    `json:"quantity"`
		} `json:"items"`
	} `json:"purchase"`
	Sandbox  bool `json:"sandbox"`
	Settings struct {
		Language  string `json:"language"`
		Currency  string `json:"currency"`
		ReturnURL string `json:"return_url"`
	} `json:"settings"`
}

type xsollaPaymentTokenResult struct {
	Token   string `json:"token"`
	OrderID int64  `json:"order_id"`
}

type xsollaUserClaims struct {
	Sub      string
	Email    string
	Username string
}

type xsollaWebhookEvent struct {
	NotificationType string `json:"notification_type"`
	User             struct {
		ID struct {
			Value string `json:"value"`
		} `json:"id"`
	} `json:"user"`
	Purchase struct {
		Items []struct {
			SKU string `json:"sku"`
		} `json:"items"`
		VirtualItems struct {
			Items []struct {
				SKU string `json:"sku"`
			} `json:"items"`
		} `json:"virtual_items"`
	} `json:"purchase"`
	Order struct {
		Status string `json:"status"`
	} `json:"order"`
	Payment struct {
		Status string `json:"status"`
	} `json:"payment"`
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

// POST /v1/payments/token
func (api *API) CreatePaymentToken(w http.ResponseWriter, r *http.Request) {
	var req createPaymentTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	authHeader := r.Header.Get("Authorization")
	log.Printf("xsolla auth header exists: %t", authHeader != "")

	userClaims, err := extractXsollaUserClaims(authHeader)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	log.Printf("xsolla extracted sub: %s", userClaims.Sub)
	log.Printf("xsolla extracted email: %s", userClaims.Email)
	log.Printf("create payment token request item_id: %s", req.ItemID)

	item, ok := api.findItemByID(req.ItemID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	log.Printf("matched server-side item: %+v", item)

	pendingPayment := api.newPendingPayment(req.ItemID)

	resp, err := api.buildPaymentTokenResponse(pendingPayment, item, userClaims)
	if err != nil {
		var missingErr missingEnvError
		if errors.As(err, &missingErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":            "missing required environment variables",
				"missing_env_vars": missingErr.Names,
			})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (api *API) findItemByID(itemID string) (models.Item, bool) {
	for _, item := range api.store.Items() {
		if item.ID == itemID {
			return item, true
		}
	}
	return models.Item{}, false
}

func (api *API) newPendingPayment(itemID string) payment {
	return payment{
		PaymentID: fmt.Sprintf("pay_%d", time.Now().UnixNano()),
		ItemID:    itemID,
		Status:    "pending",
	}
}

func (api *API) buildPaymentTokenResponse(pendingPayment payment, item models.Item, userClaims xsollaUserClaims) (paymentTokenResponse, error) {
	cfg, err := loadXsollaConfig()
	if err != nil {
		return paymentTokenResponse{}, err
	}

	xsollaToken, err := api.createXsollaPaymentToken(cfg, item, userClaims)
	if err != nil {
		return paymentTokenResponse{}, err
	}

	paymentID := pendingPayment.PaymentID
	if xsollaToken.OrderID != 0 {
		paymentID = fmt.Sprintf("%d", xsollaToken.OrderID)
	}

	return paymentTokenResponse{
		PaymentID:     paymentID,
		Token:         xsollaToken.Token,
		PayStationURL: "https://sandbox-secure.xsolla.com/paystation4/?token=" + xsollaToken.Token,
	}, nil
}

func loadXsollaConfig() (xsollaConfig, error) {
	cfg := xsollaConfig{
		ProjectID: os.Getenv("XSOLLA_PROJECT_ID"),
		APIKey:    os.Getenv("XSOLLA_API_KEY"),
		ReturnURL: getReturnURL(),
	}
	var missing []string
	if cfg.ProjectID == "" {
		missing = append(missing, "XSOLLA_PROJECT_ID")
	}
	if cfg.APIKey == "" {
		missing = append(missing, "XSOLLA_API_KEY")
	}
	if len(missing) > 0 {
		return xsollaConfig{}, missingEnvError{Names: missing}
	}
	return cfg, nil
}

func getReturnURL() string {
	url := os.Getenv("XSOLLA_RETURN_URL")
	if url == "" {
		return "https://crystal-collector-ten.vercel.app/shop"
	}
	return url
}

func (e missingEnvError) Error() string {
	return "missing required environment variables: " + strings.Join(e.Names, ", ")
}

func extractXsollaUserClaims(authHeader string) (xsollaUserClaims, error) {
	if authHeader == "" {
		return xsollaUserClaims{}, fmt.Errorf("missing Authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" || token == authHeader {
		return xsollaUserClaims{}, fmt.Errorf("invalid Authorization header")
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return xsollaUserClaims{}, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return xsollaUserClaims{}, fmt.Errorf("failed to decode JWT payload")
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return xsollaUserClaims{}, fmt.Errorf("invalid JWT claims")
	}

	userClaims := xsollaUserClaims{
		Sub:   stringClaim(claims, "sub"),
		Email: stringClaim(claims, "email"),
		Username: firstNonEmptyString(
			stringClaim(claims, "preferred_username"),
			stringClaim(claims, "username"),
			stringClaim(claims, "name"),
			stringClaim(claims, "nickname"),
		),
	}
	if userClaims.Sub == "" {
		return xsollaUserClaims{}, fmt.Errorf("missing sub claim in JWT")
	}

	if userClaims.Username == "" {
		userClaims.Username = userClaims.Sub
	}

	return userClaims, nil
}

func stringClaim(claims map[string]any, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (api *API) createXsollaPaymentToken(cfg xsollaConfig, item models.Item, userClaims xsollaUserClaims) (xsollaPaymentTokenResult, error) {
	payload := xsollaPaymentTokenRequest{
		Sandbox: true,
	}
	payload.User.ID.Value = userClaims.Sub
	payload.User.Name.Value = firstNonEmptyString(userClaims.Username, userClaims.Email)
	payload.User.Email.Value = userClaims.Email
	payload.User.Country.Value = "MY"
	payload.User.Country.AllowModify = false
	payload.Purchase.Items = []struct {
		SKU      string `json:"sku"`
		Quantity int    `json:"quantity"`
	}{
		{
			SKU:      item.ID,
			Quantity: 1,
		},
	}
	payload.Settings.Language = "en"
	payload.Settings.Currency = "USD"
	payload.Settings.ReturnURL = cfg.ReturnURL
	log.Printf("xsolla return url used when creating token: %s", cfg.ReturnURL)

	if len(payload.Purchase.Items) != 1 {
		return xsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items must contain exactly 1 item")
	}
	if payload.Purchase.Items[0].SKU == "" {
		return xsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items[0].sku must be non-empty")
	}
	if payload.Purchase.Items[0].Quantity != 1 {
		return xsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items[0].quantity must be 1")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return xsollaPaymentTokenResult{}, fmt.Errorf("marshal xsolla request: %w", err)
	}

	log.Printf("xsolla final payload user id: %s", payload.User.ID.Value)
	log.Printf("xsolla user id=%s email=%s", payload.User.ID.Value, payload.User.Email.Value)
	log.Printf("xsolla payment token request body: %s", string(body))

	url := fmt.Sprintf("https://store.xsolla.com/api/v3/project/%s/admin/payment/token", cfg.ProjectID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return xsollaPaymentTokenResult{}, fmt.Errorf("build xsolla request: %w", err)
	}

	authValue := base64.StdEncoding.EncodeToString([]byte(cfg.ProjectID + ":" + cfg.APIKey))
	req.Header.Set("Authorization", "Basic "+authValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Ip", "8.8.8.8")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return xsollaPaymentTokenResult{}, fmt.Errorf("call xsolla: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return xsollaPaymentTokenResult{}, fmt.Errorf("read xsolla response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return xsollaPaymentTokenResult{}, fmt.Errorf("xsolla error: %s", strings.TrimSpace(string(respBody)))
	}

	var result xsollaPaymentTokenResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return xsollaPaymentTokenResult{}, fmt.Errorf("decode xsolla response: %w", err)
	}

	if result.Token == "" {
		return xsollaPaymentTokenResult{}, fmt.Errorf("xsolla response missing token")
	}

	return result, nil
}

// POST /v1/webhooks/xsolla
func (api *API) XsollaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read webhook body"})
		return
	}

	log.Printf("xsolla webhook raw body: %s", string(body))

	var event xsollaWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook json"})
		return
	}

	notificationType := strings.ToLower(event.NotificationType)
	userID := event.User.ID.Value
	log.Printf("xsolla webhook notification_type=%s user.id=%s", notificationType, userID)

	switch notificationType {
	case "user_validation":
		if userID != "" {
			if err := api.store.EnsurePlayer(userID); err != nil {
				log.Printf("xsolla user_validation: failed to ensure player %s: %v", userID, err)
			} else {
				log.Printf("xsolla user_validation: player ensured for %s", userID)
			}
		}
	default:
		if isSuccessfulPaymentNotification(event) {
			sku := extractWebhookSKU(event)
			log.Printf("xsolla payment: user.id=%s sku=%s", userID, sku)
			if userID != "" && sku != "" {
				if err := api.store.GrantItem(userID, sku); err != nil {
					log.Printf("xsolla payment: failed to grant item %s to %s: %v", sku, userID, err)
				} else {
					log.Printf("xsolla payment: granted item %s to player %s", sku, userID)
				}
			} else {
				log.Printf("xsolla payment: missing user.id=%q or sku=%q, skipping grant", userID, sku)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func extractWebhookSKU(event xsollaWebhookEvent) string {
	if len(event.Purchase.VirtualItems.Items) > 0 {
		return event.Purchase.VirtualItems.Items[0].SKU
	}
	if len(event.Purchase.Items) > 0 {
		return event.Purchase.Items[0].SKU
	}
	return ""
}

func isSuccessfulPaymentNotification(event xsollaWebhookEvent) bool {
	normalizedType := strings.ToLower(event.NotificationType)
	if strings.Contains(normalizedType, "payment") || strings.Contains(normalizedType, "order") {
		status := strings.ToLower(event.Payment.Status)
		if status == "" {
			status = strings.ToLower(event.Order.Status)
		}
		return status == "" || status == "done" || status == "paid" || status == "successful" || status == "success"
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
