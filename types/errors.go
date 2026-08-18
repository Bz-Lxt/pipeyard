// Package types 描述作业规格、节点状态与可识别错误。
package types

import "errors"

var (
	ErrClosed      = errors.New("yard closed")
	ErrReadOnly    = errors.New("yard read only")
	ErrEmptyGraph  = errors.New("empty graph")
	ErrCycle       = errors.New("graph has a cycle")
	ErrUnknownNode = errors.New("unknown node")
	ErrUnknownKind = errors.New("unknown stage kind")
	ErrQuota       = errors.New("quota exceeded")
	ErrLeaseHeld   = errors.New("lease still held")
	ErrNotFound    = errors.New("job not found")
	ErrCanceled    = errors.New("job canceled")
	ErrConflict    = errors.New("job name conflict")
	ErrBadSpec     = errors.New("invalid spec")
	ErrWAL         = errors.New("wal failure")
	ErrCorrupt     = errors.New("corrupt record")
	ErrNotReady    = errors.New("node not ready")
	ErrDone        = errors.New("job already terminal")
)

// Kind 是阶段类型。未知值不得提交。
type Kind string

const (
	KindValidate Kind = "validate"
	KindMap      Kind = "map"
	KindFilter   Kind = "filter"
	KindReduce   Kind = "reduce"
	KindFanout   Kind = "fanout"
	KindJoin     Kind = "join"
	KindPersist  Kind = "persist"
	KindNotify   Kind = "notify"
)

func KnownKind(k Kind) bool {
	switch k {
	case KindValidate, KindMap, KindFilter, KindReduce, KindFanout, KindJoin, KindPersist, KindNotify:
		return true
	default:
		return false
	}
}

func AllKinds() []Kind {
	return []Kind{KindValidate, KindMap, KindFilter, KindReduce, KindFanout, KindJoin, KindPersist, KindNotify}
}
