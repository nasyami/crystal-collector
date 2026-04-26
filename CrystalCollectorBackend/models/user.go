package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Crystals int    `json:"crystals"`
}
