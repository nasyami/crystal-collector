package store

import (
	"database/sql"
	"game-backend/models"

	_ "github.com/lib/pq"
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

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS shop_items (
			id BIGSERIAL PRIMARY KEY,
			sku TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			price_cents INTEGER NOT NULL CHECK (price_cents >= 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS players (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			crystals INTEGER NOT NULL DEFAULT 0 CHECK (crystals >= 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS player_items (
			id BIGSERIAL PRIMARY KEY,
			player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			shop_item_id BIGINT NOT NULL REFERENCES shop_items(id) ON DELETE RESTRICT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (player_id, shop_item_id)
		);
		CREATE TABLE IF NOT EXISTS payments (
			id BIGSERIAL PRIMARY KEY,
			player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			provider_payment_id TEXT NOT NULL UNIQUE,
			amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
			currency CHAR(3) NOT NULL CHECK (currency = UPPER(currency)),
			status TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'refunded')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func SeedShopItems(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM shop_items").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := db.Exec(`
			INSERT INTO shop_items (sku, name, description, price_cents) VALUES
			('skin_blue', 'Blue Skin', 'A cool blue color skin for your character.', 199),
			('skin_red', 'Red Skin', 'A bold red color skin for your character.', 199),
			('skin_green', 'Green Skin', 'A fresh green color skin for your character.', 199),
			('skin_purple', 'Purple Skin', 'A vibrant purple color skin for your character.', 299),
			('skin_gold', 'Gold Skin', 'A premium gold color skin for your character.', 499)
		`)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) Migrate() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS shop_items (
			id          INTEGER PRIMARY KEY,
			sku         TEXT    NOT NULL UNIQUE,
			name        TEXT    NOT NULL,
			description TEXT    NOT NULL,
			price_cents INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS players (
			id       BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS player_items (
			player_id    BIGINT  NOT NULL REFERENCES players(id),
			shop_item_id INTEGER NOT NULL REFERENCES shop_items(id),
			UNIQUE (player_id, shop_item_id)
		);

		CREATE TABLE IF NOT EXISTS payments (
			id         TEXT PRIMARY KEY,
			player_id  TEXT NOT NULL,
			item_id    TEXT NOT NULL,
			status     TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *PostgresStore) SeedItems() error {
	_, err := s.DB.Exec(`
		INSERT INTO shop_items (id, sku, name, description, price_cents) VALUES
			(1, 'skin_blue',   'Blue Skin',   'A cool blue crystal skin',   199),
			(2, 'skin_red',    'Red Skin',    'A fiery red crystal skin',   199),
			(3, 'skin_green',  'Green Skin',  'A lush green crystal skin',  199),
			(4, 'skin_purple', 'Purple Skin', 'A royal purple crystal skin',199),
			(5, 'skin_orange', 'Orange Skin', 'A vibrant orange crystal skin',199)
		ON CONFLICT (id) DO NOTHING
	`)
	return err
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
	s.DB.Exec(`ALTER TABLE player_items ALTER COLUMN player_id TYPE TEXT`)

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

func (s *PostgresStore) EnsurePlayer(xsollaSub string) error {
	_, err := s.DB.Exec(
		`INSERT INTO players (username) VALUES ($1) ON CONFLICT (username) DO NOTHING`,
		xsollaSub,
	)
	return err
}
