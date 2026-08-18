package policy

import "github.com/Bz-Lxt/pipeyard/types"

func (r Rule) WithPause(kinds ...types.Kind) Rule {
	out := r.Clone()
	out.PauseKinds = append(out.PauseKinds, kinds...)
	return out
}

func (r Rule) WithoutPause(k types.Kind) Rule {
	out := r.Clone()
	var keep []types.Kind
	for _, p := range out.PauseKinds {
		if p != k {
			keep = append(keep, p)
		}
	}
	out.PauseKinds = keep
	return out
}

func HumanGate(k types.Kind) bool {
	return k == types.KindNotify || k == types.KindPersist
}

func AutoKinds() []types.Kind {
	var out []types.Kind
	for _, k := range types.AllKinds() {
		if !HumanGate(k) {
			out = append(out, k)
		}
	}
	return out
}

func MaxAttemptOr(r Rule, fallback int) int {
	if r.MaxAttempt > 0 {
		return r.MaxAttempt
	}
	if fallback > 0 {
		return fallback
	}
	return 1
}
