-- +goose Up
CREATE TABLE IF NOT EXISTS categories(
	id          bigserial   NOT     NULL,
	name        text        NOT     NULL UNIQUE,
	description text        NOT     NULL,
	created_at  timestamptz DEFAULT NOW(),
	updated_at  timestamptz DEFAULT NOW(),
	version     int         DEFAULT 1,

	PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE IF EXISTS categories;
