package store

import (
	"context"
	"database/sql"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

// RefreshJobStatus 在同一事务里重算作业状态并落库。
// 读取节点状态、推导作业状态、写回 status/error 必须原子，否则并发 Tick
// 各自带着旧快照回写会把已成功的作业盖回 running。
func (db *DB) RefreshJobStatus(ctx context.Context, id digest.Digest) error {
	return db.Tx(ctx, func(tx *sql.Tx) error {
		var oldErr string
		if err := tx.QueryRowContext(ctx, `SELECT error FROM jobs WHERE id=?`, string(id)).Scan(&oldErr); err != nil {
			if err == sql.ErrNoRows {
				return types.ErrNotFound
			}
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT status FROM nodes WHERE job_id=?`, string(id))
		if err != nil {
			return err
		}
		var statuses []types.NodeStatus
		for rows.Next() {
			var st string
			if err := rows.Scan(&st); err != nil {
				_ = rows.Close()
				return err
			}
			statuses = append(statuses, types.NodeStatus(st))
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		_ = rows.Close()
		next := types.Rollup(statuses)
		_, err = tx.ExecContext(ctx, `UPDATE jobs SET status=?, error=?, updated_at=? WHERE id=?`,
			string(next), oldErr, db.NowText(), string(id))
		return err
	})
}
