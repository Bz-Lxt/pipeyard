package engine

import (
	"context"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (y *Yard) Get(ctx context.Context, id digest.Digest) (types.Job, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.closed {
		return types.Job{}, types.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return types.Job{}, err
	}
	job, err := y.db.GetJob(ctx, id)
	if err != nil {
		return types.Job{}, err
	}
	return job.Clone(), nil
}

// List 返回副本。调用方改切片不得污染账本。
func (y *Yard) List(ctx context.Context) ([]types.Job, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.closed {
		return nil, types.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	jobs, err := y.db.ListJobs(ctx)
	if err != nil {
		return nil, err
	}
	return types.CloneJobs(jobs), nil
}

func (y *Yard) Summaries(ctx context.Context) ([]types.Summary, error) {
	jobs, err := y.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.Summary, len(jobs))
	for i, j := range jobs {
		out[i] = j.Summary()
	}
	return out, nil
}
