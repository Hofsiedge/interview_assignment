CREATE TABLE IF NOT EXISTS prices (
    id serial PRIMARY KEY,
    coin text NOT NULL,
    price numeric NOT NULL,
    timestamp bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_prices_coin_timestamp ON prices (coin, timestamp);

