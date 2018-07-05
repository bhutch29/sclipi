package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"strings"
	"os/exec"
	"os"
)

type scpiManager struct {
	provider ScpiProvider
	inst     instrument
	history History
}

func newScpiManager(i instrument) scpiManager {
	sm := scpiManager{}
	sm.inst = i
	sm.prepareScpiCompleter()
	return sm
}

func (sm *scpiManager) executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
	}

	sm.history.addEntry(s)

	if string(s[0]) == ":" {
		sm.handleScpi(s, sm.inst)
	} else if string(s[0]) == "-"{
		sm.handleOptions(s)
	} else {
		sm.handlePassThrough(s)
	}
}
func (sm *scpiManager) handlePassThrough(s string) {
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}
}

func (sm *scpiManager) handleOptions(s string) {
	switch s {
	case "-history":
		sm.printHistory()
	case "-copy":
		sm.copyPreviousToClipboard()
	}
}
func (sm *scpiManager) copyPreviousToClipboard() {
	//TODO
}
func (sm *scpiManager) printHistory() {
	for _, entry := range sm.history.entries{
		if entry != "-history" {
			fmt.Println(entry)
		}
	}
}

func (sm *scpiManager) handleScpi(s string, inst instrument) {
	if strings.Contains(s, "?") {
		r, err := inst.Query(s)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
	} else {
		inst.Command(s)
	}
}

func (sm *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	if string(d.Text[0]) == ":" {
		tree := sm.provider.getTree(sm.inst)
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		current := sm.getCurrentNode(tree, inputs)

		return prompt.FilterHasPrefix(sm.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	}

	suggests := []prompt.Suggest{
		{Text: "-history", Description: "Not Supported: Show all commands sent this session"},
		{Text: "-copy", Description: "Not Supported: Copy most recent output to clipboard"},
		{Text: "quit", Description: "Exit SCliPI"},
	}

	return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
}

func (sm *scpiManager) prepareScpiCompleter() {
	sm.provider.getTree(sm.inst)
}

func (sm *scpiManager) getCurrentNode(tree scpiNode, inputs []string) scpiNode {
	current := tree
	for _, item := range inputs {
		if success, node := sm.getNodeChildByContent(current, item); success {
			current = node
		}
	}
	return current
}

func (sm *scpiManager) suggestsFromNode(node scpiNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children {
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func (sm *scpiManager) getNodeChildByContent(parent scpiNode, item string) (bool, scpiNode) {
	for _, node := range parent.Children {
		if node.Content == item {
			return true, node
		}
	}
	return false, scpiNode{}
}
