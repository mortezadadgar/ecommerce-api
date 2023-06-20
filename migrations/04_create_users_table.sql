-- +migrate Up
CREATE TABLE IF NOT EXISTS users(
	id            bigserial   NOT NULL,
	email         text        NOT NULL UNIQUE,
	password_hash bytea       NOT NULL,
	is_admin      bool        DEFAULT  false,
	created_at    timestamptz DEFAULT NOW(),
	updated_at    timestamptz DEFAULT NOW(),

	PRIMARY KEY(id)
);

-- +migrate Down
DROP TABLE IF EXISTS users;
