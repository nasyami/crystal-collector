package models

import "strings"

type XsollaConfig struct {
	ProjectID string
	APIKey    string
	ReturnURL string
}

type MissingEnvError struct {
	Names []string
}

func (e MissingEnvError) Error() string {
	return "missing required environment variables: " + strings.Join(e.Names, ", ")
}

type XsollaPaymentTokenRequest struct {
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

type XsollaPaymentTokenResult struct {
	Token   string `json:"token"`
	OrderID int64  `json:"order_id"`
}

type XsollaUserClaims struct {
	Sub      string
	Email    string
	Username string
}

type XsollaWebhookEvent struct {
	NotificationType string `json:"notification_type"`
	User             struct {
		ID         string `json:"id"`
		ExternalID string `json:"external_id"`
	} `json:"user"`
	Items []struct {
		SKU string `json:"sku"`
	} `json:"items"`
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
