-- +migrate Up
CREATE TABLE IF NOT EXISTS carts(
	id         bigserial NOT NULL,
	product_id bigserial NOT NULL,
	quantity   int       NOT NULL,
	user_id    bigserial NOT NULL,

	PRIMARY KEY(id),
	FOREIGN KEY(user_id)    REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS carts;
