package engine

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/catalog"
	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/export"
	"github.com/Bz-Lxt/pipeyard/types"
)

func (y *Yard) Inspect(ctx context.Context, id digest.Digest) (string, error) {
	job, err := y.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return export.JobText(job), nil
}

func (y *Yard) Describe(ctx context.Context, id digest.Digest) (string, error) {
	job, err := y.Get(ctx, id)
	if err != nil {
		return "", err
	}
	spec := types.Spec{Name: job.Name, Edges: job.Edges}
	for _, n := range job.Nodes {
		spec.Nodes = append(spec.Nodes, types.NodeSpec{ID: n.ID, Kind: n.Kind, Param: n.Param, Weight: n.Weight})
	}
	return catalog.DescribeSpec(spec), nil
}

func (y *Yard) NodeError(ctx context.Context, id digest.Digest, node string) (string, error) {
	job, err := y.Get(ctx, id)
	if err != nil {
		return "", err
	}
	n, ok := job.Node(node)
	if !ok {
		return "", types.ErrUnknownNode
	}
	return n.Error, nil
}

func (y *Yard) MustOpen() error {
	if y == nil || y.closed {
		return types.ErrClosed
	}
	return nil
}

func (y *Yard) JobDOT(ctx context.Context, id digest.Digest) (string, error) {
	job, err := y.Get(ctx, id)
	if err != nil {
		return "", err
	}
	y.mu.Lock()
	g := y.graphs[string(id)]
	y.mu.Unlock()
	if g == nil {
		_ = job
		return "", fmt.Errorf("graph missing")
	}
	return g.DOT(), nil
}
