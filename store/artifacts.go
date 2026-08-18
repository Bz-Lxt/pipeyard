package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (db *DB) PutArtifact(ctx context.Context, job digest.Digest, nodeID string, a types.Artifact) error {
	if !a.Valid() {
		return fmt.Errorf("artifact invalid")
	}
	items, err := json.Marshal(a.Items)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, `INSERT INTO artifacts(digest,job_id,node_id,kind,body,items_json,bytes,created_at)
		VALUES(?,?,?,?,?,?,?,?)
		ON CONFLICT(digest) DO UPDATE SET body=excluded.body, items_json=excluded.items_json, bytes=excluded.bytes`,
		string(a.Digest), string(job), nodeID, string(a.Kind), a.Body, string(items), a.Bytes, db.NowText())
	return err
}

func (db *DB) GetArtifact(ctx context.Context, d digest.Digest) (types.Artifact, error) {
	var a types.Artifact
	var items, kind, id string
	err := db.sql.QueryRowContext(ctx, `SELECT digest,kind,body,items_json,bytes FROM artifacts WHERE digest=?`, string(d)).
		Scan(&id, &kind, &a.Body, &items, &a.Bytes)
	a.Digest = digest.Digest(id)
	if err == sql.ErrNoRows {
		return types.Artifact{}, types.ErrNotFound
	}
	if err != nil {
		return types.Artifact{}, err
	}
	a.Kind = types.Kind(kind)
	if items != "" {
		_ = json.Unmarshal([]byte(items), &a.Items)
	}
	return a, nil
}

func (db *DB) ListArtifacts(ctx context.Context, job digest.Digest) ([]types.Artifact, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT digest FROM artifacts WHERE job_id=? ORDER BY created_at, digest`, string(job))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []types.Artifact
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		a, err := db.GetArtifact(ctx, digest.Digest(s))
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (db *DB) ArtifactCount(ctx context.Context) (int, error) {
	var n int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM artifacts`).Scan(&n)
	return n, err
}
