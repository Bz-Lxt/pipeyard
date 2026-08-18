package event

func FilterKind(ev []Event, k Kind) []Event {
	var out []Event
	for _, e := range ev {
		if e.Kind == k {
			out = append(out, e)
		}
	}
	return out
}

func FilterJob(ev []Event, job string) []Event {
	var out []Event
	for _, e := range ev {
		if e.Job == job {
			out = append(out, e)
		}
	}
	return out
}

func Last(ev []Event, k Kind) (Event, bool) {
	for i := len(ev) - 1; i >= 0; i-- {
		if ev[i].Kind == k {
			return ev[i], true
		}
	}
	return Event{}, false
}

func CountByKind(ev []Event) map[Kind]int {
	m := map[Kind]int{}
	for _, e := range ev {
		m[e.Kind]++
	}
	return m
}
