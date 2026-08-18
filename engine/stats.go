package engine

import (
	"context"

	"github.com/Bz-Lxt/pipeyard/metric"
	"github.com/Bz-Lxt/pipeyard/repair"
	"github.com/Bz-Lxt/pipeyard/types"
)

type Stats struct {
	Jobs     int
	Arts     int
	Metrics  metric.Snapshot
	Repair   repair.Report
	Workers  int
	LeaseSec int
}

func (y *Yard) Stats(ctx context.Context) (Stats, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.closed {
		return Stats{}, y.closedErr()
	}
	n, err := y.db.JobCount(ctx)
	if err != nil {
		return Stats{}, err
	}
	a, err := y.db.ArtifactCount(ctx)
	if err != nil {
		return Stats{}, err
	}
	rep, err := repair.Inspect(ctx, y.db, y.journal)
	if err != nil {
		return Stats{}, err
	}
	return Stats{
		Jobs: n, Arts: a, Metrics: y.metrics.Snapshot(), Repair: rep,
		Workers: y.cfg.Workers, LeaseSec: y.cfg.LeaseSec,
	}, nil
}

func (y *Yard) closedErr() error {
	return types.ErrClosed
}
