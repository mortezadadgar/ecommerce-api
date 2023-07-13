-- +goose Up
CREATE TABLE IF NOT EXISTS users(
	id            bigserial   NOT NULL,
	email         text        NOT NULL UNIQUE,
	password_hash bytea       NOT NULL,
	created_at    timestamptz DEFAULT NOW(),
	updated_at    timestamptz DEFAULT NOW(),

	PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE IF EXISTS users;
