package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"strings"
	"os/exec"
	"os"
	"github.com/atotto/clipboard"
	"io/ioutil"
	"bufio"
	"log"
)

type scpiManager struct {
	inst     instrument
	commandHistory History
	responseHistory History
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

	sm.commandHistory.addEntry(s)

	if string(s[0]) == ":" || string(s[0]) == "*"{
		sm.handleScpi(s)
	} else if string(s[0]) == "-"{
		sm.handleDashCommands(s)
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

func (sm *scpiManager) handleDashCommands(s string) {
	if s == "-history"{
		sm.printCommandHistory()
	} else if s == "-copy" {
		sm.copyPreviousToClipboard()
	} else if strings.HasPrefix(s, "-saveCommands") {
		sm.saveCommandsToFile(strings.TrimPrefix(s, "-saveCommands"))
	} else {
		fmt.Println(s + ": command not found")
	}
}

func (sm *scpiManager) saveCommandsToFile(fileName string) {
	file := strings.TrimSpace(fileName)
	if file == "" {
		file = "ScpiCommands.txt"
	}
	commands := sm.commandHistory.String()
	if err := ioutil.WriteFile(file, []byte(commands), 0644); err != nil {
		fmt.Println(err)
	}
}

func (sm *scpiManager) copyPreviousToClipboard() {
	if err := clipboard.WriteAll(sm.responseHistory.latest()); err != nil {
		fmt.Println("Copy to clipboard failed: " + err.Error())
	}
}

func (sm *scpiManager) printCommandHistory() {
	fmt.Print(sm.commandHistory.String())
	sm.responseHistory.addEntry(sm.commandHistory.String())
}

func (sm *scpiManager) handleScpi(s string) {
	if strings.Contains(s, "?") {
		r, err := sm.inst.Query(s)
		if err != nil {
			fmt.Println(err)
			sm.responseHistory.addEntry(err.Error())
		}
		fmt.Print(r)
		sm.responseHistory.addEntry(r)
	} else {
		err := sm.inst.Command(s); if err != nil {
			fmt.Println(err)
		}
	}
}

func (sm *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	if string(d.Text[0]) == ":" || string(d.Text[0]) == "*" {
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		inputs = inputs[1:] // Discard first input, is empty string
		current := sm.getCurrentNode(sm.tree, inputs)

		return prompt.FilterHasPrefix(sm.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	}

	if string(d.Text[0]) == "-" || string(d.Text[0]) == "q" {
		suggests := []prompt.Suggest{
			{Text: "-history", Description: "Show all commands sent this session"},
			{Text: "-copy", Description: "Copy most recent output to clipboard"},
			{Text: "-saveCommands", Description: "Save command history to provided filename. If none is provided, will save to ScpiCommands.txt"},
			{Text: "quit", Description: "Exit SCliPI"},
		}

		return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
	}

	return []prompt.Suggest{}
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
	//Only entered a ':'
	if len(inputs) == 1 {
		return tree
	}

	current := tree
	for i, item := range inputs {
		if success, node := sm.getNodeChildByContent(current, item); success {
			current = node
			continue
		} else if i < len(inputs) - 1 {
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
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println("> " + scanner.Text())
		sm.handleScpi(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
