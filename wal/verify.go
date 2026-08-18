package wal

func Verify(recs []Record) error {
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
