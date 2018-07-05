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
}

func newScpiManager(i instrument) scpiManager {
	sm := scpiManager{}
	sm.inst = i
	sm.prepareScpiCompleter()
	return sm
}

func (sc *scpiManager) executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
	}

	if string(s[0]) == ":" {
		handleScpi(s, sc.inst)
	} else if string(s[0]) == "-"{
		handleOptions(s)
	} else {
		handlePassThrough(s)
	}
}
func handlePassThrough(s string) {
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}
}

func handleOptions(s string) {
	switch s {
	case "-history":
		printHistory()
	}
}
func printHistory() {
	//TODO: Store my own history outside of the prompt library
}

func handleScpi(s string, inst instrument) {
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

func (sc *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	if string(d.Text[0]) == ":" {
		tree := sc.provider.getTree(sc.inst)
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		current := sc.getCurrentNode(tree, inputs)

		return prompt.FilterHasPrefix(sc.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	}

	suggests := []prompt.Suggest{
		{Text: "-history", Description: "Not Supported: Show all commands sent this session"},
		{Text: "-copy", Description: "Not Supported: Copy most recent output to clipboard"},
		{Text: "quit", Description: "Exit SCliPI"},
	}

	return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
}

func (sc *scpiManager) prepareScpiCompleter() {
	sc.provider.getTree(sc.inst)
}

func (sc *scpiManager) getCurrentNode(tree scpiNode, inputs []string) scpiNode {
	current := tree
	for _, item := range inputs {
		if success, node := sc.getNodeChildByContent(current, item); success {
			current = node
		}
	}
	return current
}

func (sc *scpiManager) suggestsFromNode(node scpiNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children {
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func (sc *scpiManager) getNodeChildByContent(parent scpiNode, item string) (bool, scpiNode) {
	for _, node := range parent.Children {
		if node.Content == item {
			return true, node
		}
	}
	return false, scpiNode{}
}
