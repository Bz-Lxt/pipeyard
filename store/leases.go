package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

// LeaseRef 指向被租约占用的节点。
type LeaseRef struct {
	Job  string
	Node string
}

// ClaimLease 在同一事务里检查租约并写入。已占用且未过期返回 ErrLeaseHeld。
func (db *DB) ClaimLease(ctx context.Context, job digest.Digest, nodeID, worker string, leaseSec int) error {
	now := db.clock.Now()
	until := clock.Format(clock.AddSeconds(now, leaseSec))
	nowText := clock.Format(now)
	return db.Tx(ctx, func(tx *sql.Tx) error {
		var st, by, untilText string
		err := tx.QueryRowContext(ctx, `SELECT status,lease_by,lease_until FROM nodes WHERE job_id=? AND node_id=?`,
			string(job), nodeID).Scan(&st, &by, &untilText)
		if err == sql.ErrNoRows {
			return types.ErrUnknownNode
		}
		if err != nil {
			return err
		}
		if types.NodeStatus(st).Terminal() {
			return types.ErrDone
		}
		if by != "" && untilText != "" && untilText > nowText && by != worker {
			return types.ErrLeaseHeld
		}
		_, err = tx.ExecContext(ctx, `UPDATE nodes SET status=?, lease_by=?, lease_until=?, attempt=attempt+1
			WHERE job_id=? AND node_id=?`,
			string(types.NodeLeased), worker, until, string(job), nodeID)
		return err
	})
}

func (db *DB) Heartbeat(ctx context.Context, job digest.Digest, nodeID, worker string, leaseSec int) error {
	now := db.clock.Now()
	until := clock.Format(clock.AddSeconds(now, leaseSec))
	res, err := db.sql.ExecContext(ctx, `UPDATE nodes SET lease_until=? WHERE job_id=? AND node_id=? AND lease_by=? AND status=?`,
		until, string(job), nodeID, worker, string(types.NodeLeased))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: heartbeat", types.ErrLeaseHeld)
	}
	return nil
}

func (db *DB) ActiveLeases(ctx context.Context) ([]LeaseRef, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT job_id,node_id FROM nodes WHERE status=? ORDER BY job_id,node_id`,
		string(types.NodeLeased))
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

func (db *DB) LeaseOwner(ctx context.Context, job digest.Digest, nodeID string) (string, string, error) {
	var by, until string
	err := db.sql.QueryRowContext(ctx, `SELECT lease_by,lease_until FROM nodes WHERE job_id=? AND node_id=?`,
		string(job), nodeID).Scan(&by, &until)
	if err == sql.ErrNoRows {
		return "", "", types.ErrUnknownNode
	}
	return by, until, err
}
