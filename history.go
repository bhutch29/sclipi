package main

import "strings"

type Class int

const (
	Command Class = iota + 1
	Response
)

type Entry struct {
	Text string
	Class Class
}
type History struct{
	entries []Entry
}

func (h *History) addCommand(s string) {
	if !strings.HasPrefix(s, "-") {
		h.entries = append(h.entries, Entry{Text: s, Class: Command})
	}
}

func (h *History) addResponse(s string) {
	h.entries = append(h.entries, Entry{Text: s, Class: Response})
}

func (h *History) latestResponse() string {
	for i := len(h.entries) - 1; i >= 0; i-- {
		if h.entries[i].Class == Response {
			return h.entries[i].Text
		}
	}
	return ""
}

func (h *History) CommandsString() string {
	var result string
	for _, entry := range h.entries{
		if entry.Class == Command {
			result += entry.Text + "\n"
		}
	}
	return result 
}

func (h *History) String() string {
	var result string
	for _, entry := range h.entries{
		if entry.Class == Command {
			result += "> " + entry.Text + "\n"
		} else {
			result += entry.Text
		}
	}
	return result
}
