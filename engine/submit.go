package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/graph"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
)

func (y *Yard) Submit(ctx context.Context, spec types.Spec) (digest.Digest, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if err := y.guard(ctx); err != nil {
		return "", err
	}
	spec, err := spec.Normalize()
	if err != nil {
		return "", err
	}
	g, err := graph.Build(spec)
	if err != nil {
		return "", err
	}
	taken, err := y.db.NameTaken(ctx, spec.Name)
	if err != nil {
		return "", err
	}
	if taken {
		return "", types.ErrConflict
	}
	if err := y.jobs.Acquire(); err != nil {
		return "", err
	}
	now := y.cfg.Clock.Now()
	id := jobID(spec, now)
	job := jobFromSpecTime(id, spec, clock.Format(now))
	rec := wal.Record{Op: wal.OpSubmit, Job: string(id), Payload: encodeSpec(spec)}
	if err := y.journal.Append(rec); err != nil {
		y.jobs.Release()
		return "", fmt.Errorf("%w: %v", types.ErrWAL, err)
	}
	if err := y.db.InsertJob(ctx, job, spec); err != nil {
		y.jobs.Release()
		return "", err
	}
	if err := y.db.SetWALApplied(ctx, y.journal.Seq()); err != nil {
		return "", err
	}
	y.graphs[string(id)] = g
	y.metrics.AddSubmitted(1)
	y.bus.Publish(event.Event{Kind: event.KindSubmit, Job: string(id), At: now})
	return id, nil
}

func jobID(spec types.Spec, now time.Time) digest.Digest {
	return digest.Pair(spec.Name, []byte(string(spec.Fingerprint())+"|"+clock.Format(now)))
}

func jobFromSpecTime(id digest.Digest, spec types.Spec, ts string) types.Job {
	nodes := make([]types.Node, 0, len(spec.Nodes))
	for _, n := range spec.Nodes {
		nodes = append(nodes, types.Node{
			ID: n.ID, Kind: n.Kind, Param: n.Param, Weight: n.Weight, Status: types.NodePending,
		})
	}
	job := types.Job{
		ID: id, Name: spec.Name, Status: types.JobPending,
		CreatedAt: ts, UpdatedAt: ts, Nodes: nodes, Edges: spec.Edges, Quota: spec.Quota,
	}
	job.RefreshStatus()
	return job
}
