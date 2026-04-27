package store

import (
	"database/sql"

	_ "github.com/lib/pq"
	"game-backend/models"
)

type PostgresStore struct {
	DB *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{DB: db}, nil
}

func (s *PostgresStore) Items() []models.Item {
	rows, err := s.DB.Query(`SELECT sku, name, description, price_cents FROM shop_items ORDER BY id`)
	if err != nil {
		return []models.Item{}
	}
	defer rows.Close()
	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (s *PostgresStore) User() models.User {
	return models.User{ID: "server", Username: "server", Crystals: 0}
}

func (s *PostgresStore) GetOwnedItems(xsollaSub string) ([]models.Item, error) {
	query := `
		SELECT si.sku, si.name, si.description, si.price_cents
		FROM player_items pi
		JOIN players p ON pi.player_id = p.id
		JOIN shop_items si ON pi.shop_item_id = si.id
		WHERE p.username = $1
	`
	rows, err := s.DB.Query(query, xsollaSub)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *PostgresStore) GrantItem(xsollaSub, sku string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var playerID int64
	err = tx.QueryRow(`SELECT id FROM players WHERE username = $1`, xsollaSub).Scan(&playerID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`INSERT INTO players (username) VALUES ($1) RETURNING id`, xsollaSub).Scan(&playerID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var shopItemID int64
	if err = tx.QueryRow(`SELECT id FROM shop_items WHERE sku = $1`, sku).Scan(&shopItemID); err != nil {
		return err
	}

	if _, err = tx.Exec(
		`INSERT INTO player_items (player_id, shop_item_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		playerID, shopItemID,
	); err != nil {
		return err
	}

	return tx.Commit()
}
