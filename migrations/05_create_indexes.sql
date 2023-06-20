-- +migrate Up notransaction
CREATE INDEX CONCURRENTLY index_names_on_title ON products USING gin (name gin_trgm_ops);

-- +migrate Down
DROP INDEX IF EXISTS index_names_on_title;
