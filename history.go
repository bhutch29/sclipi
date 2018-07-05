package main

type History struct{
	entries []string
}

func (h *History) addEntry(s string) {
	h.entries = append(h.entries, s)
}

//most recent command other than command used to query latest
func (h *History) latest() string {
	return h.entries[len(h.entries) - 2]
}