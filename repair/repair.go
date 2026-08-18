// Package repair 提供作业场自检：过期租约回收、WAL 与 SQLite 序号对齐。
package repair

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/store"
	"github.com/Bz-Lxt/pipeyard/wal"
)

type Report struct {
	Released int
	Applied  uint64
	WALSeq   uint64
	Jobs     int
}

func SweepLeases(ctx context.Context, db *store.DB) (int, error) {
	return db.ReleaseExpired(ctx, db.NowText())
}

func Inspect(ctx context.Context, db *store.DB, journal *wal.Journal) (Report, error) {
	var r Report
	var err error
	r.Released, err = SweepLeases(ctx, db)
	if err != nil {
		return r, err
	}
	r.Applied, err = db.WALApplied(ctx)
	if err != nil {
		return r, err
	}
	raw, err := journal.ReadAll()
	if err != nil {
		return r, err
	}
	recs, err := wal.Replay(raw)
	if err != nil {
		return r, fmt.Errorf("replay: %w", err)
	}
	r.WALSeq = wal.LastSeq(recs)
	r.Jobs, err = db.JobCount(ctx)
	return r, err
}

func AlignSeq(journal *wal.Journal, applied uint64) {
	if applied > journal.Seq() {
		journal.SetSeq(applied)
	}
}
