package digest

func SumAll(blobs [][]byte) []Digest {
	out := make([]Digest, len(blobs))
	for i, b := range blobs {
		out[i] = Sum(b)
	}
	return out
}

func Unique(in []Digest) []Digest {
	seen := map[Digest]bool{}
	var out []Digest
	for _, d := range in {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		out = append(out, d)
	}
	return out
}

func Contains(in []Digest, want Digest) bool {
	for _, d := range in {
		if Equal(d, want) {
			return true
		}
	}
	return false
}

func JoinHex(in []Digest) string {
	raw := make([]byte, 0, len(in)*Size)
	for _, d := range in {
		b, err := d.Bytes()
		if err != nil {
			continue
		}
		raw = append(raw, b...)
	}
	return string(Sum(raw))
}
