// Package ledger 记录作业状态跳转，便于对照 WAL。
package ledger

import (
	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/types"
)

type Entry struct {
	At     string
	Job    string
	Node   string
	From   string
	To     string
	Reason string
}

func NewEntry(clk clock.Clock, job, node, from, to, reason string) Entry {
	at := ""
	if clk != nil {
		at = clock.Format(clk.Now())
	}
	return Entry{At: at, Job: job, Node: node, From: from, To: to, Reason: reason}
}

func JobTransition(clk clock.Clock, job string, from, to types.JobStatus) Entry {
	return NewEntry(clk, job, "", string(from), string(to), "job")
}

func NodeTransition(clk clock.Clock, job, node string, from, to types.NodeStatus, reason string) Entry {
	return NewEntry(clk, job, node, string(from), string(to), reason)
}
