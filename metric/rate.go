package metric

func (s Snapshot) SuccessRate() float64 {
	den := s.Completed + s.Failed
	if den == 0 {
		return 0
	}
	return float64(s.Completed) / float64(den)
}

func (s Snapshot) CancelRate() float64 {
	den := s.Submitted
	if den == 0 {
		return 0
	}
	return float64(s.Canceled) / float64(den)
}

func (s Snapshot) Idle() bool {
	return s.Ticked == 0 && s.Submitted == 0
}

func (s Snapshot) Busy() bool {
	return s.Leases > s.Completed+s.Failed+s.Canceled
}

func Diff(a, b Snapshot) Snapshot {
	return Snapshot{
		Submitted:   b.Submitted - a.Submitted,
		Ticked:      b.Ticked - a.Ticked,
		Completed:   b.Completed - a.Completed,
		Failed:      b.Failed - a.Failed,
		Canceled:    b.Canceled - a.Canceled,
		Checkpoints: b.Checkpoints - a.Checkpoints,
		Leases:      b.Leases - a.Leases,
	}
}
