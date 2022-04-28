package database

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"github.com/scrolllockdev/test-devops/internal/model"
)

func CreateStorageTable(ctx context.Context, db *sql.DB, tableName string) error {
	if db == nil {
		return errors.New("you haven`t opened the database connection")
	}
	_, err := db.ExecContext(ctx, "CREATE TABLE if not exists "+tableName+" (id varchar(255) NOT NULL UNIQUE, type varchar(20) NOT NULL, delta integer default NULL, value double precision default NULL);")
	if err != nil {
		return err
	}

	return nil
}

func InsertCounterValueToTable(ctx context.Context, db *sql.DB, metric model.Metric, tableName string) error {
	if db == nil {
		return nil
	}
	_, err := db.ExecContext(ctx, "INSERT INTO "+tableName+"(id, type, delta) VALUES($1,$2,$3) "+
		"ON CONFLICT (id) DO UPDATE SET delta = storage.delta + EXCLUDED.delta",
		metric.ID,
		metric.MType,
		metric.Delta)
	if err != nil {
		return err
	}
	return nil
}

func InsertGaugeValueToTable(ctx context.Context, db *sql.DB, metric model.Metric, tableName string) error {
	if db == nil {
		return nil
	}
	_, err := db.ExecContext(ctx, "INSERT INTO "+tableName+"(id, type, value) VALUES($1,$2,$3) "+
		"ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value",
		metric.ID,
		metric.MType,
		metric.Value)
	if err != nil {
		return err
	}
	return nil
}

func MultiplyUpdates(ctx context.Context, db *sql.DB, metrics []model.Metric, tableName string) error {
	if db == nil {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO "+tableName+"(id, type, value, delta) VALUES($1,$2,$3,$4) "+
		"ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, delta = storage.delta + EXCLUDED.delta")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, metric := range metrics {
		if _, err = stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Value, metric.Delta); err != nil {
			return err
		}
	}
	return tx.Commit()
}
