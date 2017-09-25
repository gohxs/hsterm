package hsterm

// History lined history
type History struct {
	index   int
	history []string
}

func (h *History) Append(item string) {
	if len(h.history) > 0 {
		last := h.history[len(h.history)-1]
		if item == last { // Ignore if equal
			return
		}
	}
	h.history = append(h.history, item)
	h.index = len(h.history)
}

func (h *History) Prev() string {
	h.index = min(h.index-1, 0)
	return h.String()
}
func (h *History) Next() string {
	h.index++
	if h.index >= len(h.history) {
		h.index = len(h.history)
		return ""
	}
	return h.String()

}

func (h *History) String() string {
	if len(h.history) == 0 {
		return ""
	}
	return h.history[h.index]
}
