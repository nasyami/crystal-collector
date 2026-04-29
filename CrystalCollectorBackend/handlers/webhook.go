package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"game-backend/models"
)

// POST /v1/webhooks/xsolla
func (api *API) XsollaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read webhook body"})
		return
	}

	log.Printf("xsolla webhook raw body: %s", string(body))

	var event models.XsollaWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid webhook json"})
		return
	}

	notificationType := strings.ToLower(event.NotificationType)
	userID := event.User.ID
	if userID == "" {
		userID = event.User.ExternalID
	}
	log.Printf("xsolla webhook notification_type=%s user.id=%s", notificationType, userID)

	switch notificationType {
	case "user_validation":
		if userID != "" {
			log.Println("user_validation received:", userID)
			if err := api.store.EnsurePlayer(userID); err != nil {
				log.Printf("DB Warning (Validation): %v", err)
			}
			w.WriteHeader(http.StatusOK)
			return
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

func extractWebhookSKU(event models.XsollaWebhookEvent) string {
	if len(event.Items) > 0 && event.Items[0].SKU != "" {
		return event.Items[0].SKU
	}
	if len(event.Purchase.VirtualItems.Items) > 0 {
		return event.Purchase.VirtualItems.Items[0].SKU
	}
	if len(event.Purchase.Items) > 0 {
		return event.Purchase.Items[0].SKU
	}
	return ""
}

func isSuccessfulPaymentNotification(event models.XsollaWebhookEvent) bool {
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
