package store

import (
	"context"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (db *DB) listNodes(ctx context.Context, id digest.Digest) ([]types.Node, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT node_id,kind,param,weight,status,lease_by,lease_until,error,artifact_json,attempt
		FROM nodes WHERE job_id=? ORDER BY node_id`, string(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []types.Node
	for rows.Next() {
		var n types.Node
		var kind, st, art string
		if err := rows.Scan(&n.ID, &kind, &n.Param, &n.Weight, &st, &n.LeaseBy, &n.LeaseUntil, &n.Error, &art, &n.Attempt); err != nil {
			return nil, err
		}
		n.Kind = types.Kind(kind)
		n.Status = types.NodeStatus(st)
		if art != "" {
			a, err := types.UnmarshalArtifact([]byte(art))
			if err != nil {
				return nil, err
			}
			n.Artifact = &a
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (db *DB) UpdateNode(ctx context.Context, job digest.Digest, n types.Node) error {
	art := ""
	if n.Artifact != nil {
		b, err := n.Artifact.Marshal()
		if err != nil {
			return err
		}
		art = string(b)
	}
	_, err := db.sql.ExecContext(ctx, `UPDATE nodes SET status=?, error=?, artifact_json=?, attempt=?
		WHERE job_id=? AND node_id=?`,
		string(n.Status), n.Error, art, n.Attempt, string(job), n.ID)
	return err
}

func (db *DB) GetNode(ctx context.Context, job digest.Digest, nodeID string) (types.Node, error) {
	nodes, err := db.listNodes(ctx, job)
	if err != nil {
		return types.Node{}, err
	}
	for _, n := range nodes {
		if n.ID == nodeID {
			return n, nil
		}
	}
	return types.Node{}, types.ErrUnknownNode
}

func (db *DB) StatusMap(ctx context.Context, job digest.Digest) (map[string]types.NodeStatus, error) {
	nodes, err := db.listNodes(ctx, job)
	if err != nil {
		return nil, err
	}
	m := make(map[string]types.NodeStatus, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n.Status
	}
	return m, nil
}

func (db *DB) ClearLease(ctx context.Context, job digest.Digest, nodeID string) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE nodes SET lease_by='', lease_until='' WHERE job_id=? AND node_id=?`,
		string(job), nodeID)
	return err
}

func (db *DB) CancelOpenNodes(ctx context.Context, job digest.Digest) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE nodes SET status=? WHERE job_id=? AND status IN (?,?,?)`,
		string(types.NodeCanceled), string(job),
		string(types.NodePending), string(types.NodeReady), string(types.NodeLeased))
	return err
}

func (db *DB) ExpiredLeases(ctx context.Context, now string) ([]LeaseRef, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT job_id,node_id FROM nodes
		WHERE status=? AND lease_until<>'' AND lease_until<?`, string(types.NodeLeased), now)
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

func (db *DB) ReleaseExpired(ctx context.Context, now string) (int, error) {
	res, err := db.sql.ExecContext(ctx, `UPDATE nodes SET status=?, lease_by='', lease_until=''
		WHERE status=? AND lease_until<>'' AND lease_until<?`,
		string(types.NodeReady), string(types.NodeLeased), now)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
