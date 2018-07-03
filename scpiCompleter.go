package main

import (
	"github.com/c-bata/go-prompt"
	"strings"
)

type scpiCompleter struct {
	provider ScpiProvider
}

func (sc *scpiCompleter) prepareScpiCompleter(){
	sc.provider.getTree()
}

func (sc *scpiCompleter) completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	tree := sc.provider.getTree()
	inputs := strings.Split(d.TextBeforeCursor(), ":")
	current := sc.getCurrentNode(tree, inputs)

	return prompt.FilterHasPrefix(sc.suggestsFromNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
}

func (sc *scpiCompleter) getCurrentNode(tree scpiNode, inputs []string) scpiNode {
	current := tree
	for _, item := range inputs {
		if success, node := sc.getNodeChildByContent(current, item); success {
			current = node
		}
	}
	return current
}

func (sc *scpiCompleter) suggestsFromNode(node scpiNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children{
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func (sc *scpiCompleter) getNodeChildByContent(parent scpiNode, item string) (bool, scpiNode) {
	for _, node := range parent.Children{
		if node.Content == item {
			return true, node
		}
	}
	return false, scpiNode{}
}