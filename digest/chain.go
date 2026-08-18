package digest

func Chain(first Digest, rest ...[]byte) Digest {
	cur := first
	for _, p := range rest {
		cur = Pair(string(cur), p)
	}
	return cur
}

func Many(blobs [][]byte) Digest {
	if len(blobs) == 0 {
		return Sum(nil)
	}
	cur := Sum(blobs[0])
	for i := 1; i < len(blobs); i++ {
		cur = Pair(string(cur), blobs[i])
	}
	return cur
}

func Strings(ss []string) Digest {
	blobs := make([][]byte, len(ss))
	for i, s := range ss {
		blobs[i] = []byte(s)
	}
	return Many(blobs)
}
