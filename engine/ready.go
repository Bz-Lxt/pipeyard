package engine

import (
	"context"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/schedule"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (y *Yard) ReadyNodes(ctx context.Context, id digest.Digest) ([]string, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.closed {
		return nil, types.ErrClosed
	}
	job, err := y.db.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	g := y.graphs[string(id)]
	if g == nil {
		return nil, types.ErrNotFound
	}
	st := map[string]types.NodeStatus{}
	for _, n := range job.Nodes {
		st[n.ID] = n.Status
	}
	return g.Ready(st), nil
}

func (y *Yard) Schedule(ctx context.Context, n int) ([]schedule.Candidate, error) {
	jobs, err := y.List(ctx)
	if err != nil {
		return nil, err
	}
	var cands []schedule.Candidate
	for _, job := range jobs {
		if job.Status.Terminal() {
			continue
		}
		ready, err := y.ReadyNodes(ctx, job.ID)
		if err != nil {
			return nil, err
		}
		cands = append(cands, schedule.FromJob(job, ready)...)
	}
	return schedule.Pick(schedule.FairShare(cands, 2), n), nil
}
