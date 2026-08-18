package store

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (db *DB) JobsByStatus(ctx context.Context, st types.JobStatus) ([]digest.Digest, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id FROM jobs WHERE status=? ORDER BY created_at, id`, string(st))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digest.Digest
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, digest.Digest(s))
	}
	return out, rows.Err()
}

func (db *DB) NodeCounts(ctx context.Context, job digest.Digest) (map[types.NodeStatus]int, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT status, COUNT(*) FROM nodes WHERE job_id=? GROUP BY status`, string(job))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[types.NodeStatus]int{}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		m[types.NodeStatus(st)] = n
	}
	return m, rows.Err()
}

func (db *DB) OpenJobIDs(ctx context.Context) ([]digest.Digest, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id FROM jobs WHERE status IN (?,?) ORDER BY created_at`,
		string(types.JobPending), string(types.JobRunning))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []digest.Digest
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, digest.Digest(s))
	}
	return out, rows.Err()
}

func (db *DB) HasJob(ctx context.Context, id digest.Digest) (bool, error) {
	var n int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM jobs WHERE id=?`, string(id)).Scan(&n)
	return n > 0, err
}

func (db *DB) TouchJob(ctx context.Context, id digest.Digest) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE jobs SET updated_at=? WHERE id=?`, db.NowText(), string(id))
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return fmt.Errorf("%w: touch", types.ErrNotFound)
	}
	return nil
}
