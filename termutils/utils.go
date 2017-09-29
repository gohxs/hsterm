package termutils

// Common utils
//
// gets first Value else leave it
func lookIndex(a []int, index int, def int) int {
	if index < len(a) && index >= 0 {
		return a[index]
	}
	return def
}
func min(n, min int) int {
	if n < min {
		return min
	}
	return n
}

func max(n, max int) int {
	if n > max {
		return max
	}
	return n
}
