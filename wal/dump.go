package wal

import (
	"bytes"
	"fmt"
)

func Dump(recs []Record) string {
	var b bytes.Buffer
	for i, r := range recs {
		fmt.Fprintf(&b, "%d seq=%d op=%s job=%s node=%s payload=%d crc=%d\n",
			i, r.Seq, OpName(r.Op), r.Job, r.Node, len(r.Payload), r.CRC)
	}
	return b.String()
}

func FilterJob(recs []Record, job string) []Record {
	var out []Record
	for _, r := range recs {
		if r.Job == job {
			out = append(out, r)
		}
	}
	return out
}

func FilterOp(recs []Record, op uint8) []Record {
	var out []Record
	for _, r := range recs {
		if r.Op == op {
			out = append(out, r)
		}
	}
	return out
}

func CountOps(recs []Record) map[string]int {
	m := map[string]int{}
	for _, r := range recs {
		m[OpName(r.Op)]++
	}
	return m
}
