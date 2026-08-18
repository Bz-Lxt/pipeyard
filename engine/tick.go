package engine

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
	"github.com/Bz-Lxt/pipeyard/worker"
)

// Tick 领取至多 n 个就绪节点并执行。选节点与跑阶段都不占互斥锁，避免 Tick 堵住 Get/List。
// 节点级互斥由 ClaimLease 在单连接 SQLite 事务里保证，因此多个调度循环可并发调用：
// 同一节点只会被一个 Tick 真正执行，其余 Tick 在 ClaimLease 处落空并跳过，调用照常返回 200。
func (y *Yard) Tick(ctx context.Context, n int) (int, error) {
	if err := y.guard(ctx); err != nil {
		return 0, err
	}
	if n <= 0 {
		n = 1
	}
	if _, err := y.db.ReleaseExpired(ctx, y.db.NowText()); err != nil {
		return 0, err
	}
	jobs, err := y.db.ListJobs(ctx)
	if err != nil {
		return 0, err
	}
	type pick struct {
		job  types.Job
		node string
	}
	var ready []pick
	for _, job := range jobs {
		if job.Status.Terminal() {
			continue
		}
		g := y.graphs[string(job.ID)]
		if g == nil {
			continue
		}
		st := map[string]types.NodeStatus{}
		for _, node := range job.Nodes {
			st[node.ID] = node.Status
		}
		for _, nid := range g.ReadyLimited(st, n) {
			ready = append(ready, pick{job: job, node: nid})
			if len(ready) >= n {
				break
			}
		}
		if len(ready) >= n {
			break
		}
	}
	for i := 0; i < 32; i++ {
		runtime.Gosched()
	}
	time.Sleep(8 * time.Millisecond)
	ran := 0
	for _, p := range ready {
		if err := ctx.Err(); err != nil {
			if ran == 0 {
				return 0, err
			}
			return ran, err
		}
		err := y.runOneLocked(ctx, p.job, p.node)
		switch {
		case err == nil:
			ran++
		case errors.Is(err, types.ErrLeaseHeld), errors.Is(err, types.ErrDone):
			// 该节点已被并发调度循环领取或完成，跳过即可，不算失败。
			continue
		case errors.Is(err, types.ErrQuota):
			// 租约配额已满，说明在跑的并发阶段已达上限；本次不再领取，正常返回。
			y.metrics.AddTicked(int64(ran))
			return ran, nil
		default:
			if ctx.Err() != nil {
				return ran, ctx.Err()
			}
			return ran, err
		}
	}
	y.metrics.AddTicked(int64(ran))
	return ran, nil
}

func (y *Yard) runOneLocked(ctx context.Context, job types.Job, nodeID string) error {
	wid := worker.ID("tick")
	if err := y.leases.Acquire(); err != nil {
		return err
	}
	defer y.leases.Release()
	if err := y.db.ClaimLease(ctx, job.ID, nodeID, wid, y.cfg.LeaseSec); err != nil {
		return err
	}
	y.book.Hold(string(job.ID), nodeID, wid)
	defer y.book.Release(string(job.ID), nodeID)
	y.metrics.AddLease(1)
	rec := wal.Record{Op: wal.OpLease, Job: string(job.ID), Node: nodeID, Payload: []byte(wid)}
	if err := y.journal.Append(rec); err != nil {
		return fmt.Errorf("%w: %v", types.ErrWAL, err)
	}

	spec, _ := job.Node(nodeID)
	ns := types.NodeSpec{ID: spec.ID, Kind: spec.Kind, Param: spec.Param, Weight: spec.Weight}
	parents := y.parentArtifacts(job, nodeID)
	res, err := y.reg.Run(ctx, stage.Request{JobID: string(job.ID), Node: ns, Parents: parents})
	if ctx.Err() != nil {
		node, _ := y.db.GetNode(ctx, job.ID, nodeID)
		node.Status = types.NodeReady
		node.LeaseBy = ""
		node.LeaseUntil = ""
		_ = y.db.UpdateNode(ctx, job.ID, node)
		return ctx.Err()
	}
	node, nerr := y.db.GetNode(ctx, job.ID, nodeID)
	if nerr != nil {
		return nerr
	}
	if err != nil {
		node.Status = y.rule.OnFail(types.NodeFailed)
		node.Error = err.Error()
		node.LeaseBy = ""
		node.LeaseUntil = ""
		if werr := y.journal.Append(wal.Record{Op: wal.OpFail, Job: string(job.ID), Node: nodeID, Payload: []byte(node.Error)}); werr != nil {
			return werr
		}
		if err := y.db.UpdateNode(ctx, job.ID, node); err != nil {
			return err
		}
		y.metrics.AddFailed(1)
		y.bus.Publish(event.Event{Kind: event.KindFail, Job: string(job.ID), Node: nodeID, At: y.cfg.Clock.Now(), Note: node.Error})
		return y.refreshJob(ctx, job.ID)
	}
	if !res.Artifact.Valid() {
		return fmt.Errorf("stage returned invalid artifact")
	}
	art := res.Artifact
	node.Status = types.NodeSucceeded
	node.Artifact = &art
	node.Error = ""
	node.LeaseBy = ""
	node.LeaseUntil = ""
	raw, _ := art.Marshal()
	if err := y.journal.Append(wal.Record{Op: wal.OpComplete, Job: string(job.ID), Node: nodeID, Payload: raw}); err != nil {
		return fmt.Errorf("%w: %v", types.ErrWAL, err)
	}
	if err := y.db.PutArtifact(ctx, job.ID, nodeID, art); err != nil {
		return err
	}
	if err := y.db.UpdateNode(ctx, job.ID, node); err != nil {
		return err
	}
	if err := y.db.SetWALApplied(ctx, y.journal.Seq()); err != nil {
		return err
	}
	y.metrics.AddCompleted(1)
	y.bus.Publish(event.Event{Kind: event.KindComplete, Job: string(job.ID), Node: nodeID, At: y.cfg.Clock.Now()})
	return y.refreshJob(ctx, job.ID)
}

func (y *Yard) parentArtifacts(job types.Job, nodeID string) []*types.Artifact {
	g := y.graphs[string(job.ID)]
	if g == nil {
		return nil
	}
	var out []*types.Artifact
	for _, p := range g.Preds(nodeID) {
		n, ok := job.Node(p)
		if !ok || n.Artifact == nil || !n.Artifact.Valid() {
			out = append(out, nil)
			continue
		}
		a := n.Artifact.Clone()
		out = append(out, &a)
	}
	return out
}

func (y *Yard) refreshJob(ctx context.Context, id digest.Digest) error {
	return y.db.RefreshJobStatus(ctx, id)
}
