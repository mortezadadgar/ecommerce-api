-- +migrate Up
CREATE TABLE IF NOT EXISTS products(
	id             bigserial     NOT NULL, 
	name           text          NOT NULL UNIQUE,
	description    text          NOT NULL UNIQUE,
	category       text          NOT NULL,
	price          int           NOT NULL,
	quantity       int           NOT NULL,
	created_at     timestamptz   DEFAULT NOW(),
	updated_at     timestamptz   DEFAULT NOW(),

	PRIMARY KEY(id),
	FOREIGN KEY(category) REFERENCES categories(name)
);

-- +migrate Down
DROP TABLE IF EXISTS products;
DROP EXTENSION IF EXISTS pg_trgm;
