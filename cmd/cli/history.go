package main

import (
	"github.com/shibukawa/configdir"
	"os"
	"path/filepath"
	"strings"
)

type Class int

const (
	Command Class = iota + 1
	Response
)

type Entry struct {
	Text string
	Class Class
}
type history struct{
	entries []Entry
}

func (h *history) addCommand(s string) {
	if !strings.HasPrefix(s, "-") {
		entry := Entry{Class: Command, Text: s}
		h.entries = append(h.entries, entry)
		h.addCommandToFile(s)
	}
}

func (h *history) addCommandToFile(s string) {
	configDirs := configdir.New("bhutch29", "sclipi")
	_ = configDirs.QueryCacheFolder().CreateParentDir("history.txt")
	file, err := os.OpenFile(filepath.Join(configDirs.QueryCacheFolder().Path, "history.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	_, _ = file.WriteString(s + "\n")
}

func (h *history) addResponse(s string) {
	entry := Entry{Text: s, Class: Response}
	h.entries = append(h.entries, entry)
}

func (h *history) latestResponse() string {
	for i := len(h.entries) - 1; i >= 0; i-- {
		if h.entries[i].Class == Response {
			return h.entries[i].Text
		}
	}
	return ""
}

func (h *history) CommandsString() string {
	var result string
	for _, entry := range h.entries{
		if entry.Class == Command {
			result += entry.Text + "\n"
		}
	}
	return result 
}

func (h *history) String() string {
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
