package engine

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
)

func (y *Yard) Cancel(ctx context.Context, id digest.Digest) error {
	y.mu.Lock()
	defer y.mu.Unlock()
	if err := y.guard(ctx); err != nil {
		return err
	}
	job, err := y.db.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job.Status.Terminal() {
		return types.ErrDone
	}
	if err := y.journal.Append(wal.Record{Op: wal.OpCancel, Job: string(id)}); err != nil {
		return fmt.Errorf("%w: %v", types.ErrWAL, err)
	}
	if err := y.db.CancelOpenNodes(ctx, id); err != nil {
		return err
	}
	if err := y.db.UpdateJobStatus(ctx, id, types.JobCanceled, "canceled"); err != nil {
		return err
	}
	if err := y.db.SetWALApplied(ctx, y.journal.Seq()); err != nil {
		return err
	}
	y.bus.Publish(event.Event{Kind: event.KindCancel, Job: string(id), At: y.cfg.Clock.Now()})
	return nil
}
