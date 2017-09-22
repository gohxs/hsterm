package hsterm

import "os"

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

type WriteFlush struct {
	*os.File
}

func (wf *WriteFlush) Write(b []byte) (n int, err error) {

	n, err = wf.File.Write(b)
	wf.File.Sync()

	return
}
