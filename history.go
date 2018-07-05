package main

type History struct{
	entries []string
}

func (h *History) addEntry(s string) {
	h.entries = append(h.entries, s)
}
