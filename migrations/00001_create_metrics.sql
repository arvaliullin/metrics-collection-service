-- +goose Up
CREATE TABLE IF NOT EXISTS metrics_gauge (
	id TEXT PRIMARY KEY,
	value DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS metrics_counter (
	id TEXT PRIMARY KEY,
	value BIGINT NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS metrics_counter;
DROP TABLE IF EXISTS metrics_gauge;
