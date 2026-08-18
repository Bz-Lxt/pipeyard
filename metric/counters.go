package metric

import "fmt"

func (s Snapshot) Map() map[string]int64 {
	return map[string]int64{
		"submitted":   s.Submitted,
		"ticked":      s.Ticked,
		"completed":   s.Completed,
		"failed":      s.Failed,
		"canceled":    s.Canceled,
		"checkpoints": s.Checkpoints,
		"leases":      s.Leases,
	}
}

func (s Snapshot) String() string {
	return fmt.Sprintf("sub=%d tick=%d ok=%d fail=%d cancel=%d ckpt=%d lease=%d",
		s.Submitted, s.Ticked, s.Completed, s.Failed, s.Canceled, s.Checkpoints, s.Leases)
}

func (s Snapshot) TotalTerminal() int64 {
	return s.Completed + s.Failed + s.Canceled
}
