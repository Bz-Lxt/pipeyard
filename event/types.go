// Package event 提供进程内事件总线，给面板刷新用。不做跨主机推送。
package event

import "time"

type Kind string

const (
	KindSubmit     Kind = "submit"
	KindTick       Kind = "tick"
	KindComplete   Kind = "complete"
	KindFail       Kind = "fail"
	KindCancel     Kind = "cancel"
	KindCheckpoint Kind = "checkpoint"
)

type Event struct {
	Kind Kind
	Job  string
	Node string
	At   time.Time
	Note string
}

func (e Event) String() string {
	if e.Node == "" {
		return string(e.Kind) + " " + e.Job
	}
	return string(e.Kind) + " " + e.Job + "/" + e.Node
}
