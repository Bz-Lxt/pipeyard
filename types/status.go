package types

// JobStatus 是作业级状态机。
type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
	JobCanceled  JobStatus = "canceled"
)

func (s JobStatus) Terminal() bool {
	return s == JobSucceeded || s == JobFailed || s == JobCanceled
}

func (s JobStatus) String() string { return string(s) }

// NodeStatus 是节点级状态机。领取后进入 leased，执行完进入 succeeded/failed。
type NodeStatus string

const (
	NodePending   NodeStatus = "pending"
	NodeReady     NodeStatus = "ready"
	NodeLeased    NodeStatus = "leased"
	NodeSucceeded NodeStatus = "succeeded"
	NodeFailed    NodeStatus = "failed"
	NodeCanceled  NodeStatus = "canceled"
	NodeSkipped   NodeStatus = "skipped"
)

func (s NodeStatus) Terminal() bool {
	return s == NodeSucceeded || s == NodeFailed || s == NodeCanceled || s == NodeSkipped
}

func (s NodeStatus) Runnable() bool {
	return s == NodePending || s == NodeReady
}

func (s NodeStatus) String() string { return string(s) }

// Rollup 由节点集合推导作业状态。任一失败则失败；全部成功则成功；有取消则取消。
func Rollup(nodes []NodeStatus) JobStatus {
	if len(nodes) == 0 {
		return JobPending
	}
	var failed, canceled, running, pending, succeeded int
	for _, n := range nodes {
		switch n {
		case NodeFailed:
			failed++
		case NodeCanceled:
			canceled++
		case NodeLeased:
			running++
		case NodeSucceeded, NodeSkipped:
			succeeded++
		default:
			pending++
		}
	}
	if failed > 0 {
		return JobFailed
	}
	if canceled > 0 && pending == 0 && running == 0 && succeeded+canceled == len(nodes) {
		return JobCanceled
	}
	if succeeded == len(nodes) {
		return JobSucceeded
	}
	if running > 0 || succeeded > 0 {
		return JobRunning
	}
	return JobPending
}
