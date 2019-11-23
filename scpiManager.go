package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"strings"
	"os/exec"
	"os"
	"github.com/atotto/clipboard"
)

type scpiManager struct {
	inst     instrument
	history History
	tree scpiNode
}

func newScpiManager(i instrument) scpiManager {
	sm := scpiManager{}
	sm.inst = i
	sm.getTree(i)
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

	if string(s[0]) == ":" || string(s[0]) == "*"{
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
	default:
		fmt.Println(s + ": No command found")
	}
}
func (sm *scpiManager) copyPreviousToClipboard() {
	if err := clipboard.WriteAll(sm.history.latest()); err != nil {
		fmt.Println("Copy to clipboard failed: " + err.Error())
	}
}

func (sm *scpiManager) printHistory() {
	for i, entry := range sm.history.entries{
		if i != len(sm.history.entries) {
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
		err := inst.Command(s); if err != nil {
			fmt.Println(err)
		}
	}
}

func (sm *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	if string(d.Text[0]) == ":" || string(d.Text[0]) == "*" {
		sm.getTree(sm.inst)
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		current := sm.getCurrentNode(sm.tree, inputs)
		//println("\ncurrent: " + current.Content + "\n")

		return prompt.FilterHasPrefix(sm.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	}

	suggests := []prompt.Suggest{
		{Text: "-history", Description: "Show all commands sent this session"},
		{Text: "-copy", Description: "Copy most recent output to clipboard"},
		{Text: "quit", Description: "Exit SCliPI"},
	}

	return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
}

func (sm *scpiManager) getTree(i instrument) {
	if len(sm.tree.Children) == 0 {
		lines, err := i.getSupportedCommands(); if err != nil {
			fmt.Println(err)
		} else {
			sm.tree = parseScpi(lines)
		}
	}
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
