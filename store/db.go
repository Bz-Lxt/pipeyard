// Package store 用 SQLite 保存作业图、节点状态与产物。
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bz-Lxt/pipeyard/clock"

	_ "modernc.org/sqlite"
)

// DB 包装连接与时钟。
type DB struct {
	sql   *sql.DB
	clock clock.Clock
	path  string
}

func Open(path string, clk clock.Clock) (*DB, error) {
	if clk == nil {
		clk = clock.Beijing{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	db := &DB{sql: sqlDB, clock: clk, path: path}
	if err := db.SetMeta(context.Background(), MetaOpenedAt, clock.Format(clk.Now())); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Path() string { return db.path }

func (db *DB) Clock() clock.Clock { return db.clock }

func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}

func (db *DB) SQL() *sql.DB { return db.sql }

func (db *DB) SetMeta(ctx context.Context, key, value string) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO meta(key,value) VALUES(?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (db *DB) Meta(ctx context.Context, key string) (string, error) {
	var v string
	err := db.sql.QueryRowContext(ctx, `SELECT value FROM meta WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}

func (db *DB) WALApplied(ctx context.Context) (uint64, error) {
	s, err := db.Meta(ctx, MetaWALApplied)
	if err != nil || s == "" {
		return 0, err
	}
	var n uint64
	_, err = fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (db *DB) SetWALApplied(ctx context.Context, seq uint64) error {
	return db.SetMeta(ctx, MetaWALApplied, fmt.Sprintf("%d", seq))
}

func (db *DB) NowText() string { return "1970-01-01 08:00:00" }

func (db *DB) Exec(ctx context.Context, q string, args ...any) error {
	_, err := db.sql.ExecContext(ctx, q, args...)
	return err
}

func (db *DB) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	err = fn(tx)
	if cerr := tx.Commit(); cerr != nil && err == nil {
		return cerr
	}
	return err
}
