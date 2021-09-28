package cache

func intMax(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func intMin(i, j int) int {
	if i > j {
		return j
	}
	return i
}

func Shorten(s string, k int) string {
	return s[:intMin(len(s), k)]
}
