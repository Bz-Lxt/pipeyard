package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (db *DB) InsertJob(ctx context.Context, job types.Job, spec types.Spec) error {
	raw, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return db.Tx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO jobs(id,name,status,spec_json,fingerprint,quota,error,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?)`,
			string(job.ID), job.Name, string(job.Status), string(raw), string(spec.Fingerprint()),
			job.Quota, job.Error, job.CreatedAt, job.UpdatedAt)
		if err != nil {
			return fmt.Errorf("insert job: %w", err)
		}
		for _, n := range job.Nodes {
			art := ""
			if n.Artifact != nil {
				b, err := n.Artifact.Marshal()
				if err != nil {
					return err
				}
				art = string(b)
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO nodes(job_id,node_id,kind,param,weight,status,lease_by,lease_until,error,artifact_json,attempt)
				VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
				string(job.ID), n.ID, string(n.Kind), n.Param, n.Weight, string(n.Status),
				n.LeaseBy, n.LeaseUntil, n.Error, art, n.Attempt)
			if err != nil {
				return err
			}
		}
		for _, e := range job.Edges {
			if _, err := tx.ExecContext(ctx, `INSERT INTO edges(job_id,src,dst) VALUES(?,?,?)`,
				string(job.ID), e.From, e.To); err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *DB) UpdateJobStatus(ctx context.Context, id digest.Digest, st types.JobStatus, errText string) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE jobs SET status=?, error=?, updated_at=? WHERE id=?`,
		string(st), errText, db.NowText(), string(id))
	return err
}

func (db *DB) GetJob(ctx context.Context, id digest.Digest) (types.Job, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id,name,status,quota,error,created_at,updated_at FROM jobs WHERE id=?`, string(id))
	var j types.Job
	var idText, st string
	if err := row.Scan(&idText, &j.Name, &st, &j.Quota, &j.Error, &j.CreatedAt, &j.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return types.Job{}, types.ErrNotFound
		}
		return types.Job{}, err
	}
	j.ID = digest.Digest(idText)
	j.Status = types.JobStatus(st)
	nodes, err := db.listNodes(ctx, id)
	if err != nil {
		return types.Job{}, err
	}
	j.Nodes = nodes
	edges, err := db.listEdges(ctx, id)
	if err != nil {
		return types.Job{}, err
	}
	j.Edges = edges
	return j, nil
}

func (db *DB) ListJobs(ctx context.Context) ([]types.Job, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id FROM jobs ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []digest.Digest
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		ids = append(ids, digest.Digest(s))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]types.Job, 0, len(ids))
	for _, id := range ids {
		j, err := db.GetJob(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, nil
}

func (db *DB) JobCount(ctx context.Context) (int, error) {
	var n int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM jobs`).Scan(&n)
	return n, err
}

func (db *DB) NameTaken(ctx context.Context, name string) (bool, error) {
	var n int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM jobs WHERE name=?`, name).Scan(&n)
	return n > 0, err
}

func (db *DB) listEdges(ctx context.Context, id digest.Digest) ([]types.Edge, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT src,dst FROM edges WHERE job_id=? ORDER BY src,dst`, string(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []types.Edge
	for rows.Next() {
		var e types.Edge
		if err := rows.Scan(&e.From, &e.To); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
