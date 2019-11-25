package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/c-bata/go-prompt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type scpiManager struct {
	inst    instrument
	history History
	tree    scpiNode
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

	switch string(s[0]) {
	case ":", "*":
		sm.handleScpi(s)
	case "-":
		sm.handleDashCommands(s)
	case "$":
		sm.handlePassThrough(s)
	case "?":
		printHelp()
	default:
		fmt.Println("Command not recognized. All commands must start with :, *, -, or $")
	}
}

func (sm *scpiManager) handlePassThrough(s string) {
	s = strings.TrimLeft(s, "$ ")
	cmd := exec.Command("sh", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}
}

func (sm *scpiManager) handleDashCommands(s string) {
	if s == "-history" {
		sm.printCommandHistory()
	} else if s == "-copy" {
		sm.copyPreviousToClipboard()
	} else if s == "-copy_all" {
		sm.copyAllToClipboard()
	} else if strings.HasPrefix(s, "-save_script") {
		sm.saveCommandsToFile(strings.TrimPrefix(s, "-save_script"))
	} else if strings.HasPrefix(s, "-run_script") {
		sm.runScript(strings.TrimPrefix(s, "-run_script"))
	} else {
		fmt.Println(s + ": command not found")
	}
}

func (sm *scpiManager) saveCommandsToFile(fileName string) {
	file := strings.TrimSpace(fileName)
	if file == "" {
		file = "ScpiCommands.txt"
	}
	commands := sm.history.CommandsString()
	if err := ioutil.WriteFile(file, []byte(commands), 0644); err != nil {
		fmt.Println(err)
	}
}

func (sm *scpiManager) copyPreviousToClipboard() {
	if err := clipboard.WriteAll(sm.history.latestResponse()); err != nil {
		fmt.Println("Copy to clipboard failed: " + err.Error())
	}
}

func (sm *scpiManager) copyAllToClipboard() {
	if err := clipboard.WriteAll(sm.history.String()); err != nil {
		fmt.Println("Copy to clipboard failed: " + err.Error())
	}
}

func (sm *scpiManager) printCommandHistory() {
	fmt.Print(sm.history.CommandsString())
}

func (sm *scpiManager) handleScpi(s string) {
	if strings.Contains(s, "?") {
		r, err := sm.inst.Query(s)
		if err != nil {
			fmt.Println(err)
			sm.history.addResponse(err.Error())
		}
		fmt.Print(r)
		sm.history.addCommand(s)
		sm.history.addResponse(r)
	} else {
		err := sm.inst.Command(s)
		if err != nil {
			fmt.Println(err)
		}
		sm.history.addCommand(s)
	}
}

func (sm *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		suggests := []prompt.Suggest{
			{Text: ":", Description: "Standard Commands"},
			{Text: "*", Description: "Common Commands"},
			{Text: "-", Description: "Actions (history, clipboard, etc.)"},
			{Text: "$", Description: "Run shell command"},
			{Text: "?", Description: "Help"},
		}
		return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), false)
	}

	firstChar := string(d.Text[0])

	if firstChar == ":" || firstChar == "*" {
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		current := sm.getCurrentNode(sm.tree, inputs[1:]) // Discard first input, is empty string
		return prompt.FilterHasPrefix(sm.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	}

	if firstChar == "-" || firstChar == "q" {
		suggests := []prompt.Suggest{
			{Text: "-history", Description: "Show all commands sent this session"},
			{Text: "-save_script", Description: "Save command history to provided filename. Default: ScpiCommands.txt"},
			{Text: "-run_script", Description: "Run script from provided filename. Default: ScpiCommands.txt"},
			{Text: "-copy", Description: "Copy most recent SCPI response to clipboard"},
			{Text: "-copy_all", Description: "Copy entire session to clipboard"},
			{Text: "quit", Description: "Exit Sclipi"},
		}

		return prompt.FilterHasPrefix(suggests, d.CurrentLine(), true)
	}

	return []prompt.Suggest{}
}

func (sm *scpiManager) getTree(i instrument) {
	if len(sm.tree.Children) == 0 {
		lines, err := i.getSupportedCommands()
		if err != nil {
			fmt.Println(err)
		} else {
			sm.tree = parseScpi(lines)
		}
	}
}

func (sm *scpiManager) getCurrentNode(tree scpiNode, inputs []string) scpiNode {
	//Only entered a ':'
	if len(inputs) == 1 {
		return tree
	}

	current := tree
	for i, item := range inputs {
		if success, node := sm.getNodeChildByContent(current, item); success {
			current = node
			continue
		} else if i < len(inputs)-1 {
			return scpiNode{}
		} else {
			break
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

func (sm *scpiManager) runScript(file string) {
	file = strings.TrimSpace(file)
	if file == "" {
		file = "ScpiCommands.txt"
	}
	lines, err := readLines(file)
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range lines {
		fmt.Println("> " + line)
		sm.handleScpi(line)
	}
}
