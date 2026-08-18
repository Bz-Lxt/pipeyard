package store

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

type JobRow struct {
	ID        digest.Digest
	Name      string
	Status    types.JobStatus
	CreatedAt string
	UpdatedAt string
}

func (db *DB) ListJobRows(ctx context.Context) ([]JobRow, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id,name,status,created_at,updated_at FROM jobs ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []JobRow
	for rows.Next() {
		var r JobRow
		var id, st string
		if err := rows.Scan(&id, &r.Name, &st, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.ID = digest.Digest(id)
		r.Status = types.JobStatus(st)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (db *DB) FailedNodes(ctx context.Context) ([]LeaseRef, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT job_id,node_id FROM nodes WHERE status=? ORDER BY job_id,node_id`,
		string(types.NodeFailed))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LeaseRef
	for rows.Next() {
		var r LeaseRef
		if err := rows.Scan(&r.Job, &r.Node); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (db *DB) LatestJob(ctx context.Context) (types.Job, error) {
	var id string
	err := db.sql.QueryRowContext(ctx, `SELECT id FROM jobs ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&id)
	if err != nil {
		return types.Job{}, fmt.Errorf("%w: latest", types.ErrNotFound)
	}
	return db.GetJob(ctx, digest.Digest(id))
}
