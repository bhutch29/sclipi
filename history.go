package main

import "strings"

type History struct{
	entries []string
}

func (h *History) addEntry(s string) {
	if !strings.HasPrefix(s, "-") {
		h.entries = append(h.entries, s)
	}
}

func (h *History) latest() string {
		return h.entries[len(h.entries) - 1]
}

func (h *History) String() string {
	var result string
	for _, entry := range h.entries{
		result += entry + "\n"
	}
	return result
}
