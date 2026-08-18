package wal

import "fmt"

func Verify(recs []Record) error {
	var last uint64
	for i, r := range recs {
		if r.Seq == 0 {
			return fmt.Errorf("record %d missing seq", i)
		}
		if last != 0 && r.Seq < last {
			return fmt.Errorf("record %d seq went backwards %d < %d", i, r.Seq, last)
		}
		last = r.Seq
		if r.Op == 0 {
			return fmt.Errorf("record %d missing op", i)
		}
	}
	return nil
}

func Consecutive(recs []Record) bool {
	if len(recs) == 0 {
		return true
	}
	first := recs[0].Seq
	for i, r := range recs {
		if r.Seq != first+uint64(i) {
			return false
		}
	}
	return true
}

func Jobs(recs []Record) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range recs {
		if r.Job == "" || seen[r.Job] {
			continue
		}
		seen[r.Job] = true
		out = append(out, r.Job)
	}
	return out
}
