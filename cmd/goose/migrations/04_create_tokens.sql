-- +goose Up
CREATE UNLOGGED TABLE IF NOT EXISTS tokens(
	hashed  bytea       NOT NULL,
	user_id bigserial   NOT NULL,
	expiry  timestamptz NOT NULL,

	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS tokens;
