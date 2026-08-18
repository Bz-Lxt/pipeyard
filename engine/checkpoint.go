package engine

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/store"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
)

// Checkpoint 把已应用到 SQLite 的日志前缀截掉，尾巴必须留下。
func (y *Yard) Checkpoint(ctx context.Context) error {
	y.mu.Lock()
	defer y.mu.Unlock()
	if err := y.guard(ctx); err != nil {
		return err
	}
	applied, err := y.db.WALApplied(ctx)
	if err != nil {
		return err
	}
	raw, err := y.journal.ReadAll()
	if err != nil {
		return err
	}
	off, err := wal.AppliedPrefix(raw, applied)
	if err != nil {
		return err
	}
	if err := y.journal.Append(wal.Record{Op: wal.OpCheckpoint, Payload: []byte(fmt.Sprintf("%d", applied))}); err != nil {
		return fmt.Errorf("%w: %v", types.ErrWAL, err)
	}
	if err := y.db.SetMeta(ctx, store.MetaCheckpoint, y.db.NowText()); err != nil {
		return err
	}
	if err := y.db.SetWALApplied(ctx, y.journal.Seq()); err != nil {
		return err
	}
	// 重新计算前缀：检查点记录本身已 applied，可以整文件截到当前结尾。
	raw, err = y.journal.ReadAll()
	if err != nil {
		return err
	}
	applied, err = y.db.WALApplied(ctx)
	if err != nil {
		return err
	}
	off, err = wal.AppliedPrefix(raw, applied)
	if err != nil {
		return err
	}
	if err := y.journal.TruncatePrefix(int64(off)); err != nil {
		return err
	}
	y.metrics.AddCheckpoint(1)
	y.bus.Publish(event.Event{Kind: event.KindCheckpoint, At: y.cfg.Clock.Now()})
	return nil
}
