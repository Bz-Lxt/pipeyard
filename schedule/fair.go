package schedule

func FairShare(cands []Candidate, perJob int) []Candidate {
	if perJob <= 0 {
		perJob = 1
	}
	used := map[string]int{}
	var out []Candidate
	for _, c := range cands {
		if used[c.Job] >= perJob {
			continue
		}
		used[c.Job]++
		out = append(out, c)
	}
	return out
}

func JobKeys(cands []Candidate) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range cands {
		if seen[c.Job] {
			continue
		}
		seen[c.Job] = true
		out = append(out, c.Job)
	}
	return out
}
