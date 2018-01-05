package termu

import (
	"strings"
)

type History interface {
	Append(string)
	Prev(string) string
	Next(string) string
	List() []string
	String() string // current item
}

// History lined history
type history struct {
	index   int
	history []string
	buf     string
}

func (h *history) Append(item string) {

	item = strings.TrimSpace(item)
	if len(item) == 0 {
		return
	}
	h.index = len(h.history) // Set to last
	if len(h.history) > 0 {
		last := h.history[len(h.history)-1]
		if item == last { // Ignore if it is the same as last one
			return
		}
	}
	h.history = append(h.history, item)
	h.index = len(h.history)
}

func (h *history) Prev(line string) string {
	if h.index == len(h.history) {
		h.buf = line
	}
	h.index = min(h.index-1, 0)
	return h.String()
}
func (h *history) Next(line string) string {
	h.index++
	if h.index >= len(h.history) {
		h.index = len(h.history)
		return h.buf
	}
	return h.String()
}
func (h *history) Last(last string) {
	h.buf = last
}
func (h *history) List() []string {
	return h.history
}

func (h *history) String() string {
	if len(h.history) == 0 || h.index == len(h.history) {
		return ""
	}
	return h.history[h.index]
}
