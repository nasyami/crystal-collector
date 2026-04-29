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
)

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
		var missingErr models.MissingEnvError
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

func (api *API) buildPaymentTokenResponse(p payment, item models.Item, userClaims models.XsollaUserClaims) (paymentTokenResponse, error) {
	cfg, err := loadXsollaConfig()
	if err != nil {
		return paymentTokenResponse{}, err
	}

	xsollaToken, err := api.createXsollaPaymentToken(cfg, item, userClaims)
	if err != nil {
		return paymentTokenResponse{}, err
	}

	paymentID := p.PaymentID
	if xsollaToken.OrderID != 0 {
		paymentID = fmt.Sprintf("%d", xsollaToken.OrderID)
	}

	return paymentTokenResponse{
		PaymentID:     paymentID,
		Token:         xsollaToken.Token,
		PayStationURL: "https://sandbox-secure.xsolla.com/paystation4/?token=" + xsollaToken.Token,
	}, nil
}

func loadXsollaConfig() (models.XsollaConfig, error) {
	cfg := models.XsollaConfig{
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
		return models.XsollaConfig{}, models.MissingEnvError{Names: missing}
	}
	return cfg, nil
}

func getReturnURL() string {
	u := os.Getenv("XSOLLA_RETURN_URL")
	if u == "" {
		return "https://crystal-collector-ten.vercel.app/shop"
	}
	return u
}

func (api *API) createXsollaPaymentToken(cfg models.XsollaConfig, item models.Item, userClaims models.XsollaUserClaims) (models.XsollaPaymentTokenResult, error) {
	payload := models.XsollaPaymentTokenRequest{Sandbox: true}
	payload.User.ID.Value = userClaims.Sub
	payload.User.Name.Value = firstNonEmptyString(userClaims.Username, userClaims.Email)
	payload.User.Email.Value = userClaims.Email
	payload.User.Country.Value = "MY"
	payload.User.Country.AllowModify = false
	payload.Purchase.Items = []struct {
		SKU      string `json:"sku"`
		Quantity int    `json:"quantity"`
	}{
		{SKU: item.ID, Quantity: 1},
	}
	payload.Settings.Language = "en"
	payload.Settings.Currency = "USD"
	payload.Settings.ReturnURL = cfg.ReturnURL
	log.Printf("xsolla return url used when creating token: %s", cfg.ReturnURL)

	if len(payload.Purchase.Items) != 1 {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items must contain exactly 1 item")
	}
	if payload.Purchase.Items[0].SKU == "" {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items[0].sku must be non-empty")
	}
	if payload.Purchase.Items[0].Quantity != 1 {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("xsolla payload validation failed: purchase.items[0].quantity must be 1")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("marshal xsolla request: %w", err)
	}

	log.Printf("xsolla final payload user id: %s", payload.User.ID.Value)
	log.Printf("xsolla user id=%s email=%s", payload.User.ID.Value, payload.User.Email.Value)
	log.Printf("xsolla payment token request body: %s", string(body))

	url := fmt.Sprintf("https://store.xsolla.com/api/v3/project/%s/admin/payment/token", cfg.ProjectID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("build xsolla request: %w", err)
	}

	authValue := base64.StdEncoding.EncodeToString([]byte(cfg.ProjectID + ":" + cfg.APIKey))
	req.Header.Set("Authorization", "Basic "+authValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Ip", "8.8.8.8")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("call xsolla: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("read xsolla response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("xsolla error: %s", strings.TrimSpace(string(respBody)))
	}

	var result models.XsollaPaymentTokenResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("decode xsolla response: %w", err)
	}

	if result.Token == "" {
		return models.XsollaPaymentTokenResult{}, fmt.Errorf("xsolla response missing token")
	}

	return result, nil
}

func extractXsollaUserClaims(authHeader string) (models.XsollaUserClaims, error) {
	if authHeader == "" {
		return models.XsollaUserClaims{}, fmt.Errorf("missing Authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" || token == authHeader {
		return models.XsollaUserClaims{}, fmt.Errorf("invalid Authorization header")
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return models.XsollaUserClaims{}, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return models.XsollaUserClaims{}, fmt.Errorf("failed to decode JWT payload")
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return models.XsollaUserClaims{}, fmt.Errorf("invalid JWT claims")
	}

	userClaims := models.XsollaUserClaims{
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
		return models.XsollaUserClaims{}, fmt.Errorf("missing sub claim in JWT")
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
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
