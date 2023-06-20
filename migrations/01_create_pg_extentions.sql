-- +migrate Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- +migrate Down
DROP EXTENSION IF EXISTS pg_trgm;
