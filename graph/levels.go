package graph

func (g *Graph) Levels() [][]string {
	d := g.Depths()
	max := 0
	for _, v := range d {
		if v > max {
			max = v
		}
	}
	out := make([][]string, max+1)
	buf := make([]string, 0, len(g.order))
	for lv := 0; lv <= max; lv++ {
		for _, id := range g.order {
			if d[id] == lv {
				buf = append(buf, id)
			}
		}
		out[lv] = buf
		buf = buf[:0]
	}
	return out
}

func (g *Graph) Width() int {
	best := 0
	for _, lv := range g.Levels() {
		if len(lv) > best {
			best = len(lv)
		}
	}
	return best
}

func (g *Graph) Independent() bool {
	return g.Width() == g.Size() && len(g.Roots()) == g.Size()
}

func (g *Graph) CriticalPath() []string {
	d := g.Depths()
	bestID := ""
	best := -1
	for _, id := range g.order {
		if d[id] > best {
			best = d[id]
			bestID = id
		}
	}
	if bestID == "" {
		return nil
	}
	path := []string{bestID}
	cur := bestID
	for {
		pred := ""
		pd := -1
		for _, p := range g.pred[cur] {
			if d[p] > pd {
				pd = d[p]
				pred = p
			}
		}
		if pred == "" {
			break
		}
		path = append([]string{pred}, path...)
		cur = pred
	}
	return path
}
