package termu

import "strings"

type History interface {
	Append(string)
	Prev(string) string
	Next(string) string
	List() []string
}

// History lined history
type history struct {
	index   int
	history []string
}

func (h *history) Append(item string) {

	item = strings.TrimSpace(item)
	if len(item) == 0 {
		return
	}
	if len(h.history) > 0 {
		last := h.history[len(h.history)-1]
		if item == last { // Ignore if equal
			return
		}
	}
	h.history = append(h.history, item)
	h.index = len(h.history)
}

func (h *history) Prev(string) string {
	h.index = min(h.index-1, 0)
	return h.String()
}
func (h *history) Next(string) string {
	h.index++
	if h.index >= len(h.history) {
		h.index = len(h.history)
		return ""
	}
	return h.String()
}
func (h *history) List() []string {
	return h.history
}

func (h *history) String() string {
	if len(h.history) == 0 {
		return ""
	}
	return h.history[h.index]
}
