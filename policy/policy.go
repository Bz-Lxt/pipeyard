// Package policy 描述重试、跳过与人工暂停规则。
package policy

import "github.com/Bz-Lxt/pipeyard/types"

// Rule 决定失败节点是否再领取。
type Rule struct {
	MaxAttempt int
	SkipOnFail bool
	PauseKinds []types.Kind
}

func Default() Rule {
	return Rule{MaxAttempt: 3, SkipOnFail: false}
}

func (r Rule) AllowRetry(attempt int) bool {
	if r.MaxAttempt <= 0 {
		return false
	}
	return attempt < r.MaxAttempt
}

func (r Rule) ShouldPause(k types.Kind) bool {
	for _, p := range r.PauseKinds {
		if p == k {
			return true
		}
	}
	return false
}

func (r Rule) OnFail(st types.NodeStatus) types.NodeStatus {
	if r.SkipOnFail {
		return types.NodeSkipped
	}
	return types.NodeFailed
}

func (r Rule) Clone() Rule {
	out := r
	if r.PauseKinds != nil {
		out.PauseKinds = append([]types.Kind(nil), r.PauseKinds...)
	}
	return out
}
