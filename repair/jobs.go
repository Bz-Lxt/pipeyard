package repair

import (
	"context"

	"github.com/Bz-Lxt/pipeyard/store"
	"github.com/Bz-Lxt/pipeyard/types"
)

func RecomputeStatuses(ctx context.Context, db *store.DB) (int, error) {
	jobs, err := db.ListJobs(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, j := range jobs {
		old := j.Status
		j.RefreshStatus()
		if j.Status != old {
			if err := db.UpdateJobStatus(ctx, j.ID, j.Status, j.Error); err != nil {
				return n, err
			}
			n++
		}
	}
	return n, nil
}

func CountOpen(jobs []types.Job) int {
	n := 0
	for _, j := range jobs {
		if !j.Status.Terminal() || j.Status == types.JobCanceled {
			n++
		}
	}
	return n
}
