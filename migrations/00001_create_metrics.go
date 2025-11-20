package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateMetrics, downCreateMetrics)
}

func upCreateMetrics(ctx context.Context, tx *sql.Tx) error {
	const createGaugeTable = `
CREATE TABLE IF NOT EXISTS metrics_gauge (
	id TEXT PRIMARY KEY,
	value DOUBLE PRECISION NOT NULL
)`
	if _, err := tx.ExecContext(ctx, createGaugeTable); err != nil {
		return err
	}

	const createCounterTable = `
CREATE TABLE IF NOT EXISTS metrics_counter (
	id TEXT PRIMARY KEY,
	value BIGINT NOT NULL DEFAULT 0
)`
	if _, err := tx.ExecContext(ctx, createCounterTable); err != nil {
		return err
	}

	return nil
}

func downCreateMetrics(ctx context.Context, tx *sql.Tx) error {

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS metrics_counter`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS metrics_gauge`); err != nil {
		return err
	}
	return nil
}
