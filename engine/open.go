package engine

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/graph"
	"github.com/Bz-Lxt/pipeyard/policy"
	"github.com/Bz-Lxt/pipeyard/quota"
	"github.com/Bz-Lxt/pipeyard/repair"
	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/store"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
	"github.com/Bz-Lxt/pipeyard/worker"
)

func Open(cfg config.Config) (*Yard, error) {
	cfg, err := config.Normalize(cfg)
	if err != nil {
		return nil, err
	}
	db, err := store.Open(cfg.SQLitePath(), cfg.Clock)
	if err != nil {
		return nil, err
	}
	j, err := wal.Open(cfg.JournalPath())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	y := &Yard{
		cfg:     cfg,
		db:      db,
		journal: j,
		reg:     stage.NewRegistry(),
		jobs:    quota.New(cfg.MaxJobs),
		leases:  quota.New(cfg.Workers * 4),
		book:    worker.NewBook(),
		bus:     event.NewBus(),
		rule:    policy.Default(),
		graphs:  map[string]*graph.Graph{},
	}
	if err := y.replayUnapplied(); err != nil {
		_ = y.Close()
		return nil, fmt.Errorf("replay: %w", err)
	}
	n, err := db.JobCount(context.Background())
	if err != nil {
		_ = y.Close()
		return nil, err
	}
	y.jobs.Reset(n)
	if err := y.cacheGraphs(context.Background()); err != nil {
		_ = y.Close()
		return nil, err
	}
	return y, nil
}

func (y *Yard) replayUnapplied() error {
	ctx := context.Background()
	applied, err := y.db.WALApplied(ctx)
	if err != nil {
		return err
	}
	raw, err := y.journal.ReadAll()
	if err != nil {
		return err
	}
	recs, err := wal.Replay(raw)
	if err != nil {
		return err
	}
	var max uint64
	for _, rec := range recs {
		if rec.Seq > max {
			max = rec.Seq
		}
		if rec.Op == wal.OpSubmit {
			if err := y.applyRecord(ctx, rec); err != nil {
				return fmt.Errorf("apply seq=%d: %w", rec.Seq, err)
			}
			if rec.Seq > applied {
				if err := y.db.SetWALApplied(ctx, rec.Seq); err != nil {
					return err
				}
			}
			continue
		}
		if rec.Seq <= applied {
			continue
		}
		if err := y.applyRecord(ctx, rec); err != nil {
			return fmt.Errorf("apply seq=%d: %w", rec.Seq, err)
		}
		if err := y.db.SetWALApplied(ctx, rec.Seq); err != nil {
			return err
		}
	}
	y.journal.SetSeq(max)
	repair.AlignSeq(y.journal, applied)
	return nil
}

func (y *Yard) cacheGraphs(ctx context.Context) error {
	jobs, err := y.db.ListJobs(ctx)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		spec := types.Spec{Name: job.Name, Quota: job.Quota, Edges: job.Edges}
		for _, n := range job.Nodes {
			spec.Nodes = append(spec.Nodes, types.NodeSpec{ID: n.ID, Kind: n.Kind, Param: n.Param, Weight: n.Weight})
		}
		g, err := graph.Build(spec)
		if err != nil {
			return err
		}
		y.graphs[string(job.ID)] = g
	}
	return nil
}

func (y *Yard) applyRecord(ctx context.Context, rec wal.Record) error {
	switch rec.Op {
	case wal.OpSubmit:
		spec, err := decodeSpec(rec.Payload)
		if err != nil {
			return err
		}
		if job, err := y.db.GetJob(ctx, digest.Digest(rec.Job)); err == nil {
			job.Status = types.JobPending
			job.Error = ""
			for i := range job.Nodes {
				job.Nodes[i].Status = types.NodePending
				job.Nodes[i].Artifact = nil
				job.Nodes[i].Error = ""
				if err := y.db.UpdateNode(ctx, job.ID, job.Nodes[i]); err != nil {
					return err
				}
			}
			return y.db.UpdateJobStatus(ctx, job.ID, types.JobPending, "")
		}
		job := jobFromSpecTime(digest.Digest(rec.Job), spec, y.db.NowText())
		return y.db.InsertJob(ctx, job, spec)
	case wal.OpComplete, wal.OpFail, wal.OpCancel, wal.OpLease, wal.OpCheckpoint:
		return nil
	default:
		return fmt.Errorf("unknown wal op %d", rec.Op)
	}
}
