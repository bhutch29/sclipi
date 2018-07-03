package main

import (
	"github.com/c-bata/go-prompt"
	"strings"
)


type ipCompleter struct{
	provider IpProvider
}

func (ic *ipCompleter) completer(d prompt.Document) []prompt.Suggest {
	tree := parseIpv4(ic.provider.getIpAddresses(ic.provider.filterIpv4))
	inputs := strings.Split(d.TextBeforeCursor(), ".")

	ipSuggests := ic.suggestsFromNode(ic.getCurrentNode(tree, inputs))
	ipSuggests = prompt.FilterHasPrefix(ipSuggests, d.GetWordBeforeCursorUntilSeparator("."), true)

	otherSuggests := []prompt.Suggest{{Text: "localhost", Description: "Connect to local machine"}}
	otherSuggests = prompt.FilterHasPrefix(otherSuggests, d.GetWordBeforeCursor(), false)

	suggests := append(ipSuggests, otherSuggests...)

	return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursorUntilSeparator("."), true)
}

func (ic *ipCompleter) getCurrentNode(node ipNode, inputs []string) ipNode {
	current := node
	for _, item := range inputs {
		if success, node := ic.getNodeChildByContent(current, item); success {
			current = node
		}
	}
	return current
}

func (ic *ipCompleter) getNodeChildByContent(parent ipNode, item string) (bool, ipNode) {
	for _, node := range parent.Children {
		if node.Content == item {
			return true, node
		}
	}
	return false, ipNode{}
}

func (ic *ipCompleter) suggestsFromNode(node ipNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children {
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}
