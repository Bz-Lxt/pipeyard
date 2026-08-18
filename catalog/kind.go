// Package catalog 描述阶段种类的运维元数据。
package catalog

import "github.com/Bz-Lxt/pipeyard/types"

type Info struct {
	Kind        types.Kind
	Title       string
	NeedsParent bool
	Produces    string
}

func All() []Info {
	return []Info{
		{types.KindValidate, "校验", false, "body"},
		{types.KindMap, "变换", true, "items"},
		{types.KindFilter, "过滤", true, "items"},
		{types.KindReduce, "归约", true, "body"},
		{types.KindFanout, "分片", true, "items"},
		{types.KindJoin, "汇合", true, "body"},
		{types.KindPersist, "落盘", true, "digest"},
		{types.KindNotify, "通知", false, "message"},
	}
}

func Lookup(k types.Kind) (Info, bool) {
	for _, i := range All() {
		if i.Kind == k {
			return i, true
		}
	}
	return Info{}, false
}

func Titles() map[types.Kind]string {
	m := map[types.Kind]string{}
	for _, i := range All() {
		m[i.Kind] = i.Title
	}
	return m
}
