package main

import (
	"github.com/c-bata/go-prompt"
	"strings"
	"SCliPI/ipParser"
	"SCliPI/scpiParser"
)

func nullCompleter(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

var scpiProvider ScpiProvider

func prepareScpiCompleter(){
	scpiProvider.GetTree()
}

func scpiCompleter(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	tree := scpiProvider.GetTree()
	inputs := strings.Split(d.TextBeforeCursor(), ":")

	current := tree
	for _, item := range inputs {
		if success, node := getScpiNodeChildByContent(current, item); success{
			current = node
		}
	}

	return prompt.FilterHasPrefix(buildSuggestsFromScpiNode(current), d.GetWordBeforeCursorUntilSeparator(":"), true)
}

func buildSuggestsFromScpiNode(node scpiParser.ScpiNode) []prompt.Suggest {
	var s []prompt.Suggest
	for _, item := range node.Children{
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func getScpiNodeChildByContent(parent scpiParser.ScpiNode, item string) (bool, scpiParser.ScpiNode) {
	for _, node := range parent.Children{
		if node.Content == item {
			return true, node
		}
	}
	return false, scpiParser.ScpiNode{}
}

var ipProvider IpProvider

func ipCompleter(d prompt.Document) []prompt.Suggest {
	tree := ipParser.ParseIpv4(ipProvider.getIpAddresses(ipProvider.filterIpv4))
	inputs := strings.Split(d.TextBeforeCursor(), ".")

	current := tree
	for _, item := range inputs {
		if success, node := getIpNodeChildByContent(current, item); success{
			current = node
		}
	}

	return prompt.FilterHasPrefix(buildSuggestsFromIpNode(current), d.GetWordBeforeCursorUntilSeparator("."), true)
}

func getIpNodeChildByContent(parent ipParser.IpNode, item string) (bool, ipParser.IpNode) {
	for _, node := range parent.Children{
		if node.Content == item {
			return true, node
		}
	}
	return false, ipParser.IpNode{}
}

func buildSuggestsFromIpNode(node ipParser.IpNode) []prompt.Suggest{
	var s []prompt.Suggest
	for _, item := range node.Children{
		s = append(s, prompt.Suggest{Text: item.Content})
	}
	return s
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}