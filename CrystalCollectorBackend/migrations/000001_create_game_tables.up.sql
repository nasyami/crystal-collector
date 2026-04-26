CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    crystals INTEGER NOT NULL DEFAULT 0 CHECK (crystals >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE shop_items (
    id BIGSERIAL PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price_cents INTEGER NOT NULL CHECK (price_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE player_items (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    shop_item_id BIGINT NOT NULL REFERENCES shop_items(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (player_id, shop_item_id)
);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    provider_payment_id TEXT NOT NULL UNIQUE,
    amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
    currency CHAR(3) NOT NULL CHECK (currency = UPPER(currency)),
    status TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO shop_items (sku, name, description, price_cents) VALUES
    ('skin_blue', 'Blue Skin', 'A cool blue color skin for your character.', 199),
    ('skin_red', 'Red Skin', 'A bold red color skin for your character.', 199),
    ('skin_green', 'Green Skin', 'A fresh green color skin for your character.', 199),
    ('skin_purple', 'Purple Skin', 'A vibrant purple color skin for your character.', 299),
    ('skin_gold', 'Gold Skin', 'A premium gold color skin for your character.', 499);
