package main

type History struct{
	entries []string
}

func (h *History) addEntry(s string) {
	h.entries = append(h.entries, s)
}
func (h *History) latest() string {
	return h.entries[len(h.entries) - 2] //most recent command other than -copy
}