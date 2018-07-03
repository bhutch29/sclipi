package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"strings"
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
	if strings.Contains(s, "?") {
		r, err := sc.inst.Query(s)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r)
	} else {
		sc.inst.Command(s)
	}
}

func (sc *scpiManager) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	//if string(d.Text[0]) == ":" {
		tree := sc.provider.getTree(sc.inst)
		inputs := strings.Split(d.TextBeforeCursor(), ":")
		current := sc.getCurrentNode(tree, inputs)

		return prompt.FilterHasPrefix(sc.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
	//}

	//return []prompt.Suggest{
	//	{Text: "history"},
	//}
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
